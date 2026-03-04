# Metaloreian — Project Guide

## What Is This?
A full-stack web app that integrates Metal Archives (Encyclopedia Metallum) with Spotify Web Playback SDK. Search metal bands, view lineups/discographies, and play albums via Spotify.

## Tech Stack
- **Frontend:** React 18 + TypeScript + Vite (Spotify Web Playback SDK)
- **Backend:** Go 1.23 (chi router, goquery scraper, net/http + FlareSolverr fallback)
- **Database:** PostgreSQL 16 (caching layer, optional for local dev)
- **Auth:** Spotify OAuth2 PKCE (no separate user accounts)
- **Reverse Proxy:** Nginx 1.27 with TLS 1.2/1.3
- **CI/CD:** GitHub Actions → GHCR → SSH deploy
- **SSL:** Let's Encrypt via certbot DNS-01 + DuckDNS

## Key Paths
```
backend/cmd/server/main.go       # Backend entry point
backend/internal/api/             # HTTP handlers & router
backend/internal/scraper/         # Metal Archives HTML scraper
backend/internal/store/store.go   # PostgreSQL queries
backend/migrations/               # SQL migrations
frontend/src/                     # React app source
nginx.conf                        # Reverse proxy config
docker-compose.prod.yml           # Production compose
docker-compose.yml                # Local dev compose
.github/workflows/deploy.yml      # CI/CD pipeline
Makefile                           # Dev commands
```

## Local Development
```bash
# Start postgres + FlareSolverr
docker compose up -d postgres
docker run -d --name flaresolverr -p 8191:8191 ghcr.io/flaresolverr/flaresolverr:latest

# Backend (reads .env)
cd backend && FLARESOLVERR_URL=http://localhost:8191 go run ./cmd/server

# Frontend
cd frontend && npm install && npm run dev

# Or with docker compose (starts everything):
docker compose up -d
```

## API Endpoints
- `GET /api/bands/search?q={query}` — Search Metal Archives
- `GET /api/bands/{maId}` — Band data + lineup + discography
- `GET /api/albums/{albumId}` — Album info + tracklist + lineup
- `POST /api/spotify/exchange` — PKCE token exchange proxy
- `POST /api/spotify/refresh` — Token refresh proxy
- `GET /health` — Health check

## Production Deploy
- Push to `main` → GitHub Actions builds Docker image → pushes to GHCR → SSHs into VM → docker compose up
- Image: `ghcr.io/ipurusho/metaloreian:latest`
- All secrets managed via GitHub repo secrets (never committed)

## Important Notes
- Metal Archives rate limiter: 1 request / 3 seconds (token bucket)
- Scraper uses plain net/http with `X-Requested-With: XMLHttpRequest` header to bypass Cloudflare in normal mode
- When Cloudflare is strict, falls back to FlareSolverr sidecar (headless browser that solves CF challenges)
- Docker image is distroless (~30MB) — no Chromium, no shell
- CORS: `FRONTEND_URL` env var must match domain exactly
