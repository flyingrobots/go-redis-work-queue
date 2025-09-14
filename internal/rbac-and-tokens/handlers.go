// Copyright 2025 James Ross
package rbacandtokens

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Handler provides HTTP handlers for RBAC operations
type Handler struct {
	manager *Manager
	logger  *zap.Logger
}

// NewHandler creates a new RBAC handler
func NewHandler(manager *Manager, logger *zap.Logger) *Handler {
	return &Handler{
		manager: manager,
		logger:  logger,
	}
}

// TokenRequest represents a token generation request
type TokenRequest struct {
	Subject string       `json:"subject"`
	Roles   []Role       `json:"roles"`
	Scopes  []Permission `json:"scopes,omitempty"`
	TTL     string       `json:"ttl,omitempty"` // Duration string like "24h"
}

// TokenResponse represents a token generation response
type TokenResponse struct {
	Token     string    `json:"token"`
	Subject   string    `json:"subject"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

// TokenInfoResponse represents token information response
type TokenInfoResponse struct {
	*TokenInfo
	Valid bool `json:"valid"`
}

// RevokeTokenRequest represents a token revocation request
type RevokeTokenRequest struct {
	TokenID string `json:"token_id"`
	Reason  string `json:"reason"`
}

// AuditQueryRequest represents an audit log query request
type AuditQueryRequest struct {
	Subject   string `json:"subject,omitempty"`
	Action    string `json:"action,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Result    string `json:"result,omitempty"`
	StartTime string `json:"start_time,omitempty"` // RFC3339 format
	EndTime   string `json:"end_time,omitempty"`   // RFC3339 format
	IP        string `json:"ip,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// GenerateToken handles POST /auth/token
func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Subject == "" {
		h.writeError(w, http.StatusBadRequest, "SUBJECT_REQUIRED", "Subject is required")
		return
	}

	if len(req.Roles) == 0 {
		h.writeError(w, http.StatusBadRequest, "ROLES_REQUIRED", "At least one role is required")
		return
	}

	// Parse TTL if provided
	var ttl time.Duration
	if req.TTL != "" {
		var err error
		ttl, err = time.ParseDuration(req.TTL)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "INVALID_TTL", "Invalid TTL format")
			return
		}
	}

	// Generate scopes from roles if not provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = h.getScopesForRoles(req.Roles)
	}

	token, err := h.manager.GenerateToken(req.Subject, req.Roles, scopes, ttl)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate token")
		return
	}

	// Get token info for response
	info, err := h.manager.GetTokenInfo(token)
	if err != nil {
		h.logger.Error("Failed to get token info", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "TOKEN_INFO_FAILED", "Failed to get token info")
		return
	}

	response := TokenResponse{
		Token:     token,
		Subject:   info.Subject,
		ExpiresAt: info.ExpiresAt,
		TokenType: string(info.TokenType),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ValidateToken handles POST /auth/validate
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, http.StatusUnauthorized, "AUTH_MISSING", "Authorization header required")
		return
	}

	claims, err := h.manager.ValidateToken(authHeader)
	if err != nil {
		h.logger.Debug("Token validation failed", zap.Error(err))
		h.writeError(w, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid token")
		return
	}

	info := &TokenInfo{
		Subject:   claims.Subject,
		Roles:     claims.Roles,
		Scopes:    claims.Scopes,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
		TokenType: claims.TokenType,
		KeyID:     claims.KeyID,
	}

	response := TokenInfoResponse{
		TokenInfo: info,
		Valid:     true,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetTokenInfo handles GET /auth/token/info
func (h *Handler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, http.StatusUnauthorized, "AUTH_MISSING", "Authorization header required")
		return
	}

	info, err := h.manager.GetTokenInfo(authHeader)
	if err != nil {
		response := TokenInfoResponse{
			Valid: false,
		}
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	response := TokenInfoResponse{
		TokenInfo: info,
		Valid:     true,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// RevokeToken handles POST /auth/token/revoke
func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.TokenID == "" {
		h.writeError(w, http.StatusBadRequest, "TOKEN_ID_REQUIRED", "Token ID is required")
		return
	}

	if req.Reason == "" {
		req.Reason = "Revoked via API"
	}

	if err := h.manager.RevokeToken(req.TokenID, req.Reason); err != nil {
		h.logger.Error("Failed to revoke token", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "REVOCATION_FAILED", "Failed to revoke token")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Token revoked successfully",
	})
}

// CheckAccess handles POST /auth/check
func (h *Handler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	// Extract action and resource from query parameters
	action := Permission(r.URL.Query().Get("action"))
	resource := r.URL.Query().Get("resource")

	if action == "" {
		h.writeError(w, http.StatusBadRequest, "ACTION_REQUIRED", "Action parameter is required")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		result := &AuthorizationResult{
			Allowed: false,
			Reason:  "no authentication token provided",
		}
		h.writeJSON(w, http.StatusOK, result)
		return
	}

	claims, err := h.manager.ValidateToken(authHeader)
	if err != nil {
		result := &AuthorizationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("invalid token: %s", err.Error()),
		}
		h.writeJSON(w, http.StatusOK, result)
		return
	}

	result, err := h.manager.Authorize(claims, action, resource)
	if err != nil {
		h.logger.Error("Authorization check failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "AUTHZ_CHECK_FAILED", "Authorization check failed")
		return
	}

	result.RequestID = h.getRequestID(r)
	h.writeJSON(w, http.StatusOK, result)
}

// QueryAudit handles GET /audit/query
func (h *Handler) QueryAudit(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := AuditFilter{
		Subject:  r.URL.Query().Get("subject"),
		Action:   r.URL.Query().Get("action"),
		Resource: r.URL.Query().Get("resource"),
		Result:   r.URL.Query().Get("result"),
		IP:       r.URL.Query().Get("ip"),
		RequestID: r.URL.Query().Get("request_id"),
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// Parse time range
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = start
		}
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = end
		}
	}

	// Default limit
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	entries, err := h.manager.audit.Query(filter)
	if err != nil {
		h.logger.Error("Failed to query audit log", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "AUDIT_QUERY_FAILED", "Failed to query audit log")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
		"filter":  filter,
	})
}

// Helper methods

func (h *Handler) getScopesForRoles(roles []Role) []Permission {
	rolePerms := GetRolePermissions()
	scopeSet := make(map[Permission]bool)

	for _, role := range roles {
		if perms, exists := rolePerms[role]; exists {
			for _, perm := range perms {
				scopeSet[perm] = true
			}
		}
	}

	scopes := make([]Permission, 0, len(scopeSet))
	for scope := range scopeSet {
		scopes = append(scopes, scope)
	}

	return scopes
}

func (h *Handler) getRequestID(r *http.Request) string {
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

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	response := map[string]interface{}{
		"error": message,
		"code":  code,
	}
	h.writeJSON(w, status, response)
}