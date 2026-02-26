package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/imman/metaloreian/internal/models"
)

// maSearchResponse represents the JSON response from MA's AJAX band search.
type maSearchResponse struct {
	// Each entry: [bandLink HTML, genre, country]
	AAData [][]string `json:"aaData"`
}

var bandIDRegex = regexp.MustCompile(`/bands/[^/]+/(\d+)`)

// SearchBands searches Metal Archives for bands matching the query.
func (c *Client) SearchBands(ctx context.Context, query string) ([]models.BandSearchResult, error) {
	u := fmt.Sprintf("%s/search/ajax-band-search/?field=name&query=%s&sEcho=1&iDisplayStart=0&iDisplayLength=50",
		baseURL, url.QueryEscape(query))

	resp, err := c.fetch(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var sr maSearchResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	var results []models.BandSearchResult
	for _, row := range sr.AAData {
		if len(row) < 3 {
			continue
		}

		// Extract band ID and name from the HTML link in row[0]
		matches := bandIDRegex.FindStringSubmatch(row[0])
		if len(matches) < 2 {
			continue
		}
		maID, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			continue
		}

		// Extract band name from the link text
		name := extractLinkText(row[0])

		results = append(results, models.BandSearchResult{
			MAID:    maID,
			Name:    name,
			Genre:   strings.TrimSpace(row[1]),
			Country: strings.TrimSpace(row[2]),
		})
	}

	return results, nil
}

// extractLinkText extracts text content from an HTML anchor tag.
func extractLinkText(html string) string {
	start := strings.Index(html, ">")
	if start == -1 {
		return html
	}
	end := strings.Index(html[start:], "<")
	if end == -1 {
		return html[start+1:]
	}
	return strings.TrimSpace(html[start+1 : start+end])
}
