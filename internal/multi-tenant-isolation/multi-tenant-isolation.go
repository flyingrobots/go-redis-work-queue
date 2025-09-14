// Copyright 2025 James Ross
package multitenantiso

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// TenantManager handles tenant operations and isolation
type TenantManager struct {
	redis     *redis.Client
	encryptor *PayloadEncryptor
	ctx       context.Context
}

// NewTenantManager creates a new tenant manager instance
func NewTenantManager(redisClient *redis.Client) *TenantManager {
	return &TenantManager{
		redis:     redisClient,
		encryptor: NewPayloadEncryptor(),
		ctx:       context.Background(),
	}
}

// CreateTenant creates a new tenant with the given configuration
func (tm *TenantManager) CreateTenant(config *TenantConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	// Check if tenant already exists
	exists, err := tm.TenantExists(config.ID)
	if err != nil {
		return NewStorageError("check", tm.getTenantConfigKey(config.ID), config.ID, err)
	}
	if exists {
		return NewTenantAlreadyExistsError(config.ID)
	}

	// Set creation timestamp
	config.CreatedAt = time.Now()
	config.UpdatedAt = config.CreatedAt

	// Set default values if not provided
	if config.Status == "" {
		config.Status = TenantStatusActive
	}
	if config.Quotas == (TenantQuotas{}) {
		config.Quotas = DefaultQuotas()
	}

	// Serialize and store tenant config
	configJSON, err := config.ToJSON()
	if err != nil {
		return NewStorageError("marshal", tm.getTenantConfigKey(config.ID), config.ID, err)
	}

	err = tm.redis.Set(tm.ctx, tm.getTenantConfigKey(config.ID), configJSON, 0).Err()
	if err != nil {
		return NewStorageError("set", tm.getTenantConfigKey(config.ID), config.ID, err)
	}

	// Initialize quota usage tracking
	usage := &QuotaUsage{
		TenantID:    config.ID,
		LastUpdated: time.Now(),
	}
	usageJSON, _ := json.Marshal(usage)
	_ = tm.redis.Set(tm.ctx, tm.getQuotaUsageKey(config.ID), usageJSON, 0)

	// Add to tenant index
	_ = tm.redis.SAdd(tm.ctx, "tenants:index", string(config.ID))

	return nil
}

// GetTenant retrieves tenant configuration by ID
func (tm *TenantManager) GetTenant(tenantID TenantID) (*TenantConfig, error) {
	if err := tenantID.Validate(); err != nil {
		return nil, err
	}

	configJSON, err := tm.redis.Get(tm.ctx, tm.getTenantConfigKey(tenantID)).Result()
	if err == redis.Nil {
		return nil, NewTenantNotFoundError(tenantID)
	}
	if err != nil {
		return nil, NewStorageError("get", tm.getTenantConfigKey(tenantID), tenantID, err)
	}

	var config TenantConfig
	if err := config.FromJSON([]byte(configJSON)); err != nil {
		return nil, NewStorageError("unmarshal", tm.getTenantConfigKey(tenantID), tenantID, err)
	}

	return &config, nil
}

// UpdateTenant updates tenant configuration
func (tm *TenantManager) UpdateTenant(config *TenantConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	// Check if tenant exists
	exists, err := tm.TenantExists(config.ID)
	if err != nil {
		return err
	}
	if !exists {
		return NewTenantNotFoundError(config.ID)
	}

	// Update timestamp
	config.UpdatedAt = time.Now()

	configJSON, err := config.ToJSON()
	if err != nil {
		return NewStorageError("marshal", tm.getTenantConfigKey(config.ID), config.ID, err)
	}

	err = tm.redis.Set(tm.ctx, tm.getTenantConfigKey(config.ID), configJSON, 0).Err()
	if err != nil {
		return NewStorageError("set", tm.getTenantConfigKey(config.ID), config.ID, err)
	}

	return nil
}

// DeleteTenant marks a tenant for deletion and cleans up resources
func (tm *TenantManager) DeleteTenant(tenantID TenantID) error {
	if err := tenantID.Validate(); err != nil {
		return err
	}

	config, err := tm.GetTenant(tenantID)
	if err != nil {
		return err
	}

	// Mark as deleted
	config.Status = TenantStatusDeleted
	config.UpdatedAt = time.Now()

	if err := tm.UpdateTenant(config); err != nil {
		return err
	}

	// Clean up tenant data
	return tm.cleanupTenantData(tenantID)
}

// TenantExists checks if a tenant exists
func (tm *TenantManager) TenantExists(tenantID TenantID) (bool, error) {
	if err := tenantID.Validate(); err != nil {
		return false, err
	}

	exists, err := tm.redis.Exists(tm.ctx, tm.getTenantConfigKey(tenantID)).Result()
	if err != nil {
		return false, NewStorageError("exists", tm.getTenantConfigKey(tenantID), tenantID, err)
	}

	return exists == 1, nil
}

// ListTenants returns a list of tenant summaries
func (tm *TenantManager) ListTenants() ([]TenantSummary, error) {
	tenantIDs, err := tm.redis.SMembers(tm.ctx, "tenants:index").Result()
	if err != nil {
		return nil, NewStorageError("smembers", "tenants:index", "", err)
	}

	summaries := make([]TenantSummary, 0, len(tenantIDs))
	for _, idStr := range tenantIDs {
		tenantID := TenantID(idStr)
		config, err := tm.GetTenant(tenantID)
		if err != nil {
			continue // Skip problematic tenants
		}

		if config.Status == TenantStatusDeleted {
			continue
		}

		usage, _ := tm.GetQuotaUsage(tenantID)

		summary := TenantSummary{
			ID:           config.ID,
			Name:         config.Name,
			Status:       string(config.Status),
			ActiveQueues: int(usage.ActiveQueues),
			QuotaHealth:  tm.calculateQuotaHealth(config, usage),
			LastActivity: usage.LastUpdated,
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// ValidateAccess checks if a user has permission to perform an action on a resource
func (tm *TenantManager) ValidateAccess(userID string, tenantID TenantID, resource, action string) error {
	// Get user's permissions for this tenant
	permissions, err := tm.getUserPermissions(userID, tenantID)
	if err != nil {
		return err
	}

	// Check if user has required permission
	hasPermission := false
	for _, perm := range permissions {
		if perm.TenantID == tenantID && perm.Resource == resource {
			for _, allowedAction := range perm.Actions {
				if allowedAction == action || allowedAction == "admin" {
					hasPermission = true
					break
				}
			}
		}
	}

	if !hasPermission {
		return NewAccessDeniedError(userID, tenantID, resource, action, "insufficient permissions")
	}

	return nil
}

// CheckQuota validates if a tenant can perform an operation within quota limits
func (tm *TenantManager) CheckQuota(tenantID TenantID, quotaType string, amount int64) error {
	config, err := tm.GetTenant(tenantID)
	if err != nil {
		return err
	}

	usage, err := tm.GetQuotaUsage(tenantID)
	if err != nil {
		return err
	}

	// Check specific quota type
	switch quotaType {
	case "jobs_per_hour":
		if usage.JobsThisHour+amount > config.Quotas.MaxJobsPerHour {
			return NewQuotaExceededError(tenantID, quotaType, usage.JobsThisHour+amount, config.Quotas.MaxJobsPerHour)
		}
	case "jobs_per_day":
		if usage.JobsThisDay+amount > config.Quotas.MaxJobsPerDay {
			return NewQuotaExceededError(tenantID, quotaType, usage.JobsThisDay+amount, config.Quotas.MaxJobsPerDay)
		}
	case "backlog_size":
		if usage.CurrentBacklogSize+amount > config.Quotas.MaxBacklogSize {
			return NewQuotaExceededError(tenantID, quotaType, usage.CurrentBacklogSize+amount, config.Quotas.MaxBacklogSize)
		}
	case "storage_bytes":
		if usage.StorageUsedBytes+amount > config.Quotas.MaxStorageBytes {
			return NewQuotaExceededError(tenantID, quotaType, usage.StorageUsedBytes+amount, config.Quotas.MaxStorageBytes)
		}
	}

	return nil
}

// IncrementQuotaUsage updates quota usage counters
func (tm *TenantManager) IncrementQuotaUsage(tenantID TenantID, quotaType string, amount int64) error {
	usage, err := tm.GetQuotaUsage(tenantID)
	if err != nil {
		return err
	}

	// Update counters
	switch quotaType {
	case "jobs_per_hour":
		usage.JobsThisHour += amount
	case "jobs_per_day":
		usage.JobsThisDay += amount
	case "backlog_size":
		usage.CurrentBacklogSize += amount
	case "storage_bytes":
		usage.StorageUsedBytes += amount
	}

	usage.LastUpdated = time.Now()
	return tm.updateQuotaUsage(tenantID, usage)
}

// GetQuotaUsage retrieves current quota usage for a tenant
func (tm *TenantManager) GetQuotaUsage(tenantID TenantID) (*QuotaUsage, error) {
	usageJSON, err := tm.redis.Get(tm.ctx, tm.getQuotaUsageKey(tenantID)).Result()
	if err == redis.Nil {
		// Return empty usage if not found
		return &QuotaUsage{
			TenantID:    tenantID,
			LastUpdated: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, NewStorageError("get", tm.getQuotaUsageKey(tenantID), tenantID, err)
	}

	var usage QuotaUsage
	if err := json.Unmarshal([]byte(usageJSON), &usage); err != nil {
		return nil, NewStorageError("unmarshal", tm.getQuotaUsageKey(tenantID), tenantID, err)
	}

	return &usage, nil
}

// GetTenantNamespace returns a key namespace for tenant operations
func (tm *TenantManager) GetTenantNamespace(tenantID TenantID) KeyNamespace {
	return KeyNamespace{TenantID: tenantID}
}

// LogAuditEvent records an audit event for a tenant action
func (tm *TenantManager) LogAuditEvent(event *AuditEvent) error {
	if event.EventID == "" {
		event.EventID = tm.generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return NewAuditError(event.Action, event.TenantID, err)
	}

	// Store event with timestamp-based key for chronological ordering
	eventKey := fmt.Sprintf("audit:%s:%d:%s", event.TenantID, event.Timestamp.Unix(), event.EventID)
	err = tm.redis.Set(tm.ctx, eventKey, eventJSON, 24*time.Hour*90).Err() // Keep for 90 days
	if err != nil {
		return NewAuditError(event.Action, event.TenantID, err)
	}

	// Add to tenant's audit index
	indexKey := tm.getAuditIndexKey(event.TenantID)
	_ = tm.redis.ZAdd(tm.ctx, indexKey, &redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: eventKey,
	})

	return nil
}

// Helper methods for Redis key generation
func (tm *TenantManager) getTenantConfigKey(tenantID TenantID) string {
	return "tenant:" + string(tenantID) + ":config"
}

func (tm *TenantManager) getQuotaUsageKey(tenantID TenantID) string {
	return "tenant:" + string(tenantID) + ":quotas"
}

func (tm *TenantManager) getAuditIndexKey(tenantID TenantID) string {
	return "tenant:" + string(tenantID) + ":audit"
}

func (tm *TenantManager) getUserPermissionsKey(userID string) string {
	return "user:" + userID + ":permissions"
}

func (tm *TenantManager) generateEventID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Helper methods
func (tm *TenantManager) getUserPermissions(userID string, tenantID TenantID) ([]TenantPermission, error) {
	permJSON, err := tm.redis.Get(tm.ctx, tm.getUserPermissionsKey(userID)).Result()
	if err == redis.Nil {
		return []TenantPermission{}, nil
	}
	if err != nil {
		return nil, NewStorageError("get", tm.getUserPermissionsKey(userID), tenantID, err)
	}

	var access UserTenantAccess
	if err := json.Unmarshal([]byte(permJSON), &access); err != nil {
		return nil, NewStorageError("unmarshal", tm.getUserPermissionsKey(userID), tenantID, err)
	}

	return access.Permissions, nil
}

func (tm *TenantManager) updateQuotaUsage(tenantID TenantID, usage *QuotaUsage) error {
	usageJSON, err := json.Marshal(usage)
	if err != nil {
		return NewStorageError("marshal", tm.getQuotaUsageKey(tenantID), tenantID, err)
	}

	err = tm.redis.Set(tm.ctx, tm.getQuotaUsageKey(tenantID), usageJSON, 0).Err()
	if err != nil {
		return NewStorageError("set", tm.getQuotaUsageKey(tenantID), tenantID, err)
	}

	return nil
}

func (tm *TenantManager) calculateQuotaHealth(config *TenantConfig, usage *QuotaUsage) string {
	// Calculate health based on soft limit threshold
	threshold := config.Quotas.SoftLimitThreshold

	checks := []float64{
		float64(usage.JobsThisHour) / float64(config.Quotas.MaxJobsPerHour),
		float64(usage.JobsThisDay) / float64(config.Quotas.MaxJobsPerDay),
		float64(usage.CurrentBacklogSize) / float64(config.Quotas.MaxBacklogSize),
		float64(usage.StorageUsedBytes) / float64(config.Quotas.MaxStorageBytes),
	}

	maxUsage := 0.0
	for _, usage := range checks {
		if usage > maxUsage {
			maxUsage = usage
		}
	}

	if maxUsage >= 1.0 {
		return "critical"
	}
	if maxUsage >= threshold {
		return "warning"
	}
	return "good"
}

func (tm *TenantManager) cleanupTenantData(tenantID TenantID) error {
	// Get all tenant keys
	ns := tm.GetTenantNamespace(tenantID)
	pattern := ns.AllKeysPattern()

	keys, err := tm.redis.Keys(tm.ctx, pattern).Result()
	if err != nil {
		return NewStorageError("keys", pattern, tenantID, err)
	}

	if len(keys) > 0 {
		err = tm.redis.Del(tm.ctx, keys...).Err()
		if err != nil {
			return NewStorageError("del", "tenant_data", tenantID, err)
		}
	}

	// Remove from tenant index
	_ = tm.redis.SRem(tm.ctx, "tenants:index", string(tenantID))

	return nil
}

// PayloadEncryptor handles tenant-specific payload encryption
type PayloadEncryptor struct{}

// NewPayloadEncryptor creates a new payload encryptor
func NewPayloadEncryptor() *PayloadEncryptor {
	return &PayloadEncryptor{}
}

// EncryptPayload encrypts a payload using tenant-specific encryption settings
func (pe *PayloadEncryptor) EncryptPayload(payload []byte, tenantConfig *TenantConfig) (*EncryptedPayload, error) {
	if !tenantConfig.Encryption.Enabled {
		return nil, ErrEncryptionNotEnabled
	}

	// Generate a random data encryption key (DEK)
	dek := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, NewEncryptionError("generate_dek", tenantConfig.ID, err)
	}

	// Encrypt the payload with AES-GCM
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, NewEncryptionError("create_cipher", tenantConfig.ID, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, NewEncryptionError("create_gcm", tenantConfig.ID, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, NewEncryptionError("generate_nonce", tenantConfig.ID, err)
	}

	ciphertext := gcm.Seal(nil, nonce, payload, nil)

	// Encrypt DEK with KEK (simplified - in production, use cloud KMS)
	encryptedDEK, err := pe.encryptDEKWithKEK(dek, tenantConfig.Encryption.KEKKeyID)
	if err != nil {
		return nil, NewEncryptionError("encrypt_dek", tenantConfig.ID, err)
	}

	return &EncryptedPayload{
		Version:          1,
		EncryptedDEK:     encryptedDEK,
		EncryptedPayload: ciphertext[:len(ciphertext)-gcm.Overhead()],
		Nonce:            nonce,
		AuthTag:          ciphertext[len(ciphertext)-gcm.Overhead():],
		CreatedAt:        time.Now().Unix(),
	}, nil
}

// DecryptPayload decrypts an encrypted payload
func (pe *PayloadEncryptor) DecryptPayload(encrypted *EncryptedPayload, tenantConfig *TenantConfig) ([]byte, error) {
	if !tenantConfig.Encryption.Enabled {
		return nil, ErrEncryptionNotEnabled
	}

	// Decrypt DEK with KEK
	dek, err := pe.decryptDEKWithKEK(encrypted.EncryptedDEK, tenantConfig.Encryption.KEKKeyID)
	if err != nil {
		return nil, NewEncryptionError("decrypt_dek", tenantConfig.ID, err)
	}

	// Decrypt payload
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, NewEncryptionError("create_cipher", tenantConfig.ID, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, NewEncryptionError("create_gcm", tenantConfig.ID, err)
	}

	// Reconstruct ciphertext with auth tag
	ciphertext := append(encrypted.EncryptedPayload, encrypted.AuthTag...)

	plaintext, err := gcm.Open(nil, encrypted.Nonce, ciphertext, nil)
	if err != nil {
		return nil, NewEncryptionError("decrypt", tenantConfig.ID, err)
	}

	return plaintext, nil
}

// Simplified KEK operations (in production, integrate with cloud KMS)
func (pe *PayloadEncryptor) encryptDEKWithKEK(dek []byte, kekKeyID string) ([]byte, error) {
	// This is a simplified implementation
	// In production, this would call AWS KMS, Google Cloud KMS, etc.
	hash := sha256.Sum256([]byte(kekKeyID + "salt"))
	key := hash[:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, dek, nil)
	return ciphertext, nil
}

func (pe *PayloadEncryptor) decryptDEKWithKEK(encryptedDEK []byte, kekKeyID string) ([]byte, error) {
	// This is a simplified implementation
	hash := sha256.Sum256([]byte(kekKeyID + "salt"))
	key := hash[:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedDEK) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedDEK[:nonceSize], encryptedDEK[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// RateLimitChecker provides rate limiting functionality for tenants
type RateLimitChecker struct {
	redis *redis.Client
	ctx   context.Context
}

// NewRateLimitChecker creates a new rate limit checker
func NewRateLimitChecker(redisClient *redis.Client) *RateLimitChecker {
	return &RateLimitChecker{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

// CheckRateLimit validates if a tenant operation is within rate limits
func (rlc *RateLimitChecker) CheckRateLimit(tenantID TenantID, operation string, config *TenantConfig) error {
	if !config.RateLimiting.Enabled {
		return nil
	}

	// Get rate limit for this operation
	var limit int32
	switch operation {
	case "enqueue":
		limit = config.Quotas.EnqueueRateLimit
	case "dequeue":
		limit = config.Quotas.DequeueRateLimit
	default:
		if customLimit, ok := config.RateLimiting.CustomLimits[operation]; ok {
			limit = customLimit
		} else {
			return nil // No limit configured for this operation
		}
	}

	if limit <= 0 {
		return nil // No limit
	}

	// Use sliding window rate limiting
	key := fmt.Sprintf("rate_limit:%s:%s", tenantID, operation)
	window := config.RateLimiting.WindowDuration.Seconds()

	// Current time window
	now := time.Now().Unix()
	windowStart := now - int64(window)

	// Remove expired entries
	_ = rlc.redis.ZRemRangeByScore(rlc.ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests in window
	count, err := rlc.redis.ZCard(rlc.ctx, key).Result()
	if err != nil {
		return NewStorageError("zcard", key, tenantID, err)
	}

	// Check if we exceed the limit
	if count >= int64(limit) {
		retryAfter := int(window)
		return NewRateLimitExceededError(tenantID, operation, int32(count), limit, retryAfter)
	}

	// Add current request to window
	member := fmt.Sprintf("%d:%s", now, rlc.generateRequestID())
	_ = rlc.redis.ZAdd(rlc.ctx, key, &redis.Z{
		Score:  float64(now),
		Member: member,
	})

	// Set expiration on the key
	_ = rlc.redis.Expire(rlc.ctx, key, time.Duration(window)*time.Second)

	return nil
}

func (rlc *RateLimitChecker) generateRequestID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}