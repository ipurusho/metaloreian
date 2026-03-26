import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getAlbum, getSimilarAlbums, spotifySearch, spotifyPlay } from '../../api/client';
import { useAuth } from '../../auth/AuthProvider';
import { usePlayer } from '../../player/PlayerContext';
import { MemberRow } from '../../components/MemberRow';
import { LoadingSpinner } from '../../components/LoadingSpinner';

export function AlbumPage() {
  const { maId, albumId } = useParams<{ maId: string; albumId: string }>();
  const { accessToken } = useAuth();
  const { deviceId } = usePlayer();

  const { data: album, isLoading, error } = useQuery({
    queryKey: ['album', albumId],
    queryFn: () => getAlbum(Number(albumId)),
    enabled: !!albumId,
  });

  const { data: similarAlbums, isLoading: similarLoading } = useQuery({
    queryKey: ['similarAlbums', albumId],
    queryFn: () => getSimilarAlbums(Number(albumId)),
    enabled: !!albumId,
    staleTime: 1000 * 60 * 60, // 1 hour — similar albums don't change
  });

  if (isLoading) return <LoadingSpinner message="Loading album data..." />;
  if (error) return <div className="error-message">Failed to load album: {(error as Error).message}</div>;
  if (!album) return <div className="error-message">Album not found</div>;

  const handlePlayAlbum = async () => {
    if (!accessToken || !deviceId) return;
    try {
      const results = await spotifySearch(`${album.band_name} ${album.name}`, 'album', accessToken);
      const spotifyAlbum = results.albums?.items?.[0];
      if (spotifyAlbum) {
        await spotifyPlay(accessToken, { context_uri: spotifyAlbum.uri });
      }
    } catch (err) {
      console.error('Failed to play album:', err);
    }
  };

  return (
    <div className="album-page">
      <div className="album-nav">
        <Link to={`/band/${maId || album.band_id}`} className="back-link">
          ← {album.band_name}
        </Link>
      </div>

      <div className="album-header">
        {album.cover_url && (
          <div className="album-cover">
            <img src={album.cover_url} alt={album.name} />
          </div>
        )}
        <div className="album-info">
          <h1 className="album-name">{album.name}</h1>
          <Link to={`/band/${album.band_id}`} className="album-band-name">
            {album.band_name}
          </Link>
          <div className="album-stats">
            <div className="stat"><span className="stat-label">Type:</span> {album.type}</div>
            <div className="stat"><span className="stat-label">Released:</span> {album.release_date}</div>
            {album.label && <div className="stat"><span className="stat-label">Label:</span> {album.label}</div>}
            {album.format && <div className="stat"><span className="stat-label">Format:</span> {album.format}</div>}
          </div>
          <button className="play-button" onClick={handlePlayAlbum} disabled={!accessToken || !deviceId}>
            ▶ Play Album on Spotify
          </button>
        </div>
      </div>

      <div className="album-content">
        <div className="album-tracklist">
          <h2 className="section-title">Tracklist</h2>
          {album.tracks && album.tracks.length > 0 ? (
            <table className="track-table">
              <thead>
                <tr>
                  <th className="track-num">#</th>
                  <th className="track-title">Title</th>
                  <th className="track-duration">Duration</th>
                </tr>
              </thead>
              <tbody>
                {album.tracks.map((track) => (
                  <tr key={track.id} className="track-row">
                    <td className="track-num">{track.track_number}</td>
                    <td className="track-title">{track.title}</td>
                    <td className="track-duration">{track.duration}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="empty-state">No tracklist data</p>
          )}
        </div>

        <div className="album-lineup-section">
          <h2 className="section-title">Album Lineup</h2>
          {album.lineup && album.lineup.length > 0 ? (
            <div className="lineup-list">
              {album.lineup.map((member) => (
                <MemberRow key={member.member_id} member={member} />
              ))}
            </div>
          ) : (
            <p className="empty-state">No lineup data</p>
          )}
        </div>
      </div>

      {similarLoading && <LoadingSpinner message="Finding similar albums..." />}
      {similarAlbums && similarAlbums.length > 0 && (
        <div className="similar-albums">
          <h2 className="section-title">Similar Albums <span className="beta-badge">(beta)</span> <span className="info-hint" title="Recommendations powered by a contrastive ML model trained on AcousticBrainz audio features. Finds albums that sound alike based on BPM, energy, loudness, key, and MFCC.">?</span></h2>
          <div className="similar-albums-list">
            {similarAlbums.map((sa) => (
              <div key={sa.album_id} className="similar-album-item">
                <div className="similar-album-info">
                  <span className="similar-album-name">{sa.name}</span>
                  <span className="similar-album-band">{sa.band_name}</span>
                  <span className="similar-album-type">{sa.type}</span>
                </div>
                <div className="similar-score">
                  <div className="similar-score-bar" style={{ width: `${Math.round(sa.score * 100)}%` }} />
                  <span className="similar-score-text">{Math.round(sa.score * 100)}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
