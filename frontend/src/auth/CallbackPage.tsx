import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './AuthProvider';
import { exchangeSpotifyToken } from '../api/client';
import { getCodeVerifier, clearCodeVerifier } from './pkce';

export function CallbackPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const errorParam = params.get('error');

    if (errorParam) {
      setError(`Spotify authorization failed: ${errorParam}`);
      return;
    }

    if (!code) {
      setError('No authorization code received');
      return;
    }

    const codeVerifier = getCodeVerifier();
    if (!codeVerifier) {
      setError('Missing code verifier — please try logging in again');
      return;
    }

    const redirectUri = 'http://127.0.0.1:5173/callback';

    exchangeSpotifyToken(code, codeVerifier, redirectUri)
      .then((data) => {
        clearCodeVerifier();
        login(data.access_token, data.refresh_token, data.expires_in);
        navigate('/dashboard', { replace: true });
      })
      .catch((err) => {
        setError(`Token exchange failed: ${err.message}`);
      });
  }, [login, navigate]);

  if (error) {
    return (
      <div className="login-page">
        <div className="login-card">
          <div className="error-message">{error}</div>
          <button className="login-button" onClick={() => navigate('/')}>
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="loading-spinner">Authenticating...</div>
      </div>
    </div>
  );
}
