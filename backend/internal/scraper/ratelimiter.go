package scraper

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter wraps a token bucket rate limiter.
// Configured for 1 request per 3 seconds to avoid Metal Archives blocking.
type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(3*time.Second), 1),
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}
