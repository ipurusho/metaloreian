package api

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/imman/metaloreian/internal/config"
)

func TestSpotifyExchange_InvalidBody(t *testing.T) {
	h := NewSpotifyHandlers(&config.Config{SpotifyClientID: "test-id"})

	req := httptest.NewRequest("POST", "/api/spotify/exchange", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.Exchange(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSpotifyExchange_EmptyBody(t *testing.T) {
	h := NewSpotifyHandlers(&config.Config{SpotifyClientID: "test-id"})

	req := httptest.NewRequest("POST", "/api/spotify/exchange", nil)
	w := httptest.NewRecorder()
	h.Exchange(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSpotifyExchange_ValidBody(t *testing.T) {
	// The Exchange/Refresh handlers proxy directly to Spotify's token endpoint
	// (hardcoded URL). Full proxy testing requires making the URL injectable.
	// Tracked in issue #19 (integration test expansion).
	t.Skip("Spotify handlers need URL injection for proper proxy testing")
}

func TestSpotifyRefresh_InvalidBody(t *testing.T) {
	h := NewSpotifyHandlers(&config.Config{SpotifyClientID: "test-id"})

	req := httptest.NewRequest("POST", "/api/spotify/refresh", bytes.NewReader([]byte("{bad")))
	w := httptest.NewRecorder()
	h.Refresh(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSpotifyRefresh_EmptyBody(t *testing.T) {
	h := NewSpotifyHandlers(&config.Config{SpotifyClientID: "test-id"})

	req := httptest.NewRequest("POST", "/api/spotify/refresh", nil)
	w := httptest.NewRecorder()
	h.Refresh(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
