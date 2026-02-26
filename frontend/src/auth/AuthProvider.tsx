import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { refreshSpotifyToken } from '../api/client';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: number | null;
}

interface AuthContextType {
  accessToken: string | null;
  isAuthenticated: boolean;
  login: (accessToken: string, refreshToken: string, expiresIn: number) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}

function loadAuth(): AuthState {
  try {
    const stored = localStorage.getItem('metaloreian_auth');
    if (stored) return JSON.parse(stored);
  } catch {}
  return { accessToken: null, refreshToken: null, expiresAt: null };
}

function saveAuth(state: AuthState) {
  localStorage.setItem('metaloreian_auth', JSON.stringify(state));
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState>(loadAuth);

  const login = useCallback((accessToken: string, refreshToken: string, expiresIn: number) => {
    const state: AuthState = {
      accessToken,
      refreshToken,
      expiresAt: Date.now() + expiresIn * 1000,
    };
    setAuth(state);
    saveAuth(state);
  }, []);

  const logout = useCallback(() => {
    const state: AuthState = { accessToken: null, refreshToken: null, expiresAt: null };
    setAuth(state);
    localStorage.removeItem('metaloreian_auth');
  }, []);

  // Auto-refresh token before expiry
  useEffect(() => {
    if (!auth.refreshToken || !auth.expiresAt) return;

    const timeUntilRefresh = auth.expiresAt - Date.now() - 60_000; // refresh 1min before expiry
    if (timeUntilRefresh <= 0) {
      // Token expired, refresh now
      refreshSpotifyToken(auth.refreshToken)
        .then((data) => {
          login(data.access_token, data.refresh_token || auth.refreshToken!, data.expires_in);
        })
        .catch(() => logout());
      return;
    }

    const timer = setTimeout(() => {
      refreshSpotifyToken(auth.refreshToken!)
        .then((data) => {
          login(data.access_token, data.refresh_token || auth.refreshToken!, data.expires_in);
        })
        .catch(() => logout());
    }, timeUntilRefresh);

    return () => clearTimeout(timer);
  }, [auth.refreshToken, auth.expiresAt, login, logout]);

  return (
    <AuthContext.Provider
      value={{
        accessToken: auth.accessToken,
        isAuthenticated: !!auth.accessToken,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
