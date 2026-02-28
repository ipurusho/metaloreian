import { usePlayer } from '../../player/PlayerContext';
import { SearchBar } from '../search/SearchBar';

export function DashboardPage() {
  const { sdkStatus, currentTrack, transferPlayback } = usePlayer();

  return (
    <div className="dashboard">
      <div className="dashboard-center">
        <svg className="dashboard-logo" viewBox="0 0 600 120" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <path id="arc" d="M 30,100 Q 300,0 570,100" fill="none" />
          </defs>
          <text>
            <textPath href="#arc" startOffset="50%" textAnchor="middle">metalöreian</textPath>
          </text>
        </svg>
        <img className="login-mascot" src="/images/snaggletooth.png" alt="Metal Nerd Snaggletooth" />
        <svg className="dashboard-tagline" viewBox="0 0 800 100" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <path id="dash-tagline-arc" d="M 20,10 Q 400,110 780,10" fill="none" />
          </defs>
          <text>
            <textPath href="#dash-tagline-arc" startOffset="50%" textAnchor="middle">music nerds</textPath>
          </text>
        </svg>
        <p className="dashboard-subtitle">Metal knowledge while you listen.</p>
        <p className="dashboard-subtitle">Spotify playback + Encyclopedia Metallum data.</p>
        <div className="dashboard-search">
          <SearchBar />
        </div>
        {sdkStatus === 'ready' && !currentTrack && (
          <div className="dashboard-playback">
            <button className="playback-btn" onClick={transferPlayback}>
              Start Playback Here
            </button>
            <p className="empty-state">
              This will transfer Spotify playback to the Metaloreian player.
              <br />
              If nothing is queued, open Spotify and play something first.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
