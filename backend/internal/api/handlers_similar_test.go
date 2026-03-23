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

type mockSimilarStore struct {
	bands []models.SimilarBand
	err   error
}

func (m *mockSimilarStore) GetSimilarBands(maID int64, limit int) ([]models.SimilarBand, error) {
	return m.bands, m.err
}

func newSimilarRouter(store SimilarBandStore) *chi.Mux {
	r := chi.NewRouter()
	h := NewSimilarHandlers(store)
	r.Get("/api/bands/{maId}/similar", h.Get)
	return r
}

func TestSimilarHandlers_InvalidID(t *testing.T) {
	r := newSimilarRouter(&mockSimilarStore{})

	tests := []struct {
		name string
		path string
	}{
		{"non-numeric", "/api/bands/abc/similar"},
		{"zero", "/api/bands/0/similar"},
		{"negative", "/api/bands/-1/similar"},
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
			if body["error"] != "invalid band ID" {
				t.Errorf("expected 'invalid band ID', got %q", body["error"])
			}
		})
	}
}

func TestSimilarHandlers_Success(t *testing.T) {
	store := &mockSimilarStore{
		bands: []models.SimilarBand{
			{MAID: 100, Name: "Dark Funeral", Genre: "Black Metal", Country: "Sweden", Score: 0.95},
			{MAID: 200, Name: "Marduk", Genre: "Black Metal", Country: "Sweden", Score: 0.88},
		},
	}
	r := newSimilarRouter(store)

	req := httptest.NewRequest("GET", "/api/bands/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var bands []models.SimilarBand
	if err := json.NewDecoder(w.Body).Decode(&bands); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(bands) != 2 {
		t.Fatalf("expected 2 bands, got %d", len(bands))
	}

	if bands[0].Name != "Dark Funeral" {
		t.Errorf("expected 'Dark Funeral', got %q", bands[0].Name)
	}
	if bands[1].Score != 0.88 {
		t.Errorf("expected score 0.88, got %f", bands[1].Score)
	}
}

func TestSimilarHandlers_EmptyEmbeddings(t *testing.T) {
	store := &mockSimilarStore{
		bands: nil,
		err:   nil,
	}
	r := newSimilarRouter(store)

	req := httptest.NewRequest("GET", "/api/bands/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var bands []models.SimilarBand
	if err := json.NewDecoder(w.Body).Decode(&bands); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(bands) != 0 {
		t.Errorf("expected empty array, got %d bands", len(bands))
	}
}

func TestSimilarHandlers_StoreError(t *testing.T) {
	store := &mockSimilarStore{
		err: errors.New("db connection failed"),
	}
	r := newSimilarRouter(store)

	req := httptest.NewRequest("GET", "/api/bands/42/similar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
