package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// NewIdempotencyStore Tests
// ============================================================================

func TestNewIdempotencyStore_DefaultConfig(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{})
	defer store.Stop()

	if store.ttl != 24*time.Hour {
		t.Errorf("expected TTL 24h, got %v", store.ttl)
	}
	if store.entries == nil {
		t.Error("entries map should be initialized")
	}
}

func TestNewIdempotencyStore_CustomConfig(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{
		TTL:     time.Hour,
		Cleanup: 5 * time.Minute,
	})
	defer store.Stop()

	if store.ttl != time.Hour {
		t.Errorf("expected TTL 1h, got %v", store.ttl)
	}
}

func TestIdempotencyStore_Stop_StopsCleanupLoop(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{
		TTL:     time.Hour,
		Cleanup: time.Millisecond, // Very short for testing
	})

	// Give cleanup a chance to run
	time.Sleep(10 * time.Millisecond)

	// Stop should not hang
	done := make(chan struct{})
	go func() {
		store.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Stop() did not return within timeout")
	}
}

// ============================================================================
// generateKey Tests
// ============================================================================

func TestGenerateKey_SameInputs_ProducesSameKey(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{"a":1}`))
	key2 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{"a":1}`))

	if key1 != key2 {
		t.Errorf("expected same key, got %s and %s", key1, key2)
	}
}

func TestGenerateKey_DifferentUserIDs_ProducesDifferentKeys(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{}`))
	key2 := generateKey("user:2", "idem-key", "POST", "/api/test", []byte(`{}`))

	if key1 == key2 {
		t.Error("different user IDs should produce different keys")
	}
}

func TestGenerateKey_DifferentIdempotencyKeys_ProducesDifferentKeys(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "key-a", "POST", "/api/test", []byte(`{}`))
	key2 := generateKey("user:1", "key-b", "POST", "/api/test", []byte(`{}`))

	if key1 == key2 {
		t.Error("different idempotency keys should produce different keys")
	}
}

func TestGenerateKey_DifferentMethods_ProducesDifferentKeys(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{}`))
	key2 := generateKey("user:1", "idem-key", "PATCH", "/api/test", []byte(`{}`))

	if key1 == key2 {
		t.Error("different methods should produce different keys")
	}
}

func TestGenerateKey_DifferentPaths_ProducesDifferentKeys(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "idem-key", "POST", "/api/a", []byte(`{}`))
	key2 := generateKey("user:1", "idem-key", "POST", "/api/b", []byte(`{}`))

	if key1 == key2 {
		t.Error("different paths should produce different keys")
	}
}

func TestGenerateKey_DifferentBodies_ProducesDifferentKeys(t *testing.T) {
	t.Parallel()
	key1 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{"a":1}`))
	key2 := generateKey("user:1", "idem-key", "POST", "/api/test", []byte(`{"a":2}`))

	if key1 == key2 {
		t.Error("different bodies should produce different keys")
	}
}

func TestGenerateKey_EmptyBody_IsValid(t *testing.T) {
	t.Parallel()
	key := generateKey("user:1", "idem-key", "POST", "/api/test", nil)

	if len(key) != 64 { // SHA256 = 32 bytes = 64 hex chars
		t.Errorf("expected 64 char hex string, got %d chars", len(key))
	}
}

// ============================================================================
// HTTP Method Filtering Tests
// ============================================================================

func TestIdempotency_SkipsGET_Proceeds(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := &captureHandler{}
	middleware := Idempotency(store)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Idempotency-Key", "test-key")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should be called for GET")
	}
	if rr.Header().Get("X-Idempotency-Replayed") != "" {
		t.Error("GET should not be idempotent")
	}
}

func TestIdempotency_SkipsDELETE_Proceeds(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := &captureHandler{}
	middleware := Idempotency(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/test", nil)
	req.Header.Set("Idempotency-Key", "test-key")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should be called for DELETE")
	}
}

func TestIdempotency_SkipsPUT_Proceeds(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := &captureHandler{}
	middleware := Idempotency(store)

	req := httptest.NewRequest(http.MethodPut, "/api/test", nil)
	req.Header.Set("Idempotency-Key", "test-key")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should be called for PUT")
	}
}

// ============================================================================
// No Idempotency Key Tests
// ============================================================================

func TestIdempotency_POST_NoKey_ProceedsNormally(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})
	middleware := Idempotency(store)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Second request (no idempotency key, so should execute again)
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if callCount != 2 {
		t.Errorf("expected handler called twice, got %d", callCount)
	}
}

func TestIdempotency_PATCH_NoKey_ProceedsNormally(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := &captureHandler{}
	middleware := Idempotency(store)

	req := httptest.NewRequest(http.MethodPatch, "/api/test", bytes.NewReader([]byte(`{}`)))
	// No Idempotency-Key header
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handler.called {
		t.Error("handler should be called")
	}
}

// ============================================================================
// Cache Miss and Cache Hit Tests
// ============================================================================

func TestIdempotency_CacheMiss_ProcessesAndCaches(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	})
	middleware := Idempotency(store)

	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "unique-key")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if rr.Body.String() != `{"id":"123"}` {
		t.Errorf("expected body %q, got %q", `{"id":"123"}`, rr.Body.String())
	}
	if rr.Header().Get("X-Idempotency-Replayed") != "" {
		t.Error("first request should not be replayed")
	}
}

func TestIdempotency_CacheHit_ReturnsCachedResponse(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"123"}`))
	})
	middleware := Idempotency(store)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "same-key")
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Second request with same key
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "same-key")
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if callCount != 1 {
		t.Errorf("expected handler called once, got %d", callCount)
	}
	if rr2.Code != http.StatusCreated {
		t.Errorf("expected cached status %d, got %d", http.StatusCreated, rr2.Code)
	}
	if rr2.Body.String() != `{"id":"123"}` {
		t.Errorf("expected cached body %q, got %q", `{"id":"123"}`, rr2.Body.String())
	}
	if rr2.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Error("replayed request should have X-Idempotency-Replayed header")
	}
}

func TestIdempotency_CacheHit_CopiesOriginalHeaders(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-123")
		w.Header().Add("X-Multi", "value1")
		w.Header().Add("X-Multi", "value2")
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "header-test")
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Second request
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "header-test")
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if rr2.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type header, got %q", rr2.Header().Get("Content-Type"))
	}
	if rr2.Header().Get("X-Request-Id") != "req-123" {
		t.Errorf("expected X-Request-Id header, got %q", rr2.Header().Get("X-Request-Id"))
	}
	multiVals := rr2.Header().Values("X-Multi")
	if len(multiVals) != 2 {
		t.Errorf("expected 2 X-Multi values, got %d", len(multiVals))
	}
}

// ============================================================================
// User ID vs RemoteAddr Tests
// ============================================================================

func TestIdempotency_UsesUserID_WhenAuthenticated(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// Request from user A
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "shared-key")
	ctx1 := context.WithValue(req1.Context(), UserIDKey, "user:A")
	req1 = req1.WithContext(ctx1)
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Request from user B with same idempotency key
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "shared-key")
	ctx2 := context.WithValue(req2.Context(), UserIDKey, "user:B")
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	// Different users should get different cache keys
	if callCount != 2 {
		t.Errorf("expected handler called twice (different users), got %d", callCount)
	}
}

func TestIdempotency_FallsBackToRemoteAddr_WhenUnauthenticated(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// Request from IP A
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "shared-key")
	req1.RemoteAddr = "10.0.0.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Request from IP B with same idempotency key
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "shared-key")
	req2.RemoteAddr = "10.0.0.2:54321"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	// Different IPs should get different cache keys
	if callCount != 2 {
		t.Errorf("expected handler called twice (different IPs), got %d", callCount)
	}
}

// ============================================================================
// In-Flight Request Handling Tests
// ============================================================================

func TestIdempotency_InFlight_SecondRequestWaits(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	var callCount int32
	requestStarted := make(chan struct{})
	proceedWithHandler := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		close(requestStarted)
		<-proceedWithHandler
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"result":"done"}`))
	})
	middleware := Idempotency(store)

	var wg sync.WaitGroup
	results := make([]*httptest.ResponseRecorder, 2)

	// First request (will be in-flight)
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Idempotency-Key", "inflight-key")
		req.RemoteAddr = "192.168.1.1:12345"
		results[0] = httptest.NewRecorder()
		middleware(handler).ServeHTTP(results[0], req)
	}()

	// Wait for first request to start processing
	<-requestStarted

	// Second request (should wait)
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Idempotency-Key", "inflight-key")
		req.RemoteAddr = "192.168.1.1:12345"
		results[1] = httptest.NewRecorder()
		middleware(handler).ServeHTTP(results[1], req)
	}()

	// Give second request time to start waiting
	time.Sleep(50 * time.Millisecond)

	// Let handler complete
	close(proceedWithHandler)
	wg.Wait()

	// Handler should only be called once
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected handler called once, got %d", callCount)
	}

	// Both should return same result
	if results[0].Code != http.StatusCreated {
		t.Errorf("first request: expected status %d, got %d", http.StatusCreated, results[0].Code)
	}
	if results[1].Code != http.StatusCreated {
		t.Errorf("second request: expected status %d, got %d", http.StatusCreated, results[1].Code)
	}
	if results[1].Header().Get("X-Idempotency-Replayed") != "true" {
		t.Error("second request should have X-Idempotency-Replayed header")
	}
}

// ============================================================================
// Cleanup Tests
// ============================================================================

func TestIdempotencyStore_Cleanup_RemovesExpiredEntries(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{
		TTL:     100 * time.Millisecond,
		Cleanup: time.Hour, // Manual cleanup
	})
	defer store.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// Make a request to create an entry
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "cleanup-test")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	// Entry should exist
	store.mu.RLock()
	entryCount := len(store.entries)
	store.mu.RUnlock()
	if entryCount != 1 {
		t.Errorf("expected 1 entry, got %d", entryCount)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Manually trigger cleanup
	store.cleanup()

	// Entry should be removed
	store.mu.RLock()
	entryCount = len(store.entries)
	store.mu.RUnlock()
	if entryCount != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", entryCount)
	}
}

func TestIdempotencyStore_Cleanup_KeepsNonExpiredEntries(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{
		TTL:     time.Hour,
		Cleanup: time.Hour,
	})
	defer store.Stop()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// Make a request to create an entry
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "keep-test")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr, req)

	// Trigger cleanup
	store.cleanup()

	// Entry should still exist
	store.mu.RLock()
	entryCount := len(store.entries)
	store.mu.RUnlock()
	if entryCount != 1 {
		t.Errorf("expected 1 entry (not expired), got %d", entryCount)
	}
}

// ============================================================================
// Response Writer Tests
// ============================================================================

func TestIdempotencyResponseWriter_CapturesStatus(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	irw := &idempotencyResponseWriter{
		ResponseWriter: rr,
		status:         http.StatusOK,
	}

	irw.WriteHeader(http.StatusCreated)

	if irw.status != http.StatusCreated {
		t.Errorf("expected captured status %d, got %d", http.StatusCreated, irw.status)
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("expected forwarded status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestIdempotencyResponseWriter_CapturesBody(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	irw := &idempotencyResponseWriter{
		ResponseWriter: rr,
		status:         http.StatusOK,
	}

	_, err := irw.Write([]byte("test body"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if irw.body.String() != "test body" {
		t.Errorf("expected captured body %q, got %q", "test body", irw.body.String())
	}
	if rr.Body.String() != "test body" {
		t.Errorf("expected forwarded body %q, got %q", "test body", rr.Body.String())
	}
}

func TestIdempotencyResponseWriter_MultipleWrites(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	irw := &idempotencyResponseWriter{
		ResponseWriter: rr,
		status:         http.StatusOK,
	}

	_, _ = irw.Write([]byte("part1"))
	_, _ = irw.Write([]byte("part2"))

	if irw.body.String() != "part1part2" {
		t.Errorf("expected combined body %q, got %q", "part1part2", irw.body.String())
	}
}

// ============================================================================
// Body Restoration Tests
// ============================================================================

func TestIdempotency_RestoresRequestBody(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	var receivedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	originalBody := `{"key":"value","nested":{"a":1}}`
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(originalBody)))
	req.Header.Set("Idempotency-Key", "body-test")
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if string(receivedBody) != originalBody {
		t.Errorf("expected body %q, got %q", originalBody, string(receivedBody))
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestIdempotency_ExpiredEntry_ProcessesAgain(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{
		TTL:     50 * time.Millisecond,
		Cleanup: time.Hour,
	})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response"))
	})
	middleware := Idempotency(store)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "expire-test")
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Second request after expiration
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "expire-test")
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if callCount != 2 {
		t.Errorf("expected handler called twice (after expiration), got %d", callCount)
	}
	// Second request should NOT be replayed (it's a fresh request)
	if rr2.Header().Get("X-Idempotency-Replayed") != "" {
		t.Error("request after expiration should not be replayed")
	}
}

func TestIdempotency_EmptyIdempotencyKey_ProceedsNormally(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	middleware := Idempotency(store)

	// Request with empty idempotency key header
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Idempotency-Key", "")
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if callCount != 1 {
		t.Errorf("expected handler called once, got %d", callCount)
	}
}

func TestIdempotency_PATCH_WithKey_Works(t *testing.T) {
	t.Parallel()
	store := NewIdempotencyStore(IdempotencyConfig{TTL: time.Hour})
	defer store.Stop()

	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("patched"))
	})
	middleware := Idempotency(store)

	// First PATCH request
	req1 := httptest.NewRequest(http.MethodPatch, "/api/test", bytes.NewReader([]byte(`{}`)))
	req1.Header.Set("Idempotency-Key", "patch-key")
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr1, req1)

	// Second PATCH request with same key
	req2 := httptest.NewRequest(http.MethodPatch, "/api/test", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Idempotency-Key", "patch-key")
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rr2, req2)

	if callCount != 1 {
		t.Errorf("expected handler called once for PATCH, got %d", callCount)
	}
	if rr2.Header().Get("X-Idempotency-Replayed") != "true" {
		t.Error("second PATCH should be replayed")
	}
}
