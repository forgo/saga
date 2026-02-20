package middleware

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// Chain Tests
// ============================================================================

func TestChain_NoMiddlewares_ReturnsHandler(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("handler"))
	})

	result := Chain(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	result.ServeHTTP(rr, req)

	if rr.Body.String() != "handler" {
		t.Errorf("expected body 'handler', got %q", rr.Body.String())
	}
}

func TestChain_SingleMiddleware_Applies(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("handler"))
	})

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("before-"))
			next.ServeHTTP(w, r)
			_, _ = w.Write([]byte("-after"))
		})
	}

	result := Chain(handler, middleware)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	result.ServeHTTP(rr, req)

	if rr.Body.String() != "before-handler-after" {
		t.Errorf("expected 'before-handler-after', got %q", rr.Body.String())
	}
}

func TestChain_MultipleMiddlewares_AppliesInOrder(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("H"))
	})

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("1"))
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("2"))
			next.ServeHTTP(w, r)
		})
	}
	mw3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("3"))
			next.ServeHTTP(w, r)
		})
	}

	result := Chain(handler, mw1, mw2, mw3)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	result.ServeHTTP(rr, req)

	// Middlewares should execute in order: mw1 -> mw2 -> mw3 -> handler
	if rr.Body.String() != "123H" {
		t.Errorf("expected '123H', got %q", rr.Body.String())
	}
}

// ============================================================================
// RequestID Tests
// ============================================================================

func TestRequestID_NoHeader_GeneratesNew(t *testing.T) {
	t.Parallel()

	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	RequestID(handler).ServeHTTP(rr, req)

	// Check response header is set
	responseID := rr.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	// Check context has request ID
	contextID := GetRequestID(handler.ctx)
	if contextID == "" {
		t.Error("expected request ID in context")
	}
	if contextID != responseID {
		t.Errorf("context ID (%q) should match response header (%q)", contextID, responseID)
	}
}

func TestRequestID_WithHeader_PreservesExisting(t *testing.T) {
	t.Parallel()

	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	rr := httptest.NewRecorder()

	RequestID(handler).ServeHTTP(rr, req)

	responseID := rr.Header().Get("X-Request-ID")
	if responseID != "existing-request-id" {
		t.Errorf("expected preserved ID 'existing-request-id', got %q", responseID)
	}

	contextID := GetRequestID(handler.ctx)
	if contextID != "existing-request-id" {
		t.Errorf("expected context ID 'existing-request-id', got %q", contextID)
	}
}

func TestRequestID_GeneratedID_IsUUID(t *testing.T) {
	t.Parallel()

	handler := &captureHandler{}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	RequestID(handler).ServeHTTP(rr, req)

	requestID := rr.Header().Get("X-Request-ID")

	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
	if len(requestID) != 36 {
		t.Errorf("expected UUID length 36, got %d", len(requestID))
	}
	if strings.Count(requestID, "-") != 4 {
		t.Errorf("expected 4 hyphens in UUID, got %d", strings.Count(requestID, "-"))
	}
}

func TestGetRequestID_Present_ReturnsValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), RequestIDKey, "req-12345")
	result := GetRequestID(ctx)

	if result != "req-12345" {
		t.Errorf("expected 'req-12345', got %q", result)
	}
}

func TestGetRequestID_Missing_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result := GetRequestID(ctx)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestGetRequestID_WrongType_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), RequestIDKey, 12345)
	result := GetRequestID(ctx)

	if result != "" {
		t.Errorf("expected empty string for wrong type, got %q", result)
	}
}

// ============================================================================
// Recovery Tests
// ============================================================================

func TestRecovery_NoPanic_ProceedsNormally(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	Recovery(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "success" {
		t.Errorf("expected body 'success', got %q", rr.Body.String())
	}
}

func TestRecovery_WithPanic_Returns500(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	Recovery(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRecovery_WithPanic_ReturnsJSON(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	Recovery(handler).ServeHTTP(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
	if !strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Errorf("expected error message in body, got %q", rr.Body.String())
	}
}

func TestRecovery_WithNilPanic_Recovers(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// Should not panic - nil panic is recovered but skipped
	Recovery(handler).ServeHTTP(rr, req)

	// With nil panic, recover() returns nil, so no 500 is written
	// The response will be empty/default
}

// ============================================================================
// CORS Tests
// ============================================================================

func TestCORS_AllowedOrigin_SetsHeader(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS([]string{"https://example.com", "https://app.example.com"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'https://example.com', got %q", allowOrigin)
	}
}

func TestCORS_DisallowedOrigin_NoHeader(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS([]string{"https://allowed.com"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("expected no Access-Control-Allow-Origin header, got %q", allowOrigin)
	}
}

func TestCORS_WildcardOrigin_AllowsAny(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS([]string{"*"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "https://any-origin.com" {
		t.Errorf("expected origin to be allowed with wildcard, got %q", allowOrigin)
	}
}

func TestCORS_PreflightRequest_Returns204(t *testing.T) {
	t.Parallel()

	handler := &captureHandler{}
	corsMiddleware := CORS([]string{"https://example.com"})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d for preflight, got %d", http.StatusNoContent, rr.Code)
	}
	if handler.called {
		t.Error("handler should not be called for preflight request")
	}
}

func TestCORS_SetsRequiredHeaders(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS([]string{"https://example.com"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	// Check all required headers are set
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
	if rr.Header().Get("Access-Control-Expose-Headers") == "" {
		t.Error("expected Access-Control-Expose-Headers header")
	}
	if rr.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("expected Access-Control-Max-Age header")
	}
}

func TestCORS_NoOriginHeader_ProceedsWithoutCORS(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := CORS([]string{"https://example.com"})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header
	rr := httptest.NewRecorder()

	corsMiddleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	// Access-Control-Allow-Origin should not be set without Origin header
	allowOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("expected no Allow-Origin header without Origin, got %q", allowOrigin)
	}
}

// ============================================================================
// Compress Tests
// ============================================================================

func TestCompress_AcceptsGzip_CompressesResponse(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, this is a test response that should be compressed."))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rr := httptest.NewRecorder()

	Compress(handler).ServeHTTP(rr, req)

	encoding := rr.Header().Get("Content-Encoding")
	if encoding != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got %q", encoding)
	}

	// Body should be gzip compressed
	reader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func() { _ = reader.Close() }()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read decompressed data: %v", err)
	}

	if string(decompressed) != "Hello, this is a test response that should be compressed." {
		t.Errorf("decompressed content mismatch: %q", string(decompressed))
	}
}

func TestCompress_NoGzipAccept_DoesNotCompress(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("uncompressed response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Accept-Encoding header
	rr := httptest.NewRecorder()

	Compress(handler).ServeHTTP(rr, req)

	encoding := rr.Header().Get("Content-Encoding")
	if encoding == "gzip" {
		t.Error("should not compress without gzip Accept-Encoding")
	}

	if rr.Body.String() != "uncompressed response" {
		t.Errorf("expected uncompressed body, got %q", rr.Body.String())
	}
}

func TestCompress_SSERequest_DoesNotCompress(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("event: message\ndata: test\n\n"))
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	Compress(handler).ServeHTTP(rr, req)

	encoding := rr.Header().Get("Content-Encoding")
	if encoding == "gzip" {
		t.Error("should not compress SSE responses")
	}
}

// ============================================================================
// Logger Tests (via responseWriter)
// ============================================================================

func TestResponseWriter_CapturesStatusCode(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)

	if rw.statusCode != http.StatusCreated {
		t.Errorf("expected captured status %d, got %d", http.StatusCreated, rw.statusCode)
	}
	if rr.Code != http.StatusCreated {
		t.Errorf("expected forwarded status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestResponseWriter_DefaultStatusOK(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr, statusCode: http.StatusOK}

	// Don't call WriteHeader, just write body
	_, _ = rw.Write([]byte("body"))

	// Default should be 200 OK
	if rw.statusCode != http.StatusOK {
		t.Errorf("expected default status %d, got %d", http.StatusOK, rw.statusCode)
	}
}

// ============================================================================
// gzipResponseWriter Tests
// ============================================================================

func TestGzipResponseWriter_WritesToGzipWriter(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	gz := gzip.NewWriter(rr)
	grw := &gzipResponseWriter{ResponseWriter: rr, Writer: gz}

	_, err := grw.Write([]byte("compressed content"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	_ = gz.Close()

	// Verify we can decompress
	reader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func() { _ = reader.Close() }()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if string(content) != "compressed content" {
		t.Errorf("expected 'compressed content', got %q", string(content))
	}
}

// ============================================================================
// Logger Integration Test (basic)
// ============================================================================

func TestLogger_CompletesRequest(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/items", nil)
	rr := httptest.NewRecorder()

	// Logger should complete without error
	Logger(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if rr.Body.String() != "created" {
		t.Errorf("expected body 'created', got %q", rr.Body.String())
	}
}
