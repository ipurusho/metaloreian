package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
)

// mockAlbumFetcher implements AlbumFetcher for testing.
type mockAlbumFetcher struct {
	fetchFn func(albumID int64) (*models.AlbumFull, error)
}

func (m *mockAlbumFetcher) FetchAlbum(albumID int64) (*models.AlbumFull, error) {
	return m.fetchFn(albumID)
}

func TestAlbumGet_InvalidID(t *testing.T) {
	h := NewAlbumHandlers(nil, &mockAlbumFetcher{})

	tests := []struct {
		name string
		id   string
	}{
		{"not a number", "xyz"},
		{"zero", "0"},
		{"negative", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/api/albums/{albumId}", h.Get)

			req := httptest.NewRequest("GET", "/api/albums/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 400 {
				t.Errorf("status = %d, want 400", w.Code)
			}
		})
	}
}

func TestAlbumGet_Success(t *testing.T) {
	fetcher := &mockAlbumFetcher{
		fetchFn: func(albumID int64) (*models.AlbumFull, error) {
			return &models.AlbumFull{
				Album: models.Album{AlbumID: albumID, Name: "Blackwater Park"},
				Tracks: []models.Track{
					{TrackNumber: 1, Title: "The Leper Affinity", Duration: "10:23"},
				},
			}, nil
		},
	}
	h := NewAlbumHandlers(nil, fetcher)

	r := chi.NewRouter()
	r.Get("/api/albums/{albumId}", h.Get)

	req := httptest.NewRequest("GET", "/api/albums/1234", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var album models.AlbumFull
	if err := json.NewDecoder(w.Body).Decode(&album); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if album.Name != "Blackwater Park" {
		t.Errorf("album name = %q, want Blackwater Park", album.Name)
	}
	if len(album.Tracks) != 1 {
		t.Errorf("tracks = %d, want 1", len(album.Tracks))
	}
}

func TestAlbumGet_FetcherError(t *testing.T) {
	fetcher := &mockAlbumFetcher{
		fetchFn: func(albumID int64) (*models.AlbumFull, error) {
			return nil, errors.New("album not found")
		},
	}
	h := NewAlbumHandlers(nil, fetcher)

	r := chi.NewRouter()
	r.Get("/api/albums/{albumId}", h.Get)

	req := httptest.NewRequest("GET", "/api/albums/9999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
