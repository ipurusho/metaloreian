import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getBand, spotifySearch, spotifyPlay } from '../../api/client';
import { useAuth } from '../../auth/AuthProvider';
import { usePlayer } from '../../player/PlayerContext';
import { MemberRow } from '../../components/MemberRow';
import { LoadingSpinner } from '../../components/LoadingSpinner';
import { Link } from 'react-router-dom';

export function BandPage() {
  const { maId } = useParams<{ maId: string }>();
  const { accessToken } = useAuth();
  const { deviceId } = usePlayer();

  const { data: band, isLoading, error } = useQuery({
    queryKey: ['band', maId],
    queryFn: () => getBand(Number(maId)),
    enabled: !!maId,
  });

  if (isLoading) return <LoadingSpinner message="Loading band data..." />;
  if (error) return <div className="error-message">Failed to load band: {(error as Error).message}</div>;
  if (!band) return <div className="error-message">Band not found</div>;

  const handlePlayTopTracks = async () => {
    if (!accessToken || !deviceId) return;
    try {
      const results = await spotifySearch(band.name, 'artist', accessToken);
      const artist = results.artists?.items?.[0];
      if (artist) {
        await spotifyPlay(accessToken, { context_uri: artist.uri });
      }
    } catch (err) {
      console.error('Failed to play:', err);
    }
  };

  return (
    <div className="band-page">
      <div className="band-header">
        {band.photo_url && (
          <div className="band-photo">
            <img src={band.photo_url} alt={band.name} />
          </div>
        )}
        <div className="band-info">
          <h1 className="band-name">{band.name}</h1>
          {band.logo_url && (
            <img className="band-logo" src={band.logo_url} alt={`${band.name} logo`} />
          )}
          <div className="band-stats">
            <div className="stat"><span className="stat-label">Genre:</span> {band.genre}</div>
            <div className="stat"><span className="stat-label">Country:</span> {band.country}</div>
            <div className="stat"><span className="stat-label">Status:</span> <span className={`status-${band.status.toLowerCase().replace(/\s+/g, '-')}`}>{band.status}</span></div>
            <div className="stat"><span className="stat-label">Themes:</span> {band.themes}</div>
            <div className="stat"><span className="stat-label">Formed:</span> {band.formed_in}</div>
            <div className="stat"><span className="stat-label">Years active:</span> {band.years_active}</div>
          </div>
          <button className="play-button" onClick={handlePlayTopTracks} disabled={!accessToken || !deviceId}>
            ▶ Play on Spotify
          </button>
        </div>
      </div>

      <div className="band-content">
        <div className="band-discography">
          <h2 className="section-title">Discography</h2>
          {band.discography && band.discography.length > 0 ? (
            <div className="discography-list">
              {band.discography.map((album) => (
                <Link
                  key={album.album_id}
                  to={`/band/${band.ma_id}/album/${album.album_id}`}
                  className="discography-item"
                >
                  <span className="album-name">{album.name}</span>
                  <span className="album-type">{album.type}</span>
                  <span className="album-year">{album.release_date}</span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="empty-state">No discography data</p>
          )}
        </div>

        <div className="band-lineup">
          <h2 className="section-title">Current Lineup</h2>
          {band.current_lineup && band.current_lineup.length > 0 ? (
            <div className="lineup-list">
              {band.current_lineup.map((member) => (
                <MemberRow key={member.member_id} member={member} />
              ))}
            </div>
          ) : (
            <p className="empty-state">No lineup data</p>
          )}

          {band.past_lineup && band.past_lineup.length > 0 && (
            <>
              <h2 className="section-title">Past Members</h2>
              <div className="lineup-list">
                {band.past_lineup.map((member) => (
                  <MemberRow key={member.member_id} member={member} />
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
