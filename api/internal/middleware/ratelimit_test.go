package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// NewRateLimiter Tests (Configuration)
// ============================================================================

func TestNewRateLimiter_DefaultConfig(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{})
	defer rl.Stop()

	if rl.rate != 100 {
		t.Errorf("expected default rate 100, got %d", rl.rate)
	}
	if rl.window != time.Minute {
		t.Errorf("expected default window 1m, got %v", rl.window)
	}
	if rl.burst != 20 {
		t.Errorf("expected default burst 20, got %d", rl.burst)
	}
}

func TestNewRateLimiter_CustomConfig(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   50,
		Window: 30 * time.Second,
		Burst:  10,
	})
	defer rl.Stop()

	if rl.rate != 50 {
		t.Errorf("expected rate 50, got %d", rl.rate)
	}
	if rl.window != 30*time.Second {
		t.Errorf("expected window 30s, got %v", rl.window)
	}
	if rl.burst != 10 {
		t.Errorf("expected burst 10, got %d", rl.burst)
	}
}

// ============================================================================
// Allow() Tests
// ============================================================================

func TestAllow_FirstRequest_AllowsAndCreatesNewBucket(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   10,
		Window: time.Minute,
		Burst:  5,
	})
	defer rl.Stop()

	allowed, remaining, _ := rl.Allow("user:123")

	if !allowed {
		t.Error("first request should be allowed")
	}
	// New bucket starts with rate + burst - 1 (minus this request)
	// So: 10 + 5 - 1 = 14
	if remaining != 14 {
		t.Errorf("expected remaining 14, got %d", remaining)
	}
}

func TestAllow_MultipleRequests_DecrementsTokens(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   10,
		Window: time.Minute,
		Burst:  5,
	})
	defer rl.Stop()

	// Make 5 requests
	for i := 0; i < 5; i++ {
		allowed, _, _ := rl.Allow("user:123")
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// Check remaining tokens
	_, remaining, _ := rl.Allow("user:123")
	// First request: bucket created with rate+burst-1 = 14, returns 14
	// Requests 2-5: deduct 1 each, bucket goes 14->13->12->11->10
	// Request 6 (this one): 10->9, returns 9
	if remaining != 9 {
		t.Errorf("expected remaining 9, got %d", remaining)
	}
}

func TestAllow_ExceedsLimit_Denies(t *testing.T) {
	t.Parallel()
	// Note: Burst=0 triggers default of 20, so we use a small explicit value
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   5,
		Window: time.Minute,
		Burst:  1, // Small burst
	})
	defer rl.Stop()

	// With rate=5 and burst=1, first request creates bucket with rate+burst-1 = 5 tokens
	// We can make 6 total requests (5+1) before being denied
	var allowed bool
	for i := 0; i < 6; i++ {
		allowed, _, _ = rl.Allow("user:123")
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 7th request should be denied
	allowed, remaining, _ := rl.Allow("user:123")

	if allowed {
		t.Error("7th request should be denied after limit exceeded")
	}
	if remaining != 0 {
		t.Errorf("expected remaining 0, got %d", remaining)
	}
}

func TestAllow_DifferentKeys_SeparateBuckets(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   5,
		Window: time.Minute,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	// Exhaust tokens for user:123 (rate+burst = 6 requests)
	for i := 0; i < 6; i++ {
		rl.Allow("user:123")
	}

	// Verify user:123 is denied
	allowed, _, _ := rl.Allow("user:123")
	if allowed {
		t.Error("user:123 should be denied")
	}

	// user:456 should still have tokens (new bucket: rate+burst-1 = 5)
	allowed, remaining, _ := rl.Allow("user:456")

	if !allowed {
		t.Error("different user should have separate bucket")
	}
	// First request creates bucket with 5+1-1 = 5 tokens remaining
	if remaining != 5 {
		t.Errorf("expected remaining 5, got %d", remaining)
	}
}

func TestAllow_FullRefill_AfterWindow(t *testing.T) {
	t.Parallel()
	// Use short window for testing
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   5,
		Window: 50 * time.Millisecond,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	// Exhaust all tokens (rate+burst = 6)
	for i := 0; i < 6; i++ {
		rl.Allow("user:123")
	}

	// Should be denied
	allowed, _, _ := rl.Allow("user:123")
	if allowed {
		t.Error("should be denied when exhausted")
	}

	// Wait for full window
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again with full refill
	allowed, remaining, _ := rl.Allow("user:123")

	if !allowed {
		t.Error("should be allowed after full refill")
	}
	// Full refill gives rate + burst = 6, then -1 for this request = 5
	if remaining != 5 {
		t.Errorf("expected remaining 5 after refill, got %d", remaining)
	}
}

func TestAllow_PartialRefill_BeforeWindow(t *testing.T) {
	t.Parallel()
	// Rate of 100 per 100ms means 1 token per 1ms
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   100,
		Window: 100 * time.Millisecond,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	// Use some tokens
	for i := 0; i < 50; i++ {
		rl.Allow("user:123")
	}

	// Wait for partial window (should add some tokens back)
	time.Sleep(30 * time.Millisecond)

	// Make another request - should have partial refill
	allowed, remaining, _ := rl.Allow("user:123")

	if !allowed {
		t.Error("should be allowed with partial refill")
	}
	// Hard to predict exact tokens due to timing, but should be positive
	if remaining < 0 {
		t.Errorf("remaining should be positive, got %d", remaining)
	}
}

func TestAllow_BurstAllowsExtra(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   5,
		Window: time.Minute,
		Burst:  3,
	})
	defer rl.Stop()

	// Should allow rate + burst requests
	allowed := true
	var count int
	for i := 0; i < 10 && allowed; i++ {
		allowed, _, _ = rl.Allow("user:123")
		if allowed {
			count++
		}
	}

	// Should have allowed 5 + 3 = 8 requests
	if count != 8 {
		t.Errorf("expected 8 allowed requests with burst, got %d", count)
	}
}

func TestAllow_TokensCapped_AtMax(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   10,
		Window: 50 * time.Millisecond,
		Burst:  5,
	})
	defer rl.Stop()

	// Make one request to create bucket
	rl.Allow("user:123")

	// Wait for multiple windows (should not exceed rate + burst)
	time.Sleep(200 * time.Millisecond)

	// Tokens should be capped at rate + burst
	_, remaining, _ := rl.Allow("user:123")

	// Max is rate + burst - 1 (for this request) = 10 + 5 - 1 = 14
	if remaining > 14 {
		t.Errorf("remaining should be capped at 14, got %d", remaining)
	}
}

func TestAllow_ReturnsResetTime(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   10,
		Window: time.Minute,
		Burst:  0,
	})
	defer rl.Stop()

	before := time.Now()
	_, _, resetTime := rl.Allow("user:123")
	after := time.Now()

	// Reset time should be approximately now + window
	expectedReset := before.Add(time.Minute)
	if resetTime.Before(expectedReset.Add(-time.Second)) || resetTime.After(after.Add(time.Minute).Add(time.Second)) {
		t.Errorf("reset time %v not in expected range around %v", resetTime, expectedReset)
	}
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestAllow_ConcurrentAccess_ThreadSafe(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   1000,
		Window: time.Minute,
		Burst:  100,
	})
	defer rl.Stop()

	var wg sync.WaitGroup
	workers := 10
	requestsPerWorker := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				rl.Allow("shared-key")
			}
		}(i)
	}

	wg.Wait()
	// If no race condition, test passes
}

func TestAllow_ConcurrentAccess_DifferentKeys_ThreadSafe(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   100,
		Window: time.Minute,
		Burst:  10,
	})
	defer rl.Stop()

	var wg sync.WaitGroup
	workers := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			key := "user:" + strconv.Itoa(workerID)
			for j := 0; j < 50; j++ {
				rl.Allow(key)
			}
		}(i)
	}

	wg.Wait()
	// Each worker should have their own bucket
}

// ============================================================================
// Cleanup Tests
// ============================================================================

func TestCleanup_RemovesExpiredBuckets(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:    10,
		Window:  50 * time.Millisecond,
		Cleanup: 10 * time.Millisecond,
	})
	defer rl.Stop()

	// Create a bucket
	rl.Allow("user:123")

	// Verify bucket exists
	rl.mu.Lock()
	_, exists := rl.buckets["user:123"]
	rl.mu.Unlock()
	if !exists {
		t.Fatal("bucket should exist after request")
	}

	// Wait for cleanup (window * 2 + some buffer for cleanup to run)
	time.Sleep(150 * time.Millisecond)

	// Bucket should be cleaned up
	rl.mu.Lock()
	_, exists = rl.buckets["user:123"]
	rl.mu.Unlock()
	if exists {
		t.Error("expired bucket should have been cleaned up")
	}
}

func TestCleanup_KeepsFreshBuckets(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:    10,
		Window:  time.Minute,
		Cleanup: 10 * time.Millisecond,
	})
	defer rl.Stop()

	// Create a bucket
	rl.Allow("user:123")

	// Wait a bit for cleanup to run
	time.Sleep(50 * time.Millisecond)

	// Bucket should still exist (not expired yet)
	rl.mu.Lock()
	_, exists := rl.buckets["user:123"]
	rl.mu.Unlock()
	if !exists {
		t.Error("fresh bucket should not be cleaned up")
	}
}

func TestStop_StopsCleanupGoroutine(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{})

	// Stop should not panic or hang
	rl.Stop()

	// Double stop should not panic
	// (closing already closed channel, but we don't do that)
}

// ============================================================================
// RateLimit Middleware Tests
// ============================================================================

func TestRateLimitMiddleware_AllowedRequest_SetsHeaders(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   100,
		Window: time.Minute,
		Burst:  20,
	})
	defer rl.Stop()

	middleware := RateLimit(rl)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called")
	}

	// Check headers
	if rr.Header().Get("X-RateLimit-Limit") != "100" {
		t.Errorf("expected X-RateLimit-Limit '100', got %q", rr.Header().Get("X-RateLimit-Limit"))
	}
	if rr.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("expected X-RateLimit-Remaining header")
	}
	if rr.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("expected X-RateLimit-Reset header")
	}
}

func TestRateLimitMiddleware_DeniedRequest_Returns429(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   2,
		Window: time.Minute,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	middleware := RateLimit(rl)
	handler := &captureHandler{}

	// Exhaust the limit (rate+burst = 3 requests)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rr, req)
	}

	// Next request should be denied
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	handler.called = false
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rr2.Code)
	}
	if handler.called {
		t.Error("handler should not have been called")
	}

	// Check Retry-After header
	retryAfter := rr2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header")
	}
}

func TestRateLimitMiddleware_UsesUserID_WhenAuthenticated(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   2,
		Window: time.Minute,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	middleware := RateLimit(rl)
	handler := &captureHandler{}

	// Request with authenticated user
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req1.Context(), UserIDKey, "user:123")
	req1 = req1.WithContext(ctx)
	req1.RemoteAddr = "192.168.1.1:12345"

	// Exhaust user:123's limit (rate+burst = 3)
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rr, req1)
	}

	// Different user from same IP should still have quota
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx2 := context.WithValue(req2.Context(), UserIDKey, "user:456")
	req2 = req2.WithContext(ctx2)
	req2.RemoteAddr = "192.168.1.1:12345" // Same IP

	rr := httptest.NewRecorder()
	handler.called = false
	middleware(handler).ServeHTTP(rr, req2)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for different user, got %d", rr.Code)
	}
	if !handler.called {
		t.Error("handler should have been called for different user")
	}
}

func TestRateLimitMiddleware_UsesIP_WhenUnauthenticated(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   2,
		Window: time.Minute,
		Burst:  1, // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	middleware := RateLimit(rl)
	handler := &captureHandler{}

	// Unauthenticated requests from same IP
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Exhaust the limit (rate+burst = 3)
	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rr, req)
	}

	// Next request should be denied
	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rr.Code)
	}

	// Different IP should still have quota
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345" // Different IP

	rr2 := httptest.NewRecorder()
	handler.called = false
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status 200 for different IP, got %d", rr2.Code)
	}
}

func TestRateLimitMiddleware_RetryAfter_MinimumOne(t *testing.T) {
	t.Parallel()
	rl := NewRateLimiter(RateLimitConfig{
		Rate:   1,
		Window: time.Millisecond, // Very short window
		Burst:  1,                // Small burst (0 triggers default of 20)
	})
	defer rl.Stop()

	middleware := RateLimit(rl)
	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Exhaust the limit (rate+burst = 2)
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rr, req)
	}

	// Next request should be denied
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req)

	// If denied, Retry-After should be at least 1
	if rr2.Code == http.StatusTooManyRequests {
		retryAfter := rr2.Header().Get("Retry-After")
		if retryAfter != "" {
			val, err := strconv.Atoi(retryAfter)
			if err == nil && val < 1 {
				t.Errorf("Retry-After should be at least 1, got %d", val)
			}
		}
	}
}
