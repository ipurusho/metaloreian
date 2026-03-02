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

The app deploys to an Oracle Cloud Always Free ARM VM with automated CI/CD via GitHub Actions. HTTPS is provided by Let's Encrypt via a free [is-a.dev](https://is-a.dev/) subdomain.

### Architecture

```
GitHub push → GitHub Actions → Build ARM64 image → Push to GHCR → SSH deploy to VM

VM:
  nginx (80/443) → backend:8080 (Go + frontend static files)
                    postgres:5432 (internal only)
  certbot (auto-renews Let's Encrypt certs)
```

### Prerequisites

- Oracle Cloud Always Free ARM A1.Flex VM (Ubuntu)
- Domain: `metaloreian.is-a.dev` (A record → VM IP)
- Spotify Developer Dashboard: `https://metaloreian.is-a.dev/callback` as redirect URI

### VM initial setup (one-time)

```bash
# 1. Install Docker
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER
# Log out and back in for group change to take effect

# 2. Open firewall (iptables + Oracle VCN Security List for 80/443)
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
sudo netfilter-persistent save

# 3. Generate deploy SSH key
ssh-keygen -t ed25519 -f ~/.ssh/deploy_key -N ""
cat ~/.ssh/deploy_key.pub >> ~/.ssh/authorized_keys
# Add the private key (~/.ssh/deploy_key) as GitHub secret VM_SSH_KEY

# 4. Clone repo
git clone https://github.com/ipurusho/metaloreian.git ~/metaloreian
cd ~/metaloreian

# 5. Create .env
cat > .env << 'EOF'
SPOTIFY_CLIENT_ID=your_spotify_client_id
POSTGRES_PASSWORD=a_strong_password
FRONTEND_URL=https://metaloreian.is-a.dev
EOF

# 6. Initial deploy
docker compose -f docker-compose.prod.yml up -d

# 7. Obtain SSL certificate
docker compose -f docker-compose.prod.yml run --rm certbot certonly \
  --webroot -w /var/www/certbot -d metaloreian.is-a.dev

# 8. Reload nginx to pick up the real cert
docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

### is-a.dev domain registration

Fork [is-a-dev/register](https://github.com/is-a-dev/register), create `domains/metaloreian.json`:

```json
{
  "owner": {
    "username": "ipurusho"
  },
  "record": {
    "A": ["<VM_PUBLIC_IP>"]
  }
}
```

Open a PR — typically merges within hours.

### GitHub secrets

| Secret | Value |
|--------|-------|
| `VM_HOST` | Oracle VM public IP |
| `VM_USER` | `ubuntu` |
| `VM_SSH_KEY` | Deploy key (ed25519 private key) |
| `VITE_SPOTIFY_CLIENT_ID` | Spotify Client ID (baked into frontend at build time) |

### CI/CD workflow

On every push to `main`, GitHub Actions:
1. Builds an ARM64 Docker image (QEMU cross-compilation)
2. Pushes to `ghcr.io/ipurusho/metaloreian:latest` and `:<sha>`
3. SSHs into the VM, pulls the new image, and restarts

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
| `FRONTEND_URL` | No | CORS origin (defaults to `https://metaloreian.is-a.dev`) |
| `PORT` | No | Backend listen port (defaults to `8080`) |

### Key considerations

- **CORS**: `FRONTEND_URL` must match the domain exactly (`https://metaloreian.is-a.dev`)
- **Oracle idle VM reclaim**: VMs idle for 7 days (CPU/memory < 20%) may be reclaimed. Chromium usage should keep memory above threshold.
- **Database persistence**: The `pgdata` volume survives container restarts. Only `docker compose down -v` destroys it.

## Notes on Metal Archives integration

- A token bucket rate limiter (1 request / 3 seconds) keeps request volume low.
- `singleflight` deduplication ensures concurrent requests for the same band/album only trigger one fetch.
- Scraped data is cached in PostgreSQL with a 7-day TTL to minimize repeated requests.

## License

MIT
