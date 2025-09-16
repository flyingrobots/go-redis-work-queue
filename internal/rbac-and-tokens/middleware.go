// Copyright 2025 James Ross
package rbacandtokens

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ContextKey represents keys used in request context
type ContextKey string

const (
	ContextKeyClaims ContextKey = "rbac_claims"
	ContextKeyAuthz  ContextKey = "rbac_authz"
)

// AuthMiddleware creates an authentication middleware that validates tokens
func AuthMiddleware(manager *Manager, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health checks and token generation
			if r.URL.Path == "/health" || r.URL.Path == "/auth/token" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "AUTH_MISSING", "Authorization header required")
				return
			}

			claims, err := manager.ValidateToken(authHeader)
			if err != nil {
				logger.Debug("Token validation failed",
					zap.Error(err),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))

				if authErr, ok := err.(*AuthenticationError); ok {
					writeError(w, http.StatusUnauthorized, authErr.Code, authErr.Message)
				} else {
					writeError(w, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid authentication token")
				}
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthzMiddleware creates an authorization middleware that checks permissions
func AuthzMiddleware(manager *Manager, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authorization for health checks and auth endpoints
			if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Get claims from context
			claims, ok := r.Context().Value(ContextKeyClaims).(*Claims)
			if !ok {
				writeError(w, http.StatusUnauthorized, "NO_CLAIMS", "No authentication claims found")
				return
			}

			// Determine required permission based on endpoint
			permission := getRequiredPermission(r.Method, r.URL.Path)
			if permission == "" {
				// No specific permission required, allow through
				next.ServeHTTP(w, r)
				return
			}

			// Check authorization
			result, err := manager.Authorize(claims, permission, r.URL.Path)
			if err != nil {
				logger.Error("Authorization check failed",
					zap.Error(err),
					zap.String("subject", claims.Subject),
					zap.String("permission", string(permission)),
					zap.String("path", r.URL.Path))

				writeError(w, http.StatusInternalServerError, "AUTHZ_ERROR", "Authorization check failed")
				return
			}

			if !result.Allowed {
				logger.Warn("Access denied",
					zap.String("subject", claims.Subject),
					zap.String("permission", string(permission)),
					zap.String("path", r.URL.Path),
					zap.String("reason", result.Reason))

				// Log audit entry for denied access
				if manager.audit != nil {
					entry := AuditEntry{
						ID:        generateID(),
						Timestamp: time.Now(),
						Subject:   claims.Subject,
						Action:    "ACCESS_DENIED",
						Resource:  r.URL.Path,
						Result:    "DENIED",
						Reason:    result.Reason,
						Details: map[string]interface{}{
							"method":              r.Method,
							"required_permission": permission,
							"user_roles":          claims.Roles,
							"user_scopes":         claims.Scopes,
						},
						IP:        getClientIP(r),
						UserAgent: r.UserAgent(),
						RequestID: getRequestID(r),
					}
					manager.audit.Log(entry)
				}

				writeError(w, http.StatusForbidden, "ACCESS_DENIED",
					fmt.Sprintf("Insufficient permissions. Required: %s", permission))
				return
			}

			// Add authorization result to context
			ctx := context.WithValue(r.Context(), ContextKeyAuthz, result)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuditMiddleware creates middleware that logs successful authorized actions
func AuditMiddleware(manager *Manager, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapper, r)

			// Log audit entry for destructive operations
			if isDestructiveOperation(r.Method, r.URL.Path) && manager.audit != nil {
				claims, _ := r.Context().Value(ContextKeyClaims).(*Claims)
				authz, _ := r.Context().Value(ContextKeyAuthz).(*AuthorizationResult)

				entry := AuditEntry{
					ID:        generateID(),
					Timestamp: start,
					Action:    fmt.Sprintf("%s %s", r.Method, r.URL.Path),
					Resource:  r.URL.Path,
					Result:    fmt.Sprintf("%d", wrapper.statusCode),
					IP:        getClientIP(r),
					UserAgent: r.UserAgent(),
					RequestID: getRequestID(r),
					Details: map[string]interface{}{
						"duration_ms": time.Since(start).Milliseconds(),
						"method":      r.Method,
					},
				}

				if claims != nil {
					entry.Subject = claims.Subject
					entry.Details["roles"] = claims.Roles
					entry.Details["scopes"] = claims.Scopes
				}

				if authz != nil {
					entry.Details["authorization_reason"] = authz.Reason
				}

				manager.audit.Log(entry)
			}
		})
	}
}

// Helper functions

func getRequiredPermission(method, path string) Permission {
	endpointPerms := GetEndpointPermissions()

	// Try exact match first
	key := fmt.Sprintf("%s %s", method, path)
	if perms, exists := endpointPerms[key]; exists && len(perms) > 0 {
		return perms[0] // Return first required permission
	}

	// Try pattern matching
	for pattern, perms := range endpointPerms {
		if matchesPattern(key, pattern) && len(perms) > 0 {
			return perms[0] // Return first required permission
		}
	}

	return "" // No specific permission required
}

func matchesPattern(path, pattern string) bool {
	// Simple pattern matching - could be enhanced with regex
	if strings.Contains(pattern, "*") {
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

func isDestructiveOperation(method, path string) bool {
	if method != "DELETE" && method != "POST" {
		return false
	}

	destructivePaths := []string{
		"/api/v1/queues/dlq",
		"/api/v1/queues/all",
		"/api/v1/bench",
		"/auth/token/revoke",
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

func getRequestID(r *http.Request) string {
	if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
		return reqID
	}
	if reqID := r.Context().Value("request_id"); reqID != nil {
		if str, ok := reqID.(string); ok {
			return str
		}
	}
	return ""
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := fmt.Sprintf(`{"error": "%s", "code": "%s"}`, message, code)
	w.Write([]byte(response))
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