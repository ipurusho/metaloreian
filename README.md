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

## Production deployment

The app deploys to an AWS EC2 t3.small instance (2 vCPU, 2GB RAM) with automated CI/CD via GitHub Actions. HTTPS is provided by Let's Encrypt via DNS-01 challenge with [DuckDNS](https://www.duckdns.org/).

### Architecture

```
GitHub push → GitHub Actions → Build amd64 image → Push to GHCR → SSH deploy to VM

EC2 t3.small (2GB RAM):
  nginx (80/443)  → backend:8080 (Go + Chromium + frontend static)
                     postgres:5432 (internal only)
  certbot (DNS-01 via DuckDNS)
```

### Prerequisites

- AWS EC2 t3.small instance (Ubuntu 22.04)
- Domain: `metaloreian-dev.duckdns.org` (DuckDNS → Elastic IP)
- Spotify Developer Dashboard: `https://metaloreian-dev.duckdns.org/callback` as redirect URI

### VM initial setup (one-time)

```bash
# 1. Install Docker
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER
# Log out and back in for group change to take effect

# 2. Create deploy user
sudo useradd -m -s /bin/bash metaloreian
sudo usermod -aG docker metaloreian

# 3. Configure firewall
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# 4. Security hardening
sudo apt-get install -y fail2ban unattended-upgrades

# 5. Set up SSH for deploy user
sudo -u metaloreian bash -c '
  mkdir -p ~/.ssh && chmod 700 ~/.ssh
'
# Copy authorized_keys to deploy user

# 6. Create app directory
sudo -u metaloreian mkdir -p ~/app

# 7. Initial deploy (triggered by GitHub Actions push to main)

# 8. Obtain SSL certificate (DNS-01 via DuckDNS)
sudo -u metaloreian bash -c 'cd ~/app && \
  docker compose -f docker-compose.prod.yml run --rm certbot certonly \
    --authenticator dns-duckdns \
    --dns-duckdns-token $DUCKDNS_TOKEN \
    --dns-duckdns-propagation-seconds 60 \
    -d metaloreian-dev.duckdns.org --agree-tos -m your@email.com'

# 9. Reload nginx to pick up the real cert
sudo -u metaloreian bash -c 'cd ~/app && docker compose -f docker-compose.prod.yml exec nginx nginx -s reload'
```

### GitHub secrets

| Secret | Value |
|--------|-------|
| `VM_HOST` | EC2 Elastic IP |
| `VM_USER` | `metaloreian` |
| `VM_SSH_KEY` | Deploy key (ed25519 private key) |
| `VM_SSH_KNOWN_HOSTS` | Output of `ssh-keyscan -H <elastic-ip>` |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `SPOTIFY_CLIENT_ID` | Spotify Client ID (runtime) |
| `DUCKDNS_TOKEN` | DuckDNS token for DNS-01 cert renewal |
| `FRONTEND_URL` | `https://metaloreian-dev.duckdns.org` |
| `VITE_SPOTIFY_CLIENT_ID` | Spotify Client ID (baked into frontend at build time) |

### CI/CD workflow

On every push to `main`, GitHub Actions:
1. Builds an amd64 Docker image
2. Pushes to `ghcr.io/ipurusho/metaloreian:latest` and `:<sha>`
3. SSHs into the VM, copies config files, writes `.env`, pulls the new image, and restarts containers

### Rollback

Every commit produces a tagged image. To roll back:

```bash
# On the VM, edit docker-compose.prod.yml to pin a specific sha
# image: ghcr.io/ipurusho/metaloreian:<commit-sha>
docker compose -f docker-compose.prod.yml up -d
```

### Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SPOTIFY_CLIENT_ID` | Yes | From Spotify Developer Dashboard |
| `POSTGRES_PASSWORD` | Yes | PostgreSQL password |
| `DUCKDNS_TOKEN` | Yes | DuckDNS token for DNS-01 cert challenges |
| `FRONTEND_URL` | No | CORS origin (defaults to `https://metaloreian-dev.duckdns.org`) |
| `PORT` | No | Backend listen port (defaults to `8080`) |

### Key considerations

- **CORS**: `FRONTEND_URL` must match the domain exactly (`https://metaloreian-dev.duckdns.org`)
- **Database persistence**: The `pgdata` volume survives container restarts. Only `docker compose down -v` destroys it.

## Notes on Metal Archives integration

- A token bucket rate limiter (1 request / 3 seconds) keeps request volume low.
- `singleflight` deduplication ensures concurrent requests for the same band/album only trigger one fetch.
- Scraped data is cached in PostgreSQL with a 7-day TTL to minimize repeated requests.

## License

MIT
