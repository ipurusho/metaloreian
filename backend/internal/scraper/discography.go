package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imman/metaloreian/internal/models"
)

var albumIDRegex = regexp.MustCompile(`/albums/[^/]+/[^/]+/(\d+)`)

// ScrapeDiscography fetches and parses the discography AJAX fragment for a band.
func (c *Client) ScrapeDiscography(ctx context.Context, bandID int64) ([]models.Album, error) {
	url := fmt.Sprintf("%s/band/discography/id/%d/tab/all", baseURL, bandID)
	doc, err := c.fetchDoc(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch discography: %w", err)
	}

	var albums []models.Album

	doc.Find(SelDiscogTable).Find(SelDiscogRow).Each(func(_ int, row *goquery.Selection) {
		tds := row.Find("td")
		if tds.Length() < 4 {
			return
		}

		album := models.Album{BandID: bandID}

		// Album name and ID from link
		link := tds.Eq(0).Find("a").First()
		album.Name = strings.TrimSpace(link.Text())
		if href, exists := link.Attr("href"); exists {
			if matches := albumIDRegex.FindStringSubmatch(href); len(matches) >= 2 {
				album.AlbumID, _ = strconv.ParseInt(matches[1], 10, 64)
			}
		}

		// Type
		album.Type = strings.TrimSpace(tds.Eq(1).Text())

		// Release date
		album.ReleaseDate = strings.TrimSpace(tds.Eq(2).Text())

		if album.AlbumID > 0 {
			albums = append(albums, album)
		}
	})

	return albums, nil
}
