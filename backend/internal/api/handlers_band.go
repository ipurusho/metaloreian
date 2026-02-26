package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
	"github.com/imman/metaloreian/internal/store"
)

type BandHandlers struct {
	store   *store.Store
	fetcher BandFetcher
}

type BandFetcher interface {
	SearchBands(query string) ([]models.BandSearchResult, error)
	FetchBand(maID int64) (*models.BandFull, error)
}

func NewBandHandlers(s *store.Store, f BandFetcher) *BandHandlers {
	return &BandHandlers{store: s, fetcher: f}
}

func (h *BandHandlers) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "q parameter required")
		return
	}
	if len(query) > 200 {
		writeError(w, http.StatusBadRequest, "query too long (max 200 characters)")
		return
	}

	results, err := h.fetcher.SearchBands(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *BandHandlers) Get(w http.ResponseWriter, r *http.Request) {
	maIDStr := chi.URLParam(r, "maId")
	maID, err := strconv.ParseInt(maIDStr, 10, 64)
	if err != nil || maID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid band ID")
		return
	}

	band, err := h.fetcher.FetchBand(maID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch band")
		return
	}

	writeJSON(w, http.StatusOK, band)
}
