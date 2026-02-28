import { redirectToSpotifyAuth } from './pkce';
import { useAuth } from './AuthProvider';
import { Navigate } from 'react-router-dom';

export function LoginPage() {
  const { isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <svg className="dashboard-logo" viewBox="0 0 600 120" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <path id="login-arc" d="M 30,100 Q 300,0 570,100" fill="none" />
          </defs>
          <text>
            <textPath href="#login-arc" startOffset="50%" textAnchor="middle">metalöreian</textPath>
          </text>
        </svg>
        <img className="login-mascot" src="/images/snaggletooth.png" alt="Metal Nerd Snaggletooth" />
        <svg className="dashboard-tagline" viewBox="0 0 800 100" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <path id="login-tagline-arc" d="M 20,10 Q 400,110 780,10" fill="none" />
          </defs>
          <text>
            <textPath href="#login-tagline-arc" startOffset="50%" textAnchor="middle">music nerds</textPath>
          </text>
        </svg>
        <p className="login-subtitle">Metal knowledge while you listen.</p>
        <p className="login-subtitle">Spotify playback + Encyclopedia Metallum data.</p>
        <button className="login-button" onClick={redirectToSpotifyAuth}>
          Connect with Spotify
        </button>
        <p className="login-note">Requires Spotify Premium for playback</p>
      </div>
    </div>
  );
}
