import { usePlayer } from '../../player/PlayerContext';
import { SearchBar } from '../search/SearchBar';

export function DashboardPage() {
  const { sdkStatus, currentTrack, transferPlayback } = usePlayer();

  return (
    <div className="dashboard">
      <div className="dashboard-center">
        <h1 className="dashboard-title">METALOREIAN</h1>
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
