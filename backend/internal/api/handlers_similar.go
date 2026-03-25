package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
)

// SimilarBandStore defines the interface for fetching similar bands.
type SimilarBandStore interface {
	GetSimilarBands(maID int64, limit int) ([]models.SimilarBand, error)
}

type SimilarHandlers struct {
	store SimilarBandStore
}

func NewSimilarHandlers(s SimilarBandStore) *SimilarHandlers {
	return &SimilarHandlers{store: s}
}

func (h *SimilarHandlers) Get(w http.ResponseWriter, r *http.Request) {
	maIDStr := chi.URLParam(r, "maId")
	maID, err := strconv.ParseInt(maIDStr, 10, 64)
	if err != nil || maID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid band ID")
		return
	}

	bands, err := h.store.GetSimilarBands(maID, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch similar bands")
		return
	}

	if bands == nil {
		bands = []models.SimilarBand{}
	}

	writeJSON(w, http.StatusOK, bands)
}
