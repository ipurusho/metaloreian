const API_BASE = '/api';

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// Metal Archives API

export interface BandSearchResult {
  ma_id: number;
  name: string;
  genre: string;
  country: string;
}

export interface Band {
  ma_id: number;
  name: string;
  genre: string;
  country: string;
  status: string;
  themes: string;
  formed_in: string;
  years_active: string;
  logo_url: string;
  photo_url: string;
}

export interface Album {
  album_id: number;
  band_id: number;
  name: string;
  type: string;
  release_date: string;
  label: string;
  format: string;
  cover_url: string;
}

export interface Track {
  id: number;
  album_id: number;
  track_number: number;
  title: string;
  duration: string;
}

export interface MemberBand {
  member_id: number;
  band_id: number;
  band_name: string;
}

export interface Member {
  member_id: number;
  name: string;
  instrument: string;
  lineup_type?: string;
  years?: string;
  other_bands?: MemberBand[];
}

export interface BandFull extends Band {
  current_lineup: Member[];
  past_lineup?: Member[];
  discography: Album[];
}

export interface AlbumFull extends Album {
  band_name: string;
  tracks: Track[];
  lineup: Member[];
}

export function searchBands(query: string): Promise<BandSearchResult[]> {
  return fetchJSON(`${API_BASE}/bands/search?q=${encodeURIComponent(query)}`);
}

export function getBand(maId: number): Promise<BandFull> {
  return fetchJSON(`${API_BASE}/bands/${maId}`);
}

export function getAlbum(albumId: number): Promise<AlbumFull> {
  return fetchJSON(`${API_BASE}/albums/${albumId}`);
}

export interface SimilarAlbum {
  album_id: number;
  name: string;
  band_name: string;
  type: string;
  year: string;
  cover_url: string;
  score: number;
}

export function getSimilarAlbums(albumId: number): Promise<SimilarAlbum[]> {
  return fetchJSON(`${API_BASE}/albums/${albumId}/similar`);
}

// Spotify token exchange — direct to Spotify (no backend proxy needed)

const SPOTIFY_CLIENT_ID = import.meta.env.VITE_SPOTIFY_CLIENT_ID;
const SPOTIFY_TOKEN_URL = 'https://accounts.spotify.com/api/token';

export async function exchangeSpotifyToken(code: string, codeVerifier: string, redirectUri: string) {
  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    code,
    redirect_uri: redirectUri,
    client_id: SPOTIFY_CLIENT_ID,
    code_verifier: codeVerifier,
  });

  const res = await fetch(SPOTIFY_TOKEN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error_description || err.error || `HTTP ${res.status}`);
  }

  return res.json() as Promise<{ access_token: string; refresh_token: string; expires_in: number }>;
}

export async function refreshSpotifyToken(refreshToken: string) {
  const body = new URLSearchParams({
    grant_type: 'refresh_token',
    refresh_token: refreshToken,
    client_id: SPOTIFY_CLIENT_ID,
  });

  const res = await fetch(SPOTIFY_TOKEN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error_description || err.error || `HTTP ${res.status}`);
  }

  return res.json() as Promise<{ access_token: string; refresh_token?: string; expires_in: number }>;
}

// Spotify Web API helpers

const SPOTIFY_API = 'https://api.spotify.com/v1';

export function spotifyFetch<T>(path: string, token: string, options?: RequestInit): Promise<T> {
  return fetchJSON(`${SPOTIFY_API}${path}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${token}`,
      ...options?.headers,
    },
  });
}

export function spotifySearch(query: string, type: string, token: string) {
  return spotifyFetch<any>(
    `/search?q=${encodeURIComponent(query)}&type=${encodeURIComponent(type)}&limit=5`,
    token
  );
}

export function spotifyPlay(token: string, body: { context_uri?: string; uris?: string[]; offset?: { position: number } }) {
  return fetch(`${SPOTIFY_API}/me/player/play`, {
    method: 'PUT',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
}
