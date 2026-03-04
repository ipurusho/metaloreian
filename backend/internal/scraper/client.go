package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/singleflight"
)

const (
	baseURL        = "https://www.metal-archives.com"
	requestTimeout = 30 * time.Second
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

type Client struct {
	http            *http.Client
	limiter         *RateLimiter
	group           singleflight.Group
	flareSolverrURL string
}

func NewClient(flareSolverrURL string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: requestTimeout,
		},
		limiter:         NewRateLimiter(),
		flareSolverrURL: flareSolverrURL,
	}
}

// fetchHTML fetches a URL via plain HTTP. If Cloudflare blocks the request
// and FlareSolverr is configured, it falls back to FlareSolverr.
func (c *Client) fetchHTML(ctx context.Context, url string) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter: %w", err)
	}

	// Try plain HTTP first.
	html, err := c.doHTTP(ctx, url)
	if err == nil {
		return html, nil
	}

	// If not a Cloudflare block or no FlareSolverr configured, give up.
	if !isCFBlock(err) || c.flareSolverrURL == "" {
		return "", err
	}

	log.Printf("cloudflare challenge for %s, falling back to FlareSolverr", url)
	return c.fetchViaFlareSolverr(ctx, url)
}

// doHTTP performs a plain HTTP GET.
func (c *Client) doHTTP(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	html := string(body)
	if strings.Contains(html, "Just a moment") || strings.Contains(html, "challenge-platform") {
		return "", &cfBlockError{url: url}
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}

	return html, nil
}

// cfBlockError signals a Cloudflare challenge.
type cfBlockError struct{ url string }

func (e *cfBlockError) Error() string {
	return fmt.Sprintf("cloudflare challenge for %s", e.url)
}

func isCFBlock(err error) bool {
	_, ok := err.(*cfBlockError)
	return ok
}

// fetchViaFlareSolverr uses FlareSolverr to fetch a URL through a real
// browser, bypassing Cloudflare challenges.
func (c *Client) fetchViaFlareSolverr(ctx context.Context, url string) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"cmd":        "request.get",
		"url":        url,
		"maxTimeout": 60000,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.flareSolverrURL+"/v1", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	fsClient := &http.Client{Timeout: 90 * time.Second}
	resp, err := fsClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("flaresolverr request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status   string `json:"status"`
		Message  string `json:"message"`
		Solution struct {
			Response string `json:"response"`
		} `json:"solution"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode flaresolverr response: %w", err)
	}

	if result.Status != "ok" {
		return "", fmt.Errorf("flaresolverr error: %s", result.Message)
	}

	return result.Solution.Response, nil
}

// FetchHTMLPublic is a public wrapper for fetchHTML (for debugging).
func (c *Client) FetchHTMLPublic(ctx context.Context, url string) (string, error) {
	return c.fetchHTML(ctx, url)
}

// fetchDoc fetches a URL and parses it as an HTML document.
func (c *Client) fetchDoc(ctx context.Context, url string) (*goquery.Document, error) {
	html, err := c.fetchHTML(ctx, url)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(strings.NewReader(html))
}

// fetchJSON fetches a URL that returns JSON.
func (c *Client) fetchJSON(ctx context.Context, url string) (string, error) {
	html, err := c.fetchHTML(ctx, url)
	if err != nil {
		return "", err
	}

	// FlareSolverr wraps JSON responses in <pre> tags.
	// Extract the content if present.
	if idx := strings.Index(html, "<pre>"); idx >= 0 {
		start := idx + 5
		if end := strings.Index(html[start:], "</pre>"); end >= 0 {
			return strings.NewReplacer(
				"&lt;", "<",
				"&gt;", ">",
				"&amp;", "&",
				"&quot;", `"`,
			).Replace(html[start : start+end]), nil
		}
	}

	return html, nil
}

// Dedup wraps a function call with singleflight deduplication.
func (c *Client) Dedup(key string, fn func() (any, error)) (any, error) {
	v, err, _ := c.group.Do(key, fn)
	return v, err
}

// Close is a no-op retained for interface compatibility.
func (c *Client) Close() {}
