import { usePlayer } from '../../player/PlayerContext';

export function DashboardPage() {
  const { sdkStatus, currentTrack, transferPlayback } = usePlayer();

  return (
    <div className="band-page">
      <h1 className="band-name">METALOREIAN</h1>
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
