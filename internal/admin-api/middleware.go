// Copyright 2025 James Ross
package adminapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type contextKey string

const (
	contextKeyClaims    contextKey = "claims"
	contextKeyRequestID contextKey = "request_id"
	contextKeyUserIP    contextKey = "user_ip"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(secret string, denyByDefault bool, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !denyByDefault {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "AUTH_MISSING", "Authorization header required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "AUTH_INVALID", "Invalid authorization format")
				return
			}

			claims, err := validateJWT(parts[1], secret)
			if err != nil {
				logger.Warn("JWT validation failed", zap.Error(err))
				writeError(w, http.StatusUnauthorized, "AUTH_INVALID", "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RateLimitMiddleware implements token bucket rate limiting
func RateLimitMiddleware(perMinute int, burst int, logger *zap.Logger) func(http.Handler) http.Handler {
	buckets := &sync.Map{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token identifier
			var key string
			if claims, ok := r.Context().Value(contextKeyClaims).(*Claims); ok {
				key = claims.Subject
			} else {
				key = getClientIP(r)
			}

			// Get or create bucket
			val, _ := buckets.LoadOrStore(key, &rateBucket{
				tokens:    float64(burst),
				lastFill:  time.Now(),
				maxTokens: burst,
				fillRate:  float64(perMinute) / 60.0,
			})
			bucket := val.(*rateBucket)

			// Check rate limit
			if !bucket.consume() {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", perMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
				writeError(w, http.StatusTooManyRequests, "RATE_LIMIT", "Rate limit exceeded")
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", perMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int(bucket.tokens)))

			next.ServeHTTP(w, r)
		})
	}
}

// AuditMiddleware logs all API actions
func AuditMiddleware(auditLog *AuditLogger, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Capture response
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r)

			// Log audit entry for destructive operations
			if isDestructiveOperation(r.Method, r.URL.Path) {
				entry := AuditEntry{
					ID:        generateID(),
					Timestamp: start,
					Action:    fmt.Sprintf("%s %s", r.Method, r.URL.Path),
					Result:    fmt.Sprintf("%d", rw.statusCode),
					IP:        getClientIP(r),
					UserAgent: r.UserAgent(),
				}

				if claims, ok := r.Context().Value(contextKeyClaims).(*Claims); ok {
					entry.User = claims.Subject
				}

				// Extract reason from body if present
				if r.Method == "DELETE" {
					var body PurgeRequest
					if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
						entry.Reason = body.Reason
					}
				}

				if err := auditLog.Log(entry); err != nil {
					logger.Error("Failed to write audit log", zap.Error(err))
				}
			}
		})
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			for _, ao := range allowedOrigins {
				if ao == "*" || ao == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware adds a unique request ID
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateID()
			}

			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RecoveryMiddleware handles panics
func RecoveryMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method))
					writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func validateJWT(tokenString string, secret string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	// Verify signature
	message := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSig := h.Sum(nil)

	if !hmac.Equal(sig, expectedSig) {
		return nil, fmt.Errorf("invalid signature")
	}

	return &claims, nil
}

func isDestructiveOperation(method, path string) bool {
	if method != "DELETE" && method != "POST" {
		return false
	}

	destructivePaths := []string{
		"/api/v1/queues/dlq",
		"/api/v1/queues/all",
		"/api/v1/bench",
	}

	for _, dp := range destructivePaths {
		if strings.Contains(path, dp) {
			return true
		}
	}

	return false
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// Rate bucket implementation
type rateBucket struct {
	mu        sync.Mutex
	tokens    float64
	lastFill  time.Time
	maxTokens int
	fillRate  float64
}

func (b *rateBucket) consume() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens = min(float64(b.maxTokens), b.tokens+elapsed*b.fillRate)
	b.lastFill = now

	// Try to consume
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Response writer wrapper
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}