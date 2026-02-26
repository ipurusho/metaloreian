import { createContext, useContext, useState, useEffect, useCallback, useRef, type ReactNode } from 'react';
import { useAuth } from '../auth/AuthProvider';

interface CurrentTrack {
  id: string;
  name: string;
  artists: { name: string; uri: string }[];
  album: {
    name: string;
    images: { url: string; width: number; height: number }[];
    uri: string;
  };
  duration_ms: number;
}

type SDKStatus = 'loading' | 'ready' | 'error';

interface PlayerContextType {
  player: Spotify.Player | null;
  deviceId: string | null;
  sdkStatus: SDKStatus;
  sdkError: string | null;
  currentTrack: CurrentTrack | null;
  isPlaying: boolean;
  position: number;
  duration: number;
  togglePlay: () => void;
  nextTrack: () => void;
  prevTrack: () => void;
  seek: (positionMs: number) => void;
  transferPlayback: () => Promise<void>;
}

const PlayerContext = createContext<PlayerContextType | null>(null);

export function usePlayer() {
  const ctx = useContext(PlayerContext);
  if (!ctx) throw new Error('usePlayer must be used within PlayerProvider');
  return ctx;
}

declare global {
  interface Window {
    onSpotifyWebPlaybackSDKReady: () => void;
  }
}

export function PlayerProvider({ children }: { children: ReactNode }) {
  const { accessToken } = useAuth();
  const tokenRef = useRef(accessToken);
  tokenRef.current = accessToken;

  const [player, setPlayer] = useState<Spotify.Player | null>(null);
  const [deviceId, setDeviceId] = useState<string | null>(null);
  const [sdkStatus, setSdkStatus] = useState<SDKStatus>('loading');
  const [sdkError, setSdkError] = useState<string | null>(null);
  const [currentTrack, setCurrentTrack] = useState<CurrentTrack | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [position, setPosition] = useState(0);
  const [duration, setDuration] = useState(0);
  const positionInterval = useRef<number | null>(null);

  // Load Spotify SDK script
  useEffect(() => {
    if (document.getElementById('spotify-sdk')) return;
    const script = document.createElement('script');
    script.id = 'spotify-sdk';
    script.src = 'https://sdk.scdn.co/spotify-player.js';
    script.async = true;
    document.body.appendChild(script);
  }, []);

  // Initialize player when token is available
  useEffect(() => {
    if (!accessToken) return;

    let p: Spotify.Player | null = null;

    const init = () => {
      setSdkStatus('loading');
      setSdkError(null);

      p = new Spotify.Player({
        name: 'Metaloreian',
        getOAuthToken: (cb) => cb(tokenRef.current!),
        volume: 0.5,
      });

      p.addListener('ready', ({ device_id }) => {
        console.log('[Metaloreian] SDK ready, device:', device_id);
        setDeviceId(device_id);
        setSdkStatus('ready');

        // Auto-transfer playback to this device
        fetch('https://api.spotify.com/v1/me/player', {
          method: 'PUT',
          headers: {
            Authorization: `Bearer ${tokenRef.current}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ device_ids: [device_id], play: false }),
        }).then((res) => {
          if (res.ok) {
            console.log('[Metaloreian] Playback transferred');
          } else {
            console.warn('[Metaloreian] Transfer failed:', res.status);
          }
        }).catch((err) => {
          console.warn('[Metaloreian] Transfer error:', err);
        });
      });

      p.addListener('not_ready', ({ device_id }) => {
        console.warn('[Metaloreian] Device not ready:', device_id);
        setDeviceId(null);
      });

      p.addListener('initialization_error', ({ message }) => {
        console.error('[Metaloreian] Init error:', message);
        setSdkStatus('error');
        setSdkError(message);
      });

      p.addListener('authentication_error', ({ message }) => {
        console.error('[Metaloreian] Auth error:', message);
        setSdkStatus('error');
        setSdkError(message);
      });

      p.addListener('account_error', ({ message }) => {
        console.error('[Metaloreian] Account error:', message);
        setSdkStatus('error');
        setSdkError(message);
      });

      p.addListener('player_state_changed', (state) => {
        if (!state) {
          setCurrentTrack(null);
          setIsPlaying(false);
          return;
        }

        const track = state.track_window.current_track;
        setCurrentTrack({
          id: track.id,
          name: track.name,
          artists: track.artists.map((a) => ({ name: a.name, uri: a.uri })),
          album: {
            name: track.album.name,
            images: track.album.images,
            uri: track.album.uri,
          },
          duration_ms: track.duration_ms,
        });
        setIsPlaying(!state.paused);
        setPosition(state.position);
        setDuration(state.duration);
      });

      p.connect().then((success) => {
        if (success) {
          console.log('[Metaloreian] Player connected');
        } else {
          console.error('[Metaloreian] Player failed to connect');
          setSdkStatus('error');
          setSdkError('Failed to connect to Spotify');
        }
      });

      setPlayer(p);
    };

    if (window.Spotify) {
      init();
    } else {
      window.onSpotifyWebPlaybackSDKReady = init;
    }

    return () => {
      p?.disconnect();
    };
  }, [accessToken]);

  // Position tracking interval
  useEffect(() => {
    if (positionInterval.current) {
      clearInterval(positionInterval.current);
    }

    if (isPlaying) {
      positionInterval.current = window.setInterval(() => {
        setPosition((prev) => prev + 500);
      }, 500);
    }

    return () => {
      if (positionInterval.current) clearInterval(positionInterval.current);
    };
  }, [isPlaying]);

  const togglePlay = useCallback(() => {
    player?.togglePlay();
  }, [player]);

  const nextTrack = useCallback(() => {
    player?.nextTrack();
  }, [player]);

  const prevTrack = useCallback(() => {
    player?.previousTrack();
  }, [player]);

  const seek = useCallback(
    (positionMs: number) => {
      player?.seek(positionMs);
      setPosition(positionMs);
    },
    [player]
  );

  const transferPlayback = useCallback(async () => {
    if (!deviceId || !tokenRef.current) return;
    await fetch('https://api.spotify.com/v1/me/player', {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${tokenRef.current}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ device_ids: [deviceId], play: true }),
    });
  }, [deviceId]);

  return (
    <PlayerContext.Provider
      value={{
        player,
        deviceId,
        sdkStatus,
        sdkError,
        currentTrack,
        isPlaying,
        position,
        duration,
        togglePlay,
        nextTrack,
        prevTrack,
        seek,
        transferPlayback,
      }}
    >
      {children}
    </PlayerContext.Provider>
  );
}
