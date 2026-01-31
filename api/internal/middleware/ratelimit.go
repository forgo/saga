package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // Requests per window
	window   time.Duration // Time window
	burst    int           // Max burst size
	cleanup  time.Duration // Cleanup interval for expired buckets
	stopChan chan struct{}
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// RateLimitConfig holds rate limiter configuration
type RateLimitConfig struct {
	Rate    int           // Requests per window (default 100)
	Window  time.Duration // Time window (default 1 minute)
	Burst   int           // Max burst (default 20)
	Cleanup time.Duration // Cleanup interval (default 5 minutes)
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.Rate == 0 {
		cfg.Rate = 100
	}
	if cfg.Window == 0 {
		cfg.Window = time.Minute
	}
	if cfg.Burst == 0 {
		cfg.Burst = 20
	}
	if cfg.Cleanup == 0 {
		cfg.Cleanup = 5 * time.Minute
	}

	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     cfg.Rate,
		window:   cfg.Window,
		burst:    cfg.Burst,
		cleanup:  cfg.Cleanup,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupExpired()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *RateLimiter) cleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.window * 2)
	for key, b := range rl.buckets {
		if b.lastReset.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) (allowed bool, remaining int, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]

	if !exists {
		// New bucket with burst tokens
		b = &bucket{
			tokens:    rl.rate + rl.burst - 1, // -1 for this request
			lastReset: now,
		}
		rl.buckets[key] = b
		return true, b.tokens, now.Add(rl.window)
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastReset)
	if elapsed >= rl.window {
		// Full refill
		b.tokens = rl.rate + rl.burst
		b.lastReset = now
	} else {
		// Partial refill based on time elapsed
		tokensToAdd := int(float64(rl.rate) * (float64(elapsed) / float64(rl.window)))
		b.tokens += tokensToAdd
		if b.tokens > rl.rate+rl.burst {
			b.tokens = rl.rate + rl.burst
		}
		if tokensToAdd > 0 {
			b.lastReset = now
		}
	}

	// Check if request is allowed
	if b.tokens > 0 {
		b.tokens--
		return true, b.tokens, b.lastReset.Add(rl.window)
	}

	return false, 0, b.lastReset.Add(rl.window)
}

// RateLimit returns a middleware that applies rate limiting
func RateLimit(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get rate limit key (user ID if authenticated, otherwise IP)
			key := GetUserID(r.Context())
			if key == "" {
				key = r.RemoteAddr
			}

			allowed, remaining, resetTime := limiter.Allow(key)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.rate))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				retryAfter := int(time.Until(resetTime).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				model.NewRateLimitError(retryAfter).WriteJSON(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
