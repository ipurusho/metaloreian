import { usePlayer } from './PlayerContext';

function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

export function PlayerBar() {
  const { currentTrack, isPlaying, position, duration, togglePlay, nextTrack, prevTrack, seek } =
    usePlayer();

  if (!currentTrack) {
    return (
      <div className="player-bar">
        <div className="player-empty">No track playing — play something on Spotify</div>
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
          ⏮
        </button>
        <button className="player-btn player-btn-play" onClick={togglePlay} title={isPlaying ? 'Pause' : 'Play'}>
          {isPlaying ? '⏸' : '▶'}
        </button>
        <button className="player-btn" onClick={nextTrack} title="Next">
          ⏭
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
