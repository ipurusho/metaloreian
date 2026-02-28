import { BrowserRouter, Routes, Route, Navigate, Link, useLocation } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from './auth/AuthProvider';
import { PlayerProvider } from './player/PlayerContext';
import { LoginPage } from './auth/LoginPage';
import { CallbackPage } from './auth/CallbackPage';
import { BandPage } from './features/band/BandPage';
import { AlbumPage } from './features/album/AlbumPage';
import { DashboardPage } from './features/band/DashboardPage';
import { SearchBar } from './features/search/SearchBar';
import { useAutoDetect } from './features/search/useAutoDetect';
import { PlayerBar } from './player/PlayerBar';
import './styles/theme.css';
import './styles/app.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
    },
  },
});

function AppLayout() {
  const { isAuthenticated, logout } = useAuth();
  const { autoDetectEnabled, setAutoDetect } = useAutoDetect();
  const location = useLocation();
  const isDashboard = location.pathname === '/dashboard';

  if (!isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="app-layout">
      {!isDashboard && (
        <header className="app-header">
          <Link to="/dashboard" className="app-logo-link">
            <h1>metalöreian</h1>
          </Link>
          <SearchBar />
          <div className="header-actions">
            <label className="auto-detect-toggle">
              <input
                type="checkbox"
                checked={autoDetectEnabled}
                onChange={(e) => setAutoDetect(e.target.checked)}
              />
              Auto-detect
            </label>
            <button className="logout-btn" onClick={logout}>
              Logout
            </button>
          </div>
        </header>
      )}
      <main className="app-main">
        <Routes>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/band/:maId" element={<BandPage />} />
          <Route path="/band/:maId/album/:albumId" element={<AlbumPage />} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </main>
      <PlayerBar />
    </div>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <PlayerProvider>
            <Routes>
              <Route path="/" element={<LoginPage />} />
              <Route path="/callback" element={<CallbackPage />} />
              <Route path="/*" element={<AppLayout />} />
            </Routes>
          </PlayerProvider>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
