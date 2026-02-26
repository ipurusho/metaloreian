package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/imman/metaloreian/internal/models"
	"github.com/imman/metaloreian/internal/store"
)

type AlbumHandlers struct {
	store   *store.Store
	fetcher AlbumFetcher
}

type AlbumFetcher interface {
	FetchAlbum(albumID int64) (*models.AlbumFull, error)
}

func NewAlbumHandlers(s *store.Store, f AlbumFetcher) *AlbumHandlers {
	return &AlbumHandlers{store: s, fetcher: f}
}

func (h *AlbumHandlers) Get(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "albumId")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil || albumID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid album ID")
		return
	}

	album, err := h.fetcher.FetchAlbum(albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch album")
		return
	}

	writeJSON(w, http.StatusOK, album)
}
