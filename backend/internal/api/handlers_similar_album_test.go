package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
)

type mockSimilarAlbumStore struct {
	albums []models.SimilarAlbum
	err    error
}

func (m *mockSimilarAlbumStore) GetSimilarAlbums(albumID int64, limit int) ([]models.SimilarAlbum, error) {
	return m.albums, m.err
}

func newSimilarAlbumRouter(store SimilarAlbumStore) *chi.Mux {
	r := chi.NewRouter()
	h := NewSimilarAlbumHandlers(store)
	r.Get("/api/albums/{albumId}/similar", h.Get)
	return r
}

func TestSimilarAlbumHandlers_InvalidID(t *testing.T) {
	r := newSimilarAlbumRouter(&mockSimilarAlbumStore{})

	tests := []struct {
		name string
		path string
	}{
		{"non-numeric", "/api/albums/abc/similar"},
		{"zero", "/api/albums/0/similar"},
		{"negative", "/api/albums/-1/similar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}

			var body map[string]string
			json.NewDecoder(w.Body).Decode(&body)
			if body["error"] != "invalid album ID" {
				t.Errorf("expected 'invalid album ID', got %q", body["error"])
			}
		})
	}
}

func TestSimilarAlbumHandlers_Success(t *testing.T) {
	store := &mockSimilarAlbumStore{
		albums: []models.SimilarAlbum{
			{AlbumID: 100, Name: "Blackwater Park", BandName: "Opeth", Type: "Full-length", Year: "2001", CoverURL: "", Score: 0.95},
			{AlbumID: 200, Name: "Still Life", BandName: "Opeth", Type: "Full-length", Year: "1999", CoverURL: "", Score: 0.88},
		},
	}
	r := newSimilarAlbumRouter(store)

	req := httptest.NewRequest("GET", "/api/albums/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var albums []models.SimilarAlbum
	if err := json.NewDecoder(w.Body).Decode(&albums); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(albums) != 2 {
		t.Fatalf("expected 2 albums, got %d", len(albums))
	}

	if albums[0].Name != "Blackwater Park" {
		t.Errorf("expected 'Blackwater Park', got %q", albums[0].Name)
	}
	if albums[1].Score != 0.88 {
		t.Errorf("expected score 0.88, got %f", albums[1].Score)
	}
}

func TestSimilarAlbumHandlers_EmptyEmbeddings(t *testing.T) {
	store := &mockSimilarAlbumStore{
		albums: nil,
		err:    nil,
	}
	r := newSimilarAlbumRouter(store)

	req := httptest.NewRequest("GET", "/api/albums/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var albums []models.SimilarAlbum
	if err := json.NewDecoder(w.Body).Decode(&albums); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(albums) != 0 {
		t.Errorf("expected empty array, got %d albums", len(albums))
	}
}

func TestSimilarAlbumHandlers_StoreError(t *testing.T) {
	store := &mockSimilarAlbumStore{
		err: errors.New("db connection failed"),
	}
	r := newSimilarAlbumRouter(store)

	req := httptest.NewRequest("GET", "/api/albums/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
