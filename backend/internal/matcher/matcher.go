package matcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/imman/metaloreian/internal/models"
	"github.com/imman/metaloreian/internal/scraper"
	"github.com/imman/metaloreian/internal/store"
	"golang.org/x/sync/singleflight"
)

// Matcher resolves Spotify artist names to Metal Archives band data.
// It acts as the orchestrator between the cache (store) and the scraper.
type Matcher struct {
	store   *store.Store
	scraper *scraper.Client
	group   singleflight.Group
}

func New(s *store.Store, sc *scraper.Client) *Matcher {
	return &Matcher{store: s, scraper: sc}
}

// SearchBands searches MA for bands, returning results.
func (m *Matcher) SearchBands(query string) ([]models.BandSearchResult, error) {
	// First check local DB cache
	cached, err := m.store.SearchBandsByName(query)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Fall through to MA search
	results, err := m.scraper.SearchBands(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("MA search: %w", err)
	}
	return results, nil
}

// FetchBand returns full band data, using cache-or-scrape pattern with singleflight dedup.
func (m *Matcher) FetchBand(maID int64) (*models.BandFull, error) {
	key := fmt.Sprintf("band:%d", maID)

	v, err, _ := m.group.Do(key, func() (any, error) {
		return m.fetchBandInner(maID)
	})
	if err != nil {
		return nil, err
	}
	return v.(*models.BandFull), nil
}

func (m *Matcher) fetchBandInner(maID int64) (*models.BandFull, error) {
	// Check cache
	cached, err := m.store.GetBand(maID)
	if err != nil {
		return nil, err
	}

	if cached != nil && m.store.IsBandFresh(cached) {
		// Serve from cache
		return m.assembleBandFromCache(cached)
	}

	// Scrape fresh data
	full, err := m.scraper.ScrapeBandFull(context.Background(), maID)
	if err != nil {
		// If we have stale cache, serve it
		if cached != nil {
			log.Printf("scrape failed for band %d, serving stale cache: %v", maID, err)
			return m.assembleBandFromCache(cached)
		}
		return nil, fmt.Errorf("scrape band %d: %w", maID, err)
	}

	// Persist to cache
	if err := m.persistBand(full); err != nil {
		log.Printf("failed to cache band %d: %v", maID, err)
	}

	return full, nil
}

func (m *Matcher) assembleBandFromCache(band *models.Band) (*models.BandFull, error) {
	full := &models.BandFull{Band: *band}

	var err error
	full.CurrentLineup, err = m.store.GetBandLineup(band.MAID, "current")
	if err != nil {
		return nil, err
	}

	full.PastLineup, err = m.store.GetBandLineup(band.MAID, "past")
	if err != nil {
		return nil, err
	}

	full.Discography, err = m.store.GetDiscography(band.MAID)
	if err != nil {
		return nil, err
	}

	return full, nil
}

func (m *Matcher) persistBand(full *models.BandFull) error {
	if err := m.store.UpsertBand(&full.Band); err != nil {
		return err
	}

	for _, member := range append(full.CurrentLineup, full.PastLineup...) {
		if err := m.store.UpsertMember(&member); err != nil {
			return err
		}

		lineupType := "current"
		if member.LineupType != "" {
			lineupType = member.LineupType
		}
		// Determine type from which list the member came from
		for _, cm := range full.CurrentLineup {
			if cm.MemberID == member.MemberID {
				lineupType = "current"
				break
			}
		}
		for _, pm := range full.PastLineup {
			if pm.MemberID == member.MemberID {
				lineupType = "past"
				break
			}
		}

		if err := m.store.UpsertBandLineup(full.Band.MAID, member.MemberID, member.Instrument, lineupType, member.Years); err != nil {
			return err
		}

		for _, ob := range member.OtherBands {
			if err := m.store.UpsertMemberBand(&ob); err != nil {
				return err
			}
		}
	}

	for _, album := range full.Discography {
		if err := m.store.UpsertAlbum(&album); err != nil {
			return err
		}
	}

	return nil
}

// FetchAlbum returns full album data with cache-or-scrape pattern.
func (m *Matcher) FetchAlbum(albumID int64) (*models.AlbumFull, error) {
	key := fmt.Sprintf("album:%d", albumID)

	v, err, _ := m.group.Do(key, func() (any, error) {
		return m.fetchAlbumInner(albumID)
	})
	if err != nil {
		return nil, err
	}
	return v.(*models.AlbumFull), nil
}

func (m *Matcher) fetchAlbumInner(albumID int64) (*models.AlbumFull, error) {
	// Check cache
	cached, err := m.store.GetAlbum(albumID)
	if err != nil {
		return nil, err
	}

	if cached != nil && m.store.IsAlbumFresh(cached) {
		return m.assembleAlbumFromCache(cached)
	}

	// Scrape fresh
	full, err := m.scraper.ScrapeAlbum(context.Background(), albumID)
	if err != nil {
		if cached != nil {
			log.Printf("scrape failed for album %d, serving stale cache: %v", albumID, err)
			return m.assembleAlbumFromCache(cached)
		}
		return nil, fmt.Errorf("scrape album %d: %w", albumID, err)
	}

	// Persist
	if err := m.persistAlbum(full); err != nil {
		log.Printf("failed to cache album %d: %v", albumID, err)
	}

	return full, nil
}

func (m *Matcher) assembleAlbumFromCache(album *models.Album) (*models.AlbumFull, error) {
	full := &models.AlbumFull{Album: *album}

	// Get band name
	band, err := m.store.GetBand(album.BandID)
	if err == nil && band != nil {
		full.BandName = band.Name
	}

	full.Tracks, err = m.store.GetAlbumTracks(album.AlbumID)
	if err != nil {
		return nil, err
	}

	full.Lineup, err = m.store.GetAlbumLineup(album.AlbumID)
	if err != nil {
		return nil, err
	}

	return full, nil
}

func (m *Matcher) persistAlbum(full *models.AlbumFull) error {
	if err := m.store.UpsertAlbum(&full.Album); err != nil {
		return err
	}

	if err := m.store.ReplaceTracks(full.AlbumID, full.Tracks); err != nil {
		return err
	}

	for _, member := range full.Lineup {
		if err := m.store.UpsertMember(&member); err != nil {
			return err
		}
		if err := m.store.UpsertAlbumLineup(full.AlbumID, member.MemberID, member.Instrument); err != nil {
			return err
		}
		for _, ob := range member.OtherBands {
			if err := m.store.UpsertMemberBand(&ob); err != nil {
				return err
			}
		}
	}

	return nil
}

// NormalizeName strips common prefixes/suffixes and normalizes for matching.
func NormalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "the ")

	// Remove diacritics and non-alphanumeric
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
