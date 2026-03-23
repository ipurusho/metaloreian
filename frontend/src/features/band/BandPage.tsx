import { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getBand, getSimilarBands, spotifySearch, spotifyPlay } from '../../api/client';
import { useAuth } from '../../auth/AuthProvider';
import { usePlayer } from '../../player/PlayerContext';
import { BandLink } from '../../components/BandLink';
import { MemberRow } from '../../components/MemberRow';
import { LoadingSpinner } from '../../components/LoadingSpinner';
import { Link } from 'react-router-dom';

const ALL_FILTER = 'All';

export function BandPage() {
  const { maId } = useParams<{ maId: string }>();
  const { accessToken } = useAuth();
  const { deviceId } = usePlayer();
  const [albumFilter, setAlbumFilter] = useState('Full-length');

  const { data: band, isLoading, error } = useQuery({
    queryKey: ['band', maId],
    queryFn: () => getBand(Number(maId)),
    enabled: !!maId,
  });

  const { data: similarBands, isLoading: similarLoading } = useQuery({
    queryKey: ['similar', maId],
    queryFn: () => getSimilarBands(Number(maId)),
    enabled: !!maId,
    staleTime: 1000 * 60 * 60, // 1 hour — similar bands don't change
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
            <>
              <DiscographyFilters
                albums={band.discography}
                activeFilter={albumFilter}
                onFilterChange={setAlbumFilter}
              />
              <div className="discography-list">
                {band.discography
                  .filter((album) => albumFilter === ALL_FILTER || album.type === albumFilter)
                  .map((album) => (
                    <div key={album.album_id} className="discography-item">
                      {accessToken && deviceId && (
                        <AlbumPlayButton
                          albumName={album.name}
                          bandName={band.name}
                          accessToken={accessToken}
                        />
                      )}
                      <Link
                        to={`/band/${band.ma_id}/album/${album.album_id}`}
                        className="album-name"
                      >
                        {album.name}
                      </Link>
                      <span className="album-year">{album.release_date.match(/\d{4}/)?.[0] ?? album.release_date}</span>
                    </div>
                  ))}
              </div>
            </>
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

      {similarLoading && <LoadingSpinner message="Finding similar bands..." />}
      {similarBands && similarBands.length > 0 && (
        <div className="similar-bands">
          <h2 className="section-title">Sonically Similar Bands <span className="beta-badge">(beta)</span> <span className="info-hint" title="Recommendations powered by a contrastive ML model trained on AcousticBrainz audio features (BPM, energy, loudness, key, MFCC). Finds bands that sound alike, not just share a genre.">?</span></h2>
          <div className="similar-bands-list">
            {similarBands.map((sb) => (
              <div key={sb.ma_id} className="similar-band-item">
                <div className="similar-band-info">
                  <BandLink bandId={sb.ma_id} bandName={sb.name} className="similar-band-name" />
                  <span className="similar-band-genre">{sb.genre}</span>
                  <span className="similar-band-country">{sb.country}</span>
                </div>
                <div className="similar-score">
                  <div className="similar-score-bar" style={{ width: `${Math.round(sb.score * 100)}%` }} />
                  <span className="similar-score-text">{Math.round(sb.score * 100)}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function AlbumPlayButton({ albumName, bandName, accessToken }: {
  albumName: string;
  bandName: string;
  accessToken: string;
}) {
  const [status, setStatus] = useState<'idle' | 'loading' | 'failed'>('idle');

  const handlePlay = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (status === 'loading') return;

    setStatus('loading');
    try {
      const results = await spotifySearch(`album:${albumName} artist:${bandName}`, 'album', accessToken);
      const album = results.albums?.items?.[0];
      if (album) {
        await spotifyPlay(accessToken, { context_uri: album.uri });
        setStatus('idle');
      } else {
        setStatus('failed');
        setTimeout(() => setStatus('idle'), 2000);
      }
    } catch {
      setStatus('failed');
      setTimeout(() => setStatus('idle'), 2000);
    }
  };

  return (
    <button
      className={`album-play-btn${status === 'failed' ? ' not-found' : ''}`}
      onClick={handlePlay}
      title={status === 'failed' ? 'Not found on Spotify' : `Play ${albumName}`}
    >
      {status === 'loading' ? '...' : status === 'failed' ? '✕' : '▶'}
    </button>
  );
}

function DiscographyFilters({ albums, activeFilter, onFilterChange }: {
  albums: { type: string }[];
  activeFilter: string;
  onFilterChange: (filter: string) => void;
}) {
  const types = useMemo(() => {
    const seen = new Set<string>();
    for (const a of albums) {
      if (a.type) seen.add(a.type);
    }
    const priority = ['Full-length', 'EP', 'Live album', 'Split'];
    const ordered = priority.filter((t) => seen.has(t));
    const rest = Array.from(seen).filter((t) => !priority.includes(t)).sort();
    return [...ordered, ...rest, ALL_FILTER];
  }, [albums]);

  if (types.length <= 2) return null;

  return (
    <div className="discography-filters">
      {types.map((type) => (
        <button
          key={type}
          className={`discography-filter${activeFilter === type ? ' active' : ''}`}
          onClick={() => onFilterChange(type)}
        >
          {type}
        </button>
      ))}
    </div>
  );
}
