import { usePlayer } from './PlayerContext';

function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

export function PlayerBar() {
  const {
    sdkStatus,
    sdkError,
    deviceId,
    currentTrack,
    isPlaying,
    position,
    duration,
    togglePlay,
    nextTrack,
    prevTrack,
    seek,
    transferPlayback,
  } = usePlayer();

  if (sdkStatus === 'error') {
    return (
      <div className="player-bar">
        <div className="player-empty" style={{ color: 'var(--accent)' }}>
          Player error: {sdkError} (Spotify Premium required)
        </div>
      </div>
    );
  }

  if (sdkStatus === 'loading' || !deviceId) {
    return (
      <div className="player-bar">
        <div className="player-empty">Connecting to Spotify...</div>
      </div>
    );
  }

  if (!currentTrack) {
    return (
      <div className="player-bar">
        <div className="player-empty">
          Player ready.{' '}
          <button
            onClick={transferPlayback}
            style={{
              background: '#1db954',
              color: '#fff',
              border: 'none',
              borderRadius: '16px',
              padding: '6px 16px',
              fontWeight: 600,
              cursor: 'pointer',
              marginLeft: '8px',
            }}
          >
            Start Playback
          </button>
        </div>
      </div>
    );
  }

  const albumArt = currentTrack.album.images[currentTrack.album.images.length - 1]?.url;
  const progress = duration > 0 ? (position / duration) * 100 : 0;

  const handleProgressClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const pct = (e.clientX - rect.left) / rect.width;
    seek(Math.floor(pct * duration));
  };

  return (
    <div className="player-bar">
      <div className="player-track-info">
        {albumArt && <img className="player-album-art" src={albumArt} alt={currentTrack.album.name} />}
        <div className="player-text">
          <div className="player-track-name">{currentTrack.name}</div>
          <div className="player-artist-name">
            {currentTrack.artists.map((a) => a.name).join(', ')} — {currentTrack.album.name}
          </div>
        </div>
      </div>

      <div className="player-controls">
        <button className="player-btn" onClick={prevTrack} title="Previous">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M3 2h2v12H3V2zm11 0L7 8l7 6V2z" />
          </svg>
        </button>
        <button className="player-btn player-btn-play" onClick={togglePlay} title={isPlaying ? 'Pause' : 'Play'}>
          {isPlaying ? (
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M3 2h4v12H3V2zm6 0h4v12H9V2z" />
            </svg>
          ) : (
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M4 2l10 6-10 6V2z" />
            </svg>
          )}
        </button>
        <button className="player-btn" onClick={nextTrack} title="Next">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M2 2l7 6-7 6V2zm9 0h2v12h-2V2z" />
          </svg>
        </button>
      </div>

      <div className="player-progress-section">
        <span className="player-time">{formatTime(position)}</span>
        <div className="player-progress-bar" onClick={handleProgressClick}>
          <div className="player-progress-fill" style={{ width: `${progress}%` }} />
        </div>
        <span className="player-time">{formatTime(duration)}</span>
      </div>
    </div>
  );
}
