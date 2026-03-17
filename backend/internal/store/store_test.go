package store

import (
	"testing"
	"time"

	"github.com/imman/metaloreian/internal/models"
)

func TestIsBandFresh(t *testing.T) {
	s := &Store{} // DB not needed for TTL checks

	tests := []struct {
		name      string
		scrapedAt time.Time
		want      bool
	}{
		{"just scraped", time.Now(), true},
		{"1 day ago", time.Now().Add(-24 * time.Hour), true},
		{"6 days ago", time.Now().Add(-6 * 24 * time.Hour), true},
		{"7 days ago exactly", time.Now().Add(-7 * 24 * time.Hour), false},
		{"8 days ago", time.Now().Add(-8 * 24 * time.Hour), false},
		{"30 days ago", time.Now().Add(-30 * 24 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			band := &models.Band{ScrapedAt: tt.scrapedAt}
			if got := s.IsBandFresh(band); got != tt.want {
				t.Errorf("IsBandFresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAlbumFresh(t *testing.T) {
	s := &Store{}

	tests := []struct {
		name      string
		scrapedAt time.Time
		want      bool
	}{
		{"just scraped", time.Now(), true},
		{"3 days ago", time.Now().Add(-3 * 24 * time.Hour), true},
		{"8 days ago", time.Now().Add(-8 * 24 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			album := &models.Album{ScrapedAt: tt.scrapedAt}
			if got := s.IsAlbumFresh(album); got != tt.want {
				t.Errorf("IsAlbumFresh() = %v, want %v", got, tt.want)
			}
		})
	}
}
