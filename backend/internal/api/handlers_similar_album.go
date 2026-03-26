package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
)

// SimilarAlbumStore defines the interface for fetching similar albums.
type SimilarAlbumStore interface {
	GetSimilarAlbums(albumID int64, limit int) ([]models.SimilarAlbum, error)
}

type SimilarAlbumHandlers struct {
	store SimilarAlbumStore
}

func NewSimilarAlbumHandlers(s SimilarAlbumStore) *SimilarAlbumHandlers {
	return &SimilarAlbumHandlers{store: s}
}

func (h *SimilarAlbumHandlers) Get(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "albumId")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil || albumID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid album ID")
		return
	}

	albums, err := h.store.GetSimilarAlbums(albumID, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch similar albums")
		return
	}

	if albums == nil {
		albums = []models.SimilarAlbum{}
	}

	writeJSON(w, http.StatusOK, albums)
}
