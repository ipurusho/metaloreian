package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/sync/singleflight"
)

const (
	baseURL        = "https://www.metal-archives.com"
	requestTimeout = 60 * time.Second
	cfWaitTime     = 8 * time.Second  // time to wait for Cloudflare challenge
	pageLoadWait   = 3 * time.Second  // time to wait after page load
)

type Client struct {
	browser *rod.Browser
	limiter *RateLimiter
	group   singleflight.Group
	mu      sync.Mutex
}

func NewClient() *Client {
	return &Client{
		limiter: NewRateLimiter(),
	}
}

// ensureBrowser lazily launches the headless browser.
func (c *Client) ensureBrowser() (*rod.Browser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browser != nil {
		return c.browser, nil
	}

	log.Println("launching headless browser...")
	path, _ := launcher.LookPath()
	if path == "" {
		log.Println("no browser found, downloading chromium...")
		path, _ = launcher.NewBrowser().Get()
	}

	u := launcher.New().
		Bin(path).
		Headless(true).
		Set("disable-gpu").
		Set("no-sandbox").
		Set("disable-dev-shm-usage").
		Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36").
		MustLaunch()

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("browser connect: %w", err)
	}

	c.browser = browser
	log.Println("headless browser ready")
	return c.browser, nil
}

// fetchHTML navigates to a URL with the headless browser, waits for
// Cloudflare to resolve, and returns the page HTML.
func (c *Client) fetchHTML(ctx context.Context, url string) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter: %w", err)
	}

	browser, err := c.ensureBrowser()
	if err != nil {
		return "", err
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}
	defer page.Close()

	// Set a reasonable timeout
	page = page.Timeout(requestTimeout)

	// Navigate
	if err := page.Navigate(url); err != nil {
		return "", fmt.Errorf("navigate to %s: %w", url, err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("wait load: %w", err)
	}

	// Wait for Cloudflare challenge to resolve — check for actual content
	deadline := time.Now().Add(cfWaitTime)
	for time.Now().Before(deadline) {
		html, err := page.HTML()
		if err != nil {
			return "", fmt.Errorf("get html: %w", err)
		}
		// Cloudflare challenge pages contain "Just a moment" or "challenge-platform"
		if !strings.Contains(html, "Just a moment") && !strings.Contains(html, "challenge-platform") {
			// Give the page a moment to fully render
			time.Sleep(500 * time.Millisecond)
			return page.HTML()
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Return whatever we have
	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("get html after cf wait: %w", err)
	}

	if strings.Contains(html, "Just a moment") {
		return "", fmt.Errorf("cloudflare challenge not resolved for %s", url)
	}

	return html, nil
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

// fetchJSON fetches a URL that returns JSON (like MA's AJAX endpoints).
// The headless browser handles it since the response is rendered into the page.
func (c *Client) fetchJSON(ctx context.Context, url string) (string, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter: %w", err)
	}

	browser, err := c.ensureBrowser()
	if err != nil {
		return "", err
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}
	defer page.Close()

	page = page.Timeout(requestTimeout)

	if err := page.Navigate(url); err != nil {
		return "", fmt.Errorf("navigate: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("wait load: %w", err)
	}

	// Wait for Cloudflare
	deadline := time.Now().Add(cfWaitTime)
	for time.Now().Before(deadline) {
		html, _ := page.HTML()
		if !strings.Contains(html, "Just a moment") && !strings.Contains(html, "challenge-platform") {
			time.Sleep(500 * time.Millisecond)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// For JSON responses, the browser wraps them in <pre> tags
	el, err := page.Element("pre")
	if err == nil {
		return el.Text()
	}

	// Fallback: get body text
	body, err := page.Element("body")
	if err != nil {
		return "", fmt.Errorf("no body element")
	}
	return body.Text()
}

// Dedup wraps a function call with singleflight deduplication.
func (c *Client) Dedup(key string, fn func() (any, error)) (any, error) {
	v, err, _ := c.group.Do(key, fn)
	return v, err
}

// Close shuts down the headless browser.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.browser != nil {
		c.browser.Close()
		c.browser = nil
	}
}
