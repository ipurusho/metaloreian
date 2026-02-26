package scraper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imman/metaloreian/internal/models"
)

// ScrapeAlbum scrapes an album page from Metal Archives.
func (c *Client) ScrapeAlbum(ctx context.Context, albumID int64) (*models.AlbumFull, error) {
	url := fmt.Sprintf("%s/albums/_/_/%d", baseURL, albumID)
	doc, err := c.fetchDoc(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch album page: %w", err)
	}

	album := &models.AlbumFull{}
	album.AlbumID = albumID

	// Album name from h1
	album.Name = strings.TrimSpace(doc.Find("h1.album_name a").Text())

	// Band name and ID from band link
	bandLink := doc.Find("h2.band_name a")
	album.BandName = strings.TrimSpace(bandLink.Text())
	if href, exists := bandLink.Attr("href"); exists {
		if matches := bandIDRegex.FindStringSubmatch(href); len(matches) >= 2 {
			album.BandID, _ = strconv.ParseInt(matches[1], 10, 64)
		}
	}

	// Cover art
	if coverHref, exists := doc.Find(SelAlbumCover).Attr("href"); exists {
		album.CoverURL = coverHref
	}

	// Album info from dl/dt/dd pairs
	doc.Find(SelAlbumInfo).Find("dl").Each(func(_ int, dl *goquery.Selection) {
		dl.Find(SelStatLabel).Each(func(_ int, dt *goquery.Selection) {
			label := strings.TrimSpace(strings.TrimSuffix(dt.Text(), ":"))
			dd := dt.Next()
			value := strings.TrimSpace(dd.Text())

			switch strings.ToLower(label) {
			case "type":
				album.Type = value
			case "release date":
				album.ReleaseDate = value
			case "label":
				album.Label = value
			case "format":
				album.Format = value
			}
		})
	})

	// Tracklist
	album.Tracks = parseTracks(doc)

	// Album lineup
	album.Lineup = parseLineup(doc.Find(SelAlbumLineup))

	return album, nil
}

// parseTracks extracts tracks from the tracklist table.
func parseTracks(doc *goquery.Document) []models.Track {
	var tracks []models.Track
	trackNum := 0

	doc.Find(SelTracklist).Find(SelTrackRow).Each(func(_ int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() < 3 {
			return
		}

		trackNum++
		track := models.Track{
			TrackNumber: trackNum,
			Title:       strings.TrimSpace(tds.Eq(1).Text()),
			Duration:    strings.TrimSpace(tds.Eq(2).Text()),
		}

		tracks = append(tracks, track)
	})

	return tracks
}
