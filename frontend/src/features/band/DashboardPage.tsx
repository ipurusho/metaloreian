import { useAuth } from '../../auth/AuthProvider';
import { usePlayer } from '../../player/PlayerContext';

export function DashboardPage() {
  const { accessToken } = useAuth();
  const { deviceId, sdkStatus, sdkError, currentTrack, isPlaying, transferPlayback } = usePlayer();

  return (
    <div className="band-page">
      <h1 className="band-name">METALOREIAN</h1>
      <div className="band-stats" style={{ maxWidth: 500, marginTop: 24 }}>
        <div className="stat">
          <span className="stat-label">Auth:</span>{' '}
          <span style={{ color: accessToken ? 'var(--success)' : 'var(--accent)' }}>
            {accessToken ? 'Connected' : 'Not connected'}
          </span>
        </div>
        <div className="stat">
          <span className="stat-label">SDK Status:</span>{' '}
          <span style={{ color: sdkStatus === 'ready' ? 'var(--success)' : sdkStatus === 'error' ? 'var(--accent)' : 'var(--warning)' }}>
            {sdkStatus === 'ready' ? `Ready (${deviceId?.slice(0, 8)}...)` : sdkStatus === 'error' ? `Error: ${sdkError}` : 'Loading...'}
          </span>
        </div>
        <div className="stat">
          <span className="stat-label">Playback:</span>{' '}
          {currentTrack
            ? `${isPlaying ? 'Playing' : 'Paused'}: ${currentTrack.artists[0]?.name} — ${currentTrack.name}`
            : 'No active track'}
        </div>
      </div>
      {sdkStatus === 'ready' && !currentTrack && (
        <div style={{ marginTop: 24 }}>
          <button
            onClick={transferPlayback}
            style={{
              background: '#1db954',
              color: '#fff',
              border: 'none',
              borderRadius: '24px',
              padding: '12px 32px',
              fontSize: '16px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Start Playback Here
          </button>
          <p className="empty-state" style={{ marginTop: 12 }}>
            This will transfer Spotify playback to the Metaloreian player.
            <br />
            If nothing is queued, open Spotify and play something first.
          </p>
        </div>
      )}
    </div>
  );
}
