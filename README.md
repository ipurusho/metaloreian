# Metaloreian

Metaloreian unites [Encyclopedia Metallum](https://www.metal-archives.com/) with the Spotify Web Playback SDK so you can explore metal history while you listen. Search for any metal band, browse discographies and lineups, follow member links across bands, and play albums — all in one place.

## Architecture

| Layer | Tech | Role |
|-------|------|------|
| Frontend | React, TypeScript, Vite | Spotify Web Playback SDK, UI, routing |
| Backend | Go (chi, goquery, rod) | REST API, Metal Archives scraper, caching |
| Database | PostgreSQL | Optional cache for scraped MA data |
| Auth | Spotify OAuth2 PKCE | No separate account system — Spotify login is the identity |

### How it works

1. You log in with Spotify (PKCE flow — no client secret needed).
2. The embedded Spotify player lets you control playback directly in the browser.
3. When you search for a band, the Go backend fetches and parses data from Metal Archives.
4. Band pages show stats, current/past lineups, and full discographies. Each member lists their other bands as clickable links, so you can fall down the rabbit hole.
5. Album pages show tracklists and album-specific lineups.
6. Scraped data is cached in PostgreSQL (when available) to reduce load on Metal Archives.

## Project structure

```
metaloreian/
├── backend/
│   ├── cmd/server/          # Entry point
│   ├── internal/
│   │   ├── api/             # HTTP handlers, middleware, router
│   │   ├── config/          # Env-based configuration
│   │   ├── matcher/         # Spotify name → MA band matching, cache orchestration
│   │   ├── models/          # Shared data types
│   │   ├── scraper/         # MA HTML scraper (band, album, lineup, search)
│   │   └── store/           # PostgreSQL queries
│   └── migrations/          # SQL schema
├── frontend/
│   └── src/
│       ├── auth/            # PKCE flow, AuthProvider
│       ├── player/          # Spotify SDK, PlayerContext, PlayerBar
│       ├── features/        # Band, Album, Search pages
│       ├── api/             # API client + Spotify helpers
│       └── components/      # Shared UI components
├── docker-compose.yml
└── Makefile
```

## Prerequisites

- **Go 1.21+**
- **Node.js 18+**
- **Chromium** (or dependencies for headless Chrome — `libnss3`, `libasound2`)
- **PostgreSQL 15+** (optional — app runs in scrape-only mode without it)
- **Spotify Premium** account (required for Web Playback SDK)

## Local development

### 1. Clone and configure

```bash
git clone https://github.com/ipurusho/metaloreian.git
cd metaloreian
```

Create environment files from the examples:

```bash
cp .env.example .env
cp frontend/.env.example frontend/.env
```

Edit both files and set your Spotify Client ID. You can register an app at the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) — add `http://127.0.0.1:5173/callback` as a redirect URI.

### 2. Start PostgreSQL (optional)

With Docker:

```bash
docker compose up -d postgres
```

Or skip this step — the backend will run in scrape-only mode (no caching).

### 3. Run database migrations (if using PostgreSQL)

```bash
psql "$DATABASE_URL" -f backend/migrations/001_initial_schema.sql
```

### 4. Start the backend

```bash
cd backend
go run ./cmd/server
```

The server starts on `:8080`. Set `SPOTIFY_CLIENT_ID` and `DATABASE_URL` via environment or `.env`.

### 5. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

Opens at `http://127.0.0.1:5173`. The Vite dev server proxies `/api` requests to the backend.

### 6. Use the app

1. Open `http://127.0.0.1:5173` and click **Connect with Spotify**.
2. After auth, use the search bar to find a band.
3. Click a band to see their Metal Archives page — lineup, discography, and stats.
4. Click an album for the tracklist and album-specific lineup.
5. Click member names to explore their other bands.

## API endpoints

```
GET  /api/bands/search?q={query}   Search Metal Archives for bands
GET  /api/bands/{maId}             Band data + lineup + discography
GET  /api/albums/{albumId}         Album info + tracklist + lineup
POST /api/spotify/exchange         Proxy PKCE token exchange
POST /api/spotify/refresh          Proxy token refresh
GET  /health                       Health check
```

## Deployment

For production, use the provided `docker-compose.yml` as a starting point:

```bash
# Set required env vars
export SPOTIFY_CLIENT_ID=your_client_id
export POSTGRES_PASSWORD=a_strong_password
export FRONTEND_URL=https://your-domain.com

docker compose up -d
```

Key environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `SPOTIFY_CLIENT_ID` | Yes | From Spotify Developer Dashboard |
| `POSTGRES_PASSWORD` | Yes | PostgreSQL password (defaults to `metaloreian_dev` for local dev) |
| `DATABASE_URL` | No | Full connection string (built from `POSTGRES_PASSWORD` in compose) |
| `FRONTEND_URL` | No | CORS origin (defaults to `http://localhost:5173`) |
| `PORT` | No | Backend listen port (defaults to `8080`) |

The backend `Dockerfile` builds a static Go binary and serves the built frontend from `dist/`. In production, update the Spotify app's redirect URI to match your domain.

## Notes on Metal Archives integration

- A token bucket rate limiter (1 request / 3 seconds) keeps request volume low.
- `singleflight` deduplication ensures concurrent requests for the same band/album only trigger one fetch.
- Scraped data is cached in PostgreSQL with a 7-day TTL to minimize repeated requests.

## License

MIT
