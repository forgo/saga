package middleware

import (
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to a handler in order
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
)

// RequestID adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Logger logs request details using structured logging
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		requestID := GetRequestID(r.Context())

		slog.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.statusCode),
			slog.Duration("duration", duration),
			slog.String("request_id", requestID),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

// Recovery recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				slog.Error("panic recovered",
					slog.Any("error", err),
					slog.String("request_id", requestID),
					slog.String("stack", string(debug.Stack())),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"type":"https://saga-api.forgo.software/errors/internal","title":"Internal Server Error","status":500,"detail":"An unexpected error occurred"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS returns a middleware that handles CORS
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, Idempotency-Key")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Compress compresses responses using gzip when supported
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression for SSE
		if r.Header.Get("Accept") == "text/event-stream" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length") // Length will change after compression

		gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// gzipResponseWriter wraps http.ResponseWriter with gzip
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}
