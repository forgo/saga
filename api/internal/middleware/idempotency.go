package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"
)

// IdempotencyStore stores idempotency key results
type IdempotencyStore struct {
	mu       sync.RWMutex
	entries  map[string]*idempotencyEntry
	ttl      time.Duration
	stopChan chan struct{}
}

type idempotencyEntry struct {
	status    int
	headers   http.Header
	body      []byte
	expiresAt time.Time
	inFlight  bool
	done      chan struct{}
}

// IdempotencyConfig holds configuration for idempotency middleware
type IdempotencyConfig struct {
	TTL     time.Duration // How long to keep idempotency results (default 24h)
	Cleanup time.Duration // Cleanup interval (default 1h)
}

// NewIdempotencyStore creates a new idempotency store
func NewIdempotencyStore(cfg IdempotencyConfig) *IdempotencyStore {
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	if cfg.Cleanup == 0 {
		cfg.Cleanup = time.Hour
	}

	store := &IdempotencyStore{
		entries:  make(map[string]*idempotencyEntry),
		ttl:      cfg.TTL,
		stopChan: make(chan struct{}),
	}

	go store.cleanupLoop(cfg.Cleanup)

	return store
}

// Stop stops the cleanup goroutine
func (s *IdempotencyStore) Stop() {
	close(s.stopChan)
}

func (s *IdempotencyStore) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopChan:
			return
		}
	}
}

func (s *IdempotencyStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.entries {
		if entry.expiresAt.Before(now) && !entry.inFlight {
			delete(s.entries, key)
		}
	}
}

// generateKey creates a unique key from user ID, idempotency key, and request fingerprint
func generateKey(userID, idempotencyKey, method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(userID))
	h.Write([]byte(idempotencyKey))
	h.Write([]byte(method))
	h.Write([]byte(path))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

// idempotencyResponseWriter captures the response for caching
type idempotencyResponseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *idempotencyResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Idempotency returns middleware that handles idempotency keys for POST/PATCH requests
func Idempotency(store *IdempotencyStore) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to POST and PATCH requests
			if r.Method != http.MethodPost && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			// Check for idempotency key header
			idempotencyKey := r.Header.Get("Idempotency-Key")
			if idempotencyKey == "" {
				// No idempotency key, proceed normally
				next.ServeHTTP(w, r)
				return
			}

			// Get user ID from context
			userID := GetUserID(r.Context())
			if userID == "" {
				userID = r.RemoteAddr // Fallback for unauthenticated requests
			}

			// Read and restore request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Generate composite key
			key := generateKey(userID, idempotencyKey, r.Method, r.URL.Path, body)

			// Check if we have a cached response
			store.mu.Lock()
			entry, exists := store.entries[key]

			if exists {
				if entry.inFlight {
					// Request is still processing, wait for it
					store.mu.Unlock()
					<-entry.done

					// Now get the completed entry
					store.mu.RLock()
					entry = store.entries[key]
					store.mu.RUnlock()

					if entry != nil && !entry.inFlight {
						// Return cached response
						for k, v := range entry.headers {
							for _, val := range v {
								w.Header().Add(k, val)
							}
						}
						w.Header().Set("X-Idempotency-Replayed", "true")
						w.WriteHeader(entry.status)
						_, _ = w.Write(entry.body)
						return
					}
				} else if entry.expiresAt.After(time.Now()) {
					// Return cached response
					store.mu.Unlock()
					for k, v := range entry.headers {
						for _, val := range v {
							w.Header().Add(k, val)
						}
					}
					w.Header().Set("X-Idempotency-Replayed", "true")
					w.WriteHeader(entry.status)
					_, _ = w.Write(entry.body)
					return
				}
			}

			// Create new entry to mark request as in-flight
			entry = &idempotencyEntry{
				inFlight: true,
				done:     make(chan struct{}),
			}
			store.entries[key] = entry
			store.mu.Unlock()

			// Wrap response writer to capture response
			irw := &idempotencyResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Process the request
			next.ServeHTTP(irw, r)

			// Cache the response
			store.mu.Lock()
			entry.status = irw.status
			entry.headers = irw.Header().Clone()
			entry.body = irw.body.Bytes()
			entry.expiresAt = time.Now().Add(store.ttl)
			entry.inFlight = false
			close(entry.done)
			store.mu.Unlock()
		})
	}
}
