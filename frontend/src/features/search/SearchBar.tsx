import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { spotifySearch, spotifyPlay, searchBands } from '../../api/client';
import { useAuth } from '../../auth/AuthProvider';
import { usePlayer } from '../../player/PlayerContext';

interface SpotifyArtist {
  id: string;
  name: string;
  genres: string[];
  uri: string;
  images: { url: string }[];
}

export function SearchBar() {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const { accessToken } = useAuth();
  const { deviceId } = usePlayer();
  const navigate = useNavigate();
  const containerRef = useRef<HTMLDivElement>(null);

  // Debounce input
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const { data: results } = useQuery({
    queryKey: ['spotify-search', debouncedQuery],
    queryFn: async () => {
      if (!accessToken) return [];
      const res = await spotifySearch(debouncedQuery, 'artist', accessToken);
      return (res.artists?.items || []) as SpotifyArtist[];
    },
    enabled: debouncedQuery.length >= 2 && !!accessToken,
  });

  const handleSelect = async (artist: SpotifyArtist) => {
    setQuery('');
    setIsOpen(false);

    // Start playback
    if (accessToken && deviceId) {
      spotifyPlay(accessToken, { context_uri: artist.uri }).catch(() => {});
    }

    // Search MA for the band and navigate to its page
    searchBands(artist.name)
      .then((maResults) => {
        if (maResults && maResults.length > 0) {
          const exact = maResults.find(
            (r) => r.name.toLowerCase() === artist.name.toLowerCase()
          );
          const match = exact || maResults[0];
          navigate(`/band/${match.ma_id}`);
        }
      })
      .catch(() => {});
  };

  return (
    <div className="search-container" ref={containerRef}>
      <input
        className="search-input"
        type="text"
        placeholder="Search artists..."
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
        }}
        onFocus={() => setIsOpen(true)}
      />
      {isOpen && results && results.length > 0 && (
        <div className="search-results">
          {results.map((artist) => (
            <button
              key={artist.id}
              className="search-result-item"
              onClick={() => handleSelect(artist)}
            >
              <span className="search-result-name">{artist.name}</span>
              <span className="search-result-meta">
                {artist.genres.slice(0, 3).join(', ') || 'No genres listed'}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
