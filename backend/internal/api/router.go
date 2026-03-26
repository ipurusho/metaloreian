package api

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/imman/metaloreian/internal/config"
	"github.com/imman/metaloreian/internal/store"
)

func NewRouter(cfg *config.Config, s *store.Store, bf BandFetcher, af AlbumFetcher) *chi.Mux {
	r := chi.NewRouter()

	r.Use(SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(LoggingMiddleware)

	spotifyH := NewSpotifyHandlers(cfg)
	bandH := NewBandHandlers(s, bf)
	albumH := NewAlbumHandlers(s, af)
	similarAlbumH := NewSimilarAlbumHandlers(s)

	r.Route("/api", func(r chi.Router) {
		r.Use(RateLimitMiddleware(60)) // 60 requests/minute per IP
		r.Get("/bands/search", bandH.Search)
		r.Get("/bands/{maId}", bandH.Get)
		r.Get("/albums/{albumId}", albumH.Get)
		r.Get("/albums/{albumId}/similar", similarAlbumH.Get)
		r.Post("/spotify/exchange", spotifyH.Exchange)
		r.Post("/spotify/refresh", spotifyH.Refresh)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Serve frontend static files if dist directory exists
	distPath := filepath.Join(".", "dist")
	if _, err := os.Stat(distPath); err == nil {
		fileServer(r, distPath)
	}

	return r
}

func fileServer(r chi.Router, root string) {
	fsys := http.Dir(root)

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try serving the file directly
		if f, err := fs.Stat(os.DirFS(root), path); err == nil && !f.IsDir() {
			http.FileServer(fsys).ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for all unmatched routes
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}
