package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":       "DENY",
		"Referrer-Policy":       "strict-origin-when-cross-origin",
	}

	for header, want := range expected {
		if got := w.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestLoggingMiddleware(t *testing.T) {
	called := false
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("inner handler was not called")
	}
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRateLimitMiddleware_AllowsBurst(t *testing.T) {
	limiter := RateLimitMiddleware(60) // 60 req/min = 1/sec, burst of 60
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request should pass
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("first request: status = %d, want 200", w.Code)
	}
}

func TestRateLimitMiddleware_BlocksWhenExhausted(t *testing.T) {
	limiter := RateLimitMiddleware(2) // tiny burst of 2
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the bucket
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < 2 && w.Code != 200 {
			t.Errorf("request %d: status = %d, want 200", i, w.Code)
		}
		if i == 2 && w.Code != 429 {
			t.Errorf("request %d: status = %d, want 429", i, w.Code)
		}
	}
}

func TestRateLimitMiddleware_UsesXRealIP(t *testing.T) {
	limiter := RateLimitMiddleware(1) // burst of 1
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with X-Real-IP
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "1.2.3.4")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("first request: status = %d, want 200", w.Code)
	}

	// Second request same IP — should be blocked
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Real-IP", "1.2.3.4")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != 429 {
		t.Errorf("second request: status = %d, want 429", w2.Code)
	}

	// Different IP should pass
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("X-Real-IP", "5.6.7.8")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	if w3.Code != 200 {
		t.Errorf("different IP: status = %d, want 200", w3.Code)
	}
}

func TestStatusWriter(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: 200}

	sw.WriteHeader(http.StatusNotFound)
	if sw.status != 404 {
		t.Errorf("status = %d, want 404", sw.status)
	}
	if w.Code != 404 {
		t.Errorf("underlying status = %d, want 404", w.Code)
	}
}
