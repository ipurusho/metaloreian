import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { usePlayer } from '../../player/PlayerContext';
import { searchBands } from '../../api/client';

export function useAutoDetect() {
  const { currentTrack } = usePlayer();
  const navigate = useNavigate();
  const lastArtist = useRef<string | null>(null);
  const [enabled, setEnabled] = useState(true);

  useEffect(() => {
    if (!enabled || !currentTrack) return;

    const artistName = currentTrack.artists[0]?.name;
    if (!artistName || artistName === lastArtist.current) return;

    lastArtist.current = artistName;

    searchBands(artistName)
      .then((results) => {
        if (results && results.length > 0) {
          // Find exact or close match
          const exact = results.find(
            (r) => r.name.toLowerCase() === artistName.toLowerCase()
          );
          const match = exact || results[0];
          navigate(`/band/${match.ma_id}`);
        }
      })
      .catch(() => {
        // Silently fail — user can manually search
      });
  }, [currentTrack, enabled, navigate]);

  return { autoDetectEnabled: enabled, setAutoDetect: setEnabled };
}
