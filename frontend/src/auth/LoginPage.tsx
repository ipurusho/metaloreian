import { redirectToSpotifyAuth } from './pkce';
import { useAuth } from './AuthProvider';
import { Navigate } from 'react-router-dom';

export function LoginPage() {
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return <Navigate to="/band/125" replace />;
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <h1 className="login-title">METALOREIAN</h1>
        <p className="login-subtitle">
          Metal knowledge while you listen.
          <br />
          Spotify playback + Encyclopedia Metallum data.
        </p>
        <button className="login-button" onClick={redirectToSpotifyAuth}>
          Connect with Spotify
        </button>
        <p className="login-note">Requires Spotify Premium for playback</p>
      </div>
    </div>
  );
}
