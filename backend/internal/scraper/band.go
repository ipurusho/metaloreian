package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imman/metaloreian/internal/models"
)

// ScrapeBand scrapes a band's main page from Metal Archives.
func (c *Client) ScrapeBand(ctx context.Context, maID int64) (*models.Band, error) {
	url := fmt.Sprintf("%s/bands/_/%d", baseURL, maID)
	doc, err := c.fetchDoc(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch band page: %w", err)
	}

	band := &models.Band{MAID: maID}

	// Band name
	band.Name = strings.TrimSpace(doc.Find(SelBandName).Text())

	// Logo and photo URLs
	band.LogoURL, _ = doc.Find(SelBandLogo).Attr("href")
	band.PhotoURL, _ = doc.Find(SelBandPhoto).Attr("href")

	// Band stats from dl/dt/dd pairs
	doc.Find(SelBandStats).Find("dl").Each(func(_ int, dl *goquery.Selection) {
		dl.Find(SelStatLabel).Each(func(i int, dt *goquery.Selection) {
			label := strings.TrimSpace(strings.TrimSuffix(dt.Text(), ":"))
			dd := dt.Next()
			value := strings.TrimSpace(dd.Text())

			switch strings.ToLower(label) {
			case "country of origin":
				band.Country = value
			case "location":
				// skip
			case "status":
				band.Status = value
			case "formed in":
				band.FormedIn = value
			case "genre":
				band.Genre = value
			case "lyrical themes":
				band.Themes = value
			case "years active":
				band.YearsActive = value
			}
		})
	})

	return band, nil
}

// ScrapeBandFull scrapes band page + lineup + discography.
func (c *Client) ScrapeBandFull(ctx context.Context, maID int64) (*models.BandFull, error) {
	band, err := c.ScrapeBand(ctx, maID)
	if err != nil {
		return nil, err
	}

	full := &models.BandFull{Band: *band}

	// Scrape lineup
	url := fmt.Sprintf("%s/bands/_/%d", baseURL, maID)
	doc, err := c.fetchDoc(ctx, url)
	if err != nil {
		return nil, err
	}

	full.CurrentLineup = parseLineup(doc.Find(SelCurrentLineup))
	full.PastLineup = parseLineup(doc.Find(SelPastLineup))

	// Scrape discography
	discog, err := c.ScrapeDiscography(ctx, maID)
	if err != nil {
		return nil, err
	}
	full.Discography = discog

	return full, nil
}
