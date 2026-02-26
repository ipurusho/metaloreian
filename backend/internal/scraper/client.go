package scraper

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/singleflight"
)

const (
	baseURL   = "https://www.metal-archives.com"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	maxRetries     = 3
	baseBackoff    = 5 * time.Second
	backoffFactor  = 3.0
	requestTimeout = 30 * time.Second
)

type Client struct {
	http    *http.Client
	limiter *RateLimiter
	group   singleflight.Group
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: requestTimeout,
		},
		limiter: NewRateLimiter(),
	}
}

// fetch performs a rate-limited GET request with retry/backoff.
func (c *Client) fetch(ctx context.Context, url string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(float64(baseBackoff) * math.Pow(backoffFactor, float64(attempt-1)))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == 429 || resp.StatusCode == 503 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// fetchDoc fetches a URL and parses it as an HTML document.
func (c *Client) fetchDoc(ctx context.Context, url string) (*goquery.Document, error) {
	resp, err := c.fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

// Dedup wraps a function call with singleflight deduplication.
func (c *Client) Dedup(key string, fn func() (any, error)) (any, error) {
	v, err, _ := c.group.Do(key, fn)
	return v, err
}
