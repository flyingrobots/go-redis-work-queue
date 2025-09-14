// Copyright 2025 James Ross
package rbacandtokens

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager handles RBAC and token operations
type Manager struct {
	config       *Config
	keys         map[string]*KeyPair
	keysMutex    sync.RWMutex
	currentKeyID string
	revoked      map[string]*RevokedToken
	revokedMutex sync.RWMutex
	audit        *AuditLogger
	logger       *zap.Logger
	authzCache   map[string]*AuthorizationResult
	cacheMutex   sync.RWMutex
}

// NewManager creates a new RBAC and token manager
func NewManager(config *Config, audit *AuditLogger, logger *zap.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	m := &Manager{
		config:     config,
		keys:       make(map[string]*KeyPair),
		revoked:    make(map[string]*RevokedToken),
		audit:      audit,
		logger:     logger,
		authzCache: make(map[string]*AuthorizationResult),
	}

	// Initialize with a default key
	if err := m.generateDefaultKey(); err != nil {
		return nil, fmt.Errorf("failed to generate default key: %w", err)
	}

	// Start key rotation if configured
	if config.KeyConfig.RotationInterval > 0 {
		go m.rotateKeysRoutine()
	}

	// Start cache cleanup if enabled
	if config.AuthzConfig.CacheEnabled {
		go m.cacheCleanupRoutine()
	}

	return m, nil
}

// GenerateToken creates a new token for the given subject with specified roles and scopes
func (m *Manager) GenerateToken(subject string, roles []Role, scopes []Permission, ttl time.Duration) (string, error) {
	m.keysMutex.RLock()
	currentKey := m.keys[m.currentKeyID]
	m.keysMutex.RUnlock()

	if currentKey == nil {
		return "", NewTokenError(ErrKeyNotFound, "KEY_NOT_FOUND", "no active signing key available", "", subject, nil)
	}

	now := time.Now()
	if ttl <= 0 {
		ttl = m.config.TokenConfig.DefaultTTL
	}
	if ttl > m.config.TokenConfig.MaxTTL {
		ttl = m.config.TokenConfig.MaxTTL
	}

	claims := &Claims{
		Subject:   subject,
		Issuer:    m.config.TokenConfig.Issuer,
		Audience:  m.config.TokenConfig.Audience,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		JWTID:     generateID(),
		KeyID:     currentKey.ID,
		Roles:     roles,
		Scopes:    scopes,
		TokenType: TokenTypeBearer,
	}

	token, err := m.signToken(claims, currentKey)
	if err != nil {
		return "", NewTokenError(err, "TOKEN_SIGN_FAILED", "failed to sign token", claims.JWTID, subject, nil)
	}

	// Log token generation
	if m.audit != nil {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: now,
			Subject:   subject,
			Action:    "TOKEN_GENERATED",
			Resource:  "token",
			Result:    "SUCCESS",
			Details: map[string]interface{}{
				"token_id": claims.JWTID,
				"roles":    roles,
				"scopes":   scopes,
				"ttl":      ttl.String(),
			},
			RequestID: "", // Will be set by middleware
		}
		m.audit.Log(entry)
	}

	return token, nil
}

// ValidateToken validates a token and returns the claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, NewAuthenticationError(ErrMissingToken, "TOKEN_MISSING", "authentication token is required", nil)
	}

	// Remove Bearer prefix if present
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	claims, err := m.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is revoked
	m.revokedMutex.RLock()
	if revoked, exists := m.revoked[claims.JWTID]; exists {
		m.revokedMutex.RUnlock()
		return nil, NewAuthenticationError(ErrRevokedToken, "TOKEN_REVOKED",
			fmt.Sprintf("token was revoked at %s: %s", revoked.RevokedAt.Format(time.RFC3339), revoked.Reason),
			map[string]interface{}{
				"revoked_at": revoked.RevokedAt,
				"reason":     revoked.Reason,
			})
	}
	m.revokedMutex.RUnlock()

	// Validate timing
	now := time.Now().Unix()
	if now >= claims.ExpiresAt {
		return nil, NewAuthenticationError(ErrExpiredToken, "TOKEN_EXPIRED", "token has expired",
			map[string]interface{}{
				"expired_at": time.Unix(claims.ExpiresAt, 0),
				"now":        time.Unix(now, 0),
			})
	}

	if claims.NotBefore > 0 && now < claims.NotBefore {
		return nil, NewAuthenticationError(ErrTokenNotYetValid, "TOKEN_NOT_YET_VALID", "token is not yet valid",
			map[string]interface{}{
				"not_before": time.Unix(claims.NotBefore, 0),
				"now":        time.Unix(now, 0),
			})
	}

	return claims, nil
}

// Authorize checks if the given claims have the required permissions for an action on a resource
func (m *Manager) Authorize(claims *Claims, action Permission, resource string) (*AuthorizationResult, error) {
	if claims == nil {
		return &AuthorizationResult{
			Allowed: false,
			Reason:  "no claims provided",
		}, nil
	}

	// Check cache first
	if m.config.AuthzConfig.CacheEnabled {
		cacheKey := fmt.Sprintf("%s:%s:%s:%s", claims.Subject, action, resource, strings.Join(roleSliceToStringSlice(claims.Roles), ","))
		m.cacheMutex.RLock()
		if cached, exists := m.authzCache[cacheKey]; exists {
			m.cacheMutex.RUnlock()
			return cached, nil
		}
		m.cacheMutex.RUnlock()
	}

	// Admin role has all permissions
	for _, role := range claims.Roles {
		if role == RoleAdmin {
			result := &AuthorizationResult{
				Allowed: true,
				Subject: claims.Subject,
				Roles:   claims.Roles,
				Scopes:  claims.Scopes,
				Reason:  "admin role grants all permissions",
			}
			m.cacheAuthResult(fmt.Sprintf("%s:%s:%s:%s", claims.Subject, action, resource, strings.Join(roleSliceToStringSlice(claims.Roles), ",")), result)
			return result, nil
		}
	}

	// Check direct scopes first
	for _, scope := range claims.Scopes {
		if scope == action {
			result := &AuthorizationResult{
				Allowed: true,
				Subject: claims.Subject,
				Roles:   claims.Roles,
				Scopes:  claims.Scopes,
				Reason:  fmt.Sprintf("granted by scope: %s", scope),
			}
			m.cacheAuthResult(fmt.Sprintf("%s:%s:%s:%s", claims.Subject, action, resource, strings.Join(roleSliceToStringSlice(claims.Roles), ",")), result)
			return result, nil
		}
	}

	// Check role-based permissions
	rolePermissions := GetRolePermissions()
	for _, role := range claims.Roles {
		if permissions, exists := rolePermissions[role]; exists {
			for _, perm := range permissions {
				if perm == action {
					result := &AuthorizationResult{
						Allowed: true,
						Subject: claims.Subject,
						Roles:   claims.Roles,
						Scopes:  claims.Scopes,
						Reason:  fmt.Sprintf("granted by role: %s", role),
					}
					m.cacheAuthResult(fmt.Sprintf("%s:%s:%s:%s", claims.Subject, action, resource, strings.Join(roleSliceToStringSlice(claims.Roles), ",")), result)
					return result, nil
				}
			}
		}
	}

	// Access denied
	result := &AuthorizationResult{
		Allowed: false,
		Subject: claims.Subject,
		Roles:   claims.Roles,
		Scopes:  claims.Scopes,
		Reason:  fmt.Sprintf("insufficient permissions for action: %s", action),
	}

	if m.config.AuthzConfig.DefaultDeny {
		m.cacheAuthResult(fmt.Sprintf("%s:%s:%s:%s", claims.Subject, action, resource, strings.Join(roleSliceToStringSlice(claims.Roles), ",")), result)
	}

	return result, nil
}

// RevokeToken revokes a token by its JWT ID
func (m *Manager) RevokeToken(jwtID, reason string) error {
	now := time.Now()

	m.revokedMutex.Lock()
	m.revoked[jwtID] = &RevokedToken{
		JWTID:     jwtID,
		RevokedAt: now,
		Reason:    reason,
	}
	m.revokedMutex.Unlock()

	// Log revocation
	if m.audit != nil {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: now,
			Action:    "TOKEN_REVOKED",
			Resource:  "token",
			Result:    "SUCCESS",
			Reason:    reason,
			Details: map[string]interface{}{
				"token_id": jwtID,
			},
		}
		m.audit.Log(entry)
	}

	return nil
}

// GetTokenInfo returns information about a token for display purposes
func (m *Manager) GetTokenInfo(tokenString string) (*TokenInfo, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Subject:   claims.Subject,
		Roles:     claims.Roles,
		Scopes:    claims.Scopes,
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
		TokenType: claims.TokenType,
		KeyID:     claims.KeyID,
	}, nil
}

// Private methods

func (m *Manager) generateDefaultKey() error {
	keyID := generateID()

	// For HS256, we just need a symmetric key
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("failed to generate random key: %w", err)
	}

	keyPair := &KeyPair{
		ID:         keyID,
		Algorithm:  m.config.KeyConfig.Algorithm,
		PrivateKey: base64.StdEncoding.EncodeToString(key),
		PublicKey:  base64.StdEncoding.EncodeToString(key), // Same for symmetric
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(m.config.KeyConfig.RotationInterval),
		Active:     true,
	}

	m.keysMutex.Lock()
	m.keys[keyID] = keyPair
	m.currentKeyID = keyID
	m.keysMutex.Unlock()

	return nil
}

func (m *Manager) parseToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, NewAuthenticationError(ErrInvalidToken, "TOKEN_FORMAT_INVALID", "token must have 3 parts", nil)
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewAuthenticationError(ErrInvalidToken, "TOKEN_DECODE_FAILED", "failed to decode token payload", nil)
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, NewAuthenticationError(ErrInvalidToken, "TOKEN_PARSE_FAILED", "failed to parse token claims", nil)
	}

	// Verify signature
	if err := m.verifySignature(parts[0]+"."+parts[1], parts[2], claims.KeyID); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (m *Manager) signToken(claims *Claims, keyPair *KeyPair) (string, error) {
	// Create header
	header := map[string]interface{}{
		"alg": keyPair.Algorithm,
		"typ": "JWT",
		"kid": keyPair.ID,
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// Encode header and payload
	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signature
	message := headerB64 + "." + payloadB64
	signature, err := m.createSignature(message, keyPair)
	if err != nil {
		return "", err
	}

	return message + "." + signature, nil
}

func (m *Manager) createSignature(message string, keyPair *KeyPair) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyPair.PrivateKey)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signature), nil
}

func (m *Manager) verifySignature(message, signatureB64, keyID string) error {
	m.keysMutex.RLock()
	keyPair, exists := m.keys[keyID]
	m.keysMutex.RUnlock()

	if !exists {
		return NewAuthenticationError(ErrInvalidKeyID, "KEY_NOT_FOUND", fmt.Sprintf("key ID %s not found", keyID), nil)
	}

	expectedSig, err := m.createSignature(message, keyPair)
	if err != nil {
		return NewAuthenticationError(ErrInvalidSignature, "SIGNATURE_CREATE_FAILED", "failed to create expected signature", nil)
	}

	if !hmac.Equal([]byte(signatureB64), []byte(expectedSig)) {
		return NewAuthenticationError(ErrInvalidSignature, "SIGNATURE_MISMATCH", "token signature does not match", nil)
	}

	return nil
}

func (m *Manager) rotateKeysRoutine() {
	ticker := time.NewTicker(m.config.KeyConfig.RotationInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.generateDefaultKey(); err != nil {
			m.logger.Error("Failed to rotate keys", zap.Error(err))
		} else {
			m.logger.Info("Keys rotated successfully", zap.String("new_key_id", m.currentKeyID))
		}
	}
}

func (m *Manager) cacheCleanupRoutine() {
	ticker := time.NewTicker(m.config.AuthzConfig.CacheTTL)
	defer ticker.Stop()

	for range ticker.C {
		m.cacheMutex.Lock()
		// Simple cache cleanup - in production this would be more sophisticated
		m.authzCache = make(map[string]*AuthorizationResult)
		m.cacheMutex.Unlock()
	}
}

func (m *Manager) cacheAuthResult(key string, result *AuthorizationResult) {
	if !m.config.AuthzConfig.CacheEnabled {
		return
	}

	m.cacheMutex.Lock()
	m.authzCache[key] = result
	m.cacheMutex.Unlock()
}

// Helper functions

func generateID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

func roleSliceToStringSlice(roles []Role) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}
	return result
}