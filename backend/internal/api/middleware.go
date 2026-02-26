package api

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.status, time.Since(start))
	})
}

// SecurityHeaders adds standard security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides a simple per-IP token bucket rate limiter for API routes.
func RateLimitMiddleware(requestsPerMinute int) func(http.Handler) http.Handler {
	type visitor struct {
		tokens    float64
		lastSeen  time.Time
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
		rate     = float64(requestsPerMinute) / 60.0 // tokens per second
		burst    = float64(requestsPerMinute)
	)

	// Background cleanup of stale entries
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				v = &visitor{tokens: burst, lastSeen: time.Now()}
				visitors[ip] = v
			}

			// Refill tokens based on elapsed time
			elapsed := time.Since(v.lastSeen).Seconds()
			v.tokens += elapsed * rate
			if v.tokens > burst {
				v.tokens = burst
			}
			v.lastSeen = time.Now()

			if v.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			v.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
