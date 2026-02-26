package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/imman/metaloreian/internal/config"
)

type SpotifyHandlers struct {
	cfg *config.Config
}

func NewSpotifyHandlers(cfg *config.Config) *SpotifyHandlers {
	return &SpotifyHandlers{cfg: cfg}
}

type tokenRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

const maxBodySize = 4 * 1024 // 4KB — token requests are small

func (h *SpotifyHandlers) Exchange(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {req.Code},
		"redirect_uri":  {req.RedirectURI},
		"client_id":     {h.cfg.SpotifyClientID},
		"code_verifier": {req.CodeVerifier},
	}

	resp, err := http.Post(
		"https://accounts.spotify.com/api/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to contact Spotify")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, io.LimitReader(resp.Body, 64*1024))
}

func (h *SpotifyHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {req.RefreshToken},
		"client_id":     {h.cfg.SpotifyClientID},
	}

	resp, err := http.Post(
		"https://accounts.spotify.com/api/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to contact Spotify")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, io.LimitReader(resp.Body, 64*1024))
}
