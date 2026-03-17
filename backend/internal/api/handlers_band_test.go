package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
)

// mockBandFetcher implements BandFetcher for testing.
type mockBandFetcher struct {
	searchFn func(query string) ([]models.BandSearchResult, error)
	fetchFn  func(maID int64) (*models.BandFull, error)
}

func (m *mockBandFetcher) SearchBands(query string) ([]models.BandSearchResult, error) {
	return m.searchFn(query)
}

func (m *mockBandFetcher) FetchBand(maID int64) (*models.BandFull, error) {
	return m.fetchFn(maID)
}

func TestBandSearch_EmptyQuery(t *testing.T) {
	h := NewBandHandlers(nil, &mockBandFetcher{})

	req := httptest.NewRequest("GET", "/api/bands/search", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBandSearch_QueryTooLong(t *testing.T) {
	h := NewBandHandlers(nil, &mockBandFetcher{})

	longQuery := make([]byte, 201)
	for i := range longQuery {
		longQuery[i] = 'a'
	}
	req := httptest.NewRequest("GET", "/api/bands/search?q="+string(longQuery), nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBandSearch_Success(t *testing.T) {
	fetcher := &mockBandFetcher{
		searchFn: func(query string) ([]models.BandSearchResult, error) {
			return []models.BandSearchResult{
				{MAID: 125, Name: "Metallica", Genre: "Thrash Metal", Country: "United States"},
			}, nil
		},
	}
	h := NewBandHandlers(nil, fetcher)

	req := httptest.NewRequest("GET", "/api/bands/search?q=metallica", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var results []models.BandSearchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Metallica" {
		t.Errorf("unexpected results: %+v", results)
	}
}

func TestBandSearch_FetcherError(t *testing.T) {
	fetcher := &mockBandFetcher{
		searchFn: func(query string) ([]models.BandSearchResult, error) {
			return nil, errors.New("network error")
		},
	}
	h := NewBandHandlers(nil, fetcher)

	req := httptest.NewRequest("GET", "/api/bands/search?q=test", nil)
	w := httptest.NewRecorder()
	h.Search(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBandGet_InvalidID(t *testing.T) {
	h := NewBandHandlers(nil, &mockBandFetcher{})

	tests := []struct {
		name string
		id   string
	}{
		{"not a number", "abc"},
		{"zero", "0"},
		{"negative", "-1"},
		{"float", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/api/bands/{maId}", h.Get)

			req := httptest.NewRequest("GET", "/api/bands/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 400 {
				t.Errorf("status = %d, want 400", w.Code)
			}
		})
	}
}

func TestBandGet_Success(t *testing.T) {
	fetcher := &mockBandFetcher{
		fetchFn: func(maID int64) (*models.BandFull, error) {
			return &models.BandFull{
				Band: models.Band{MAID: maID, Name: "Opeth", Genre: "Progressive Metal"},
			}, nil
		},
	}
	h := NewBandHandlers(nil, fetcher)

	r := chi.NewRouter()
	r.Get("/api/bands/{maId}", h.Get)

	req := httptest.NewRequest("GET", "/api/bands/482", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var band models.BandFull
	if err := json.NewDecoder(w.Body).Decode(&band); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if band.Name != "Opeth" {
		t.Errorf("band name = %q, want Opeth", band.Name)
	}
}

func TestBandGet_FetcherError(t *testing.T) {
	fetcher := &mockBandFetcher{
		fetchFn: func(maID int64) (*models.BandFull, error) {
			return nil, errors.New("scrape failed")
		},
	}
	h := NewBandHandlers(nil, fetcher)

	r := chi.NewRouter()
	r.Get("/api/bands/{maId}", h.Get)

	req := httptest.NewRequest("GET", "/api/bands/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
