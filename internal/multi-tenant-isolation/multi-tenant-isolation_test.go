// Copyright 2025 James Ross
package multitenantiso

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test database
	})

	// Clear test database
	err := client.FlushDB(context.Background()).Err()
	require.NoError(t, err)

	return client
}

func TestTenantManager_CreateTenant(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: DefaultQuotas(),
	}

	// Test successful creation
	err := tm.CreateTenant(config)
	assert.NoError(t, err)

	// Test duplicate creation
	err = tm.CreateTenant(config)
	assert.Error(t, err)
	assert.True(t, IsTenantNotFound(err))

	// Verify tenant exists
	exists, err := tm.TenantExists("test-tenant")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestTenantManager_GetTenant(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: DefaultQuotas(),
	}

	// Test get non-existent tenant
	_, err := tm.GetTenant("non-existent")
	assert.Error(t, err)
	assert.True(t, IsTenantNotFound(err))

	// Create and get tenant
	err = tm.CreateTenant(config)
	require.NoError(t, err)

	retrieved, err := tm.GetTenant("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, config.ID, retrieved.ID)
	assert.Equal(t, config.Name, retrieved.Name)
	assert.Equal(t, TenantStatusActive, retrieved.Status)
}

func TestTenantManager_UpdateTenant(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: DefaultQuotas(),
	}

	// Test update non-existent tenant
	err := tm.UpdateTenant(config)
	assert.Error(t, err)
	assert.True(t, IsTenantNotFound(err))

	// Create and update tenant
	err = tm.CreateTenant(config)
	require.NoError(t, err)

	config.Name = "Updated Tenant"
	config.Status = TenantStatusSuspended

	err = tm.UpdateTenant(config)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := tm.GetTenant("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Tenant", retrieved.Name)
	assert.Equal(t, TenantStatusSuspended, retrieved.Status)
}

func TestTenantManager_DeleteTenant(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: DefaultQuotas(),
	}

	// Test delete non-existent tenant
	err := tm.DeleteTenant("non-existent")
	assert.Error(t, err)
	assert.True(t, IsTenantNotFound(err))

	// Create and delete tenant
	err = tm.CreateTenant(config)
	require.NoError(t, err)

	err = tm.DeleteTenant("test-tenant")
	assert.NoError(t, err)

	// Verify tenant is marked as deleted
	retrieved, err := tm.GetTenant("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, TenantStatusDeleted, retrieved.Status)
}

func TestTenantManager_ListTenants(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	// Test empty list
	summaries, err := tm.ListTenants()
	assert.NoError(t, err)
	assert.Empty(t, summaries)

	// Create multiple tenants
	tenants := []TenantConfig{
		{ID: "tenant-1", Name: "Tenant 1", Quotas: DefaultQuotas()},
		{ID: "tenant-2", Name: "Tenant 2", Quotas: DefaultQuotas()},
		{ID: "tenant-3", Name: "Tenant 3", Quotas: DefaultQuotas()},
	}

	for _, tenant := range tenants {
		err := tm.CreateTenant(&tenant)
		require.NoError(t, err)
	}

	// Delete one tenant
	err = tm.DeleteTenant("tenant-2")
	require.NoError(t, err)

	// List should return only active tenants
	summaries, err = tm.ListTenants()
	assert.NoError(t, err)
	assert.Len(t, summaries, 2)

	names := make([]string, len(summaries))
	for i, summary := range summaries {
		names[i] = summary.Name
	}
	assert.Contains(t, names, "Tenant 1")
	assert.Contains(t, names, "Tenant 3")
	assert.NotContains(t, names, "Tenant 2")
}

func TestTenantManager_CheckQuota(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: TenantQuotas{
			MaxJobsPerHour: 100,
			MaxJobsPerDay:  1000,
			MaxBacklogSize: 50,
		},
	}

	err := tm.CreateTenant(config)
	require.NoError(t, err)

	// Test quota within limits
	err = tm.CheckQuota("test-tenant", "jobs_per_hour", 50)
	assert.NoError(t, err)

	// Test quota exceeding limits
	err = tm.CheckQuota("test-tenant", "jobs_per_hour", 150)
	assert.Error(t, err)
	assert.True(t, IsQuotaExceeded(err))

	// Increment usage and test again
	err = tm.IncrementQuotaUsage("test-tenant", "jobs_per_hour", 80)
	assert.NoError(t, err)

	err = tm.CheckQuota("test-tenant", "jobs_per_hour", 30)
	assert.Error(t, err)
	assert.True(t, IsQuotaExceeded(err))

	err = tm.CheckQuota("test-tenant", "jobs_per_hour", 15)
	assert.NoError(t, err)
}

func TestTenantManager_QuotaUsage(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	config := &TenantConfig{
		ID:   "test-tenant",
		Name: "Test Tenant",
		Quotas: DefaultQuotas(),
	}

	err := tm.CreateTenant(config)
	require.NoError(t, err)

	// Get initial usage
	usage, err := tm.GetQuotaUsage("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), usage.JobsThisHour)
	assert.Equal(t, int64(0), usage.CurrentBacklogSize)

	// Increment usage
	err = tm.IncrementQuotaUsage("test-tenant", "jobs_per_hour", 25)
	assert.NoError(t, err)

	err = tm.IncrementQuotaUsage("test-tenant", "backlog_size", 10)
	assert.NoError(t, err)

	// Verify updated usage
	usage, err = tm.GetQuotaUsage("test-tenant")
	assert.NoError(t, err)
	assert.Equal(t, int64(25), usage.JobsThisHour)
	assert.Equal(t, int64(10), usage.CurrentBacklogSize)
}

func TestTenantManager_ValidateAccess(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	// Test access denial for non-existent permissions
	err := tm.ValidateAccess("user1", "test-tenant", "queues", "read")
	assert.Error(t, err)
	assert.True(t, IsAccessDenied(err))
}

func TestTenantManager_AuditLogging(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	tm := NewTenantManager(redisClient)

	event := &AuditEvent{
		TenantID: "test-tenant",
		UserID:   "user1",
		Action:   "CREATE_QUEUE",
		Resource: "t:test-tenant:queue1",
		Details:  map[string]interface{}{"queue_name": "queue1"},
		Result:   "SUCCESS",
	}

	err := tm.LogAuditEvent(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.EventID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestTenantID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      TenantID
		wantErr bool
	}{
		{"valid id", "test-tenant", false},
		{"valid with numbers", "tenant-123", false},
		{"too short", "ab", true},
		{"too long", "a-very-long-tenant-id-that-exceeds-32-characters", true},
		{"starts with hyphen", "-invalid", true},
		{"ends with hyphen", "invalid-", true},
		{"contains uppercase", "Test-Tenant", true},
		{"contains underscore", "test_tenant", true},
		{"contains space", "test tenant", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyNamespace(t *testing.T) {
	ns := KeyNamespace{TenantID: "test-tenant"}

	assert.Equal(t, "t:test-tenant:queue1", ns.QueueKey("queue1"))
	assert.Equal(t, "t:test-tenant:queue1:jobs", ns.JobsKey("queue1"))
	assert.Equal(t, "t:test-tenant:queue1:dlq", ns.DLQKey("queue1"))
	assert.Equal(t, "t:test-tenant:queue1:workers", ns.WorkersKey("queue1"))
	assert.Equal(t, "t:test-tenant:queue1:metrics", ns.MetricsKey("queue1"))
	assert.Equal(t, "tenant:test-tenant:config", ns.ConfigKey())
	assert.Equal(t, "tenant:test-tenant:quotas", ns.QuotasKey())
	assert.Equal(t, "tenant:test-tenant:keys", ns.KeysKey())
	assert.Equal(t, "tenant:test-tenant:audit", ns.AuditKey())
	assert.Equal(t, "t:test-tenant:*", ns.AllKeysPattern())
}

func TestDefaultQuotas(t *testing.T) {
	quotas := DefaultQuotas()

	assert.Equal(t, int64(10000), quotas.MaxJobsPerHour)
	assert.Equal(t, int64(100000), quotas.MaxJobsPerDay)
	assert.Equal(t, int64(50000), quotas.MaxBacklogSize)
	assert.Equal(t, int64(1024*1024), quotas.MaxJobSizeBytes)
	assert.Equal(t, int32(10), quotas.MaxQueuesPerTenant)
	assert.Equal(t, int32(50), quotas.MaxWorkersPerQueue)
	assert.Equal(t, float64(0.8), quotas.SoftLimitThreshold)
}

func TestTenantConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TenantConfig
		wantErr bool
	}{
		{
			"valid config",
			TenantConfig{
				ID:     "test-tenant",
				Name:   "Test Tenant",
				Quotas: DefaultQuotas(),
			},
			false,
		},
		{
			"invalid tenant ID",
			TenantConfig{
				ID:     "a", // Too short
				Name:   "Test Tenant",
				Quotas: DefaultQuotas(),
			},
			true,
		},
		{
			"missing name",
			TenantConfig{
				ID:     "test-tenant",
				Name:   "",
				Quotas: DefaultQuotas(),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Validate that default status is set
				if tt.config.Status == "" {
					assert.Equal(t, TenantStatusActive, tt.config.Status)
				}
			}
		})
	}
}

func TestTenantConfig_JSON(t *testing.T) {
	config := &TenantConfig{
		ID:     "test-tenant",
		Name:   "Test Tenant",
		Status: TenantStatusActive,
		Quotas: DefaultQuotas(),
	}

	// Test serialization
	data, err := config.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test deserialization
	var newConfig TenantConfig
	err = newConfig.FromJSON(data)
	assert.NoError(t, err)
	assert.Equal(t, config.ID, newConfig.ID)
	assert.Equal(t, config.Name, newConfig.Name)
	assert.Equal(t, config.Status, newConfig.Status)
}

func TestErrorTypes(t *testing.T) {
	// Test TenantNotFoundError
	err := NewTenantNotFoundError("test-tenant")
	assert.True(t, IsTenantNotFound(err))
	assert.Contains(t, err.Error(), "test-tenant")

	// Test AccessDeniedError
	err = NewAccessDeniedError("user1", "test-tenant", "queues", "read", "no permission")
	assert.True(t, IsAccessDenied(err))
	assert.Contains(t, err.Error(), "user1")
	assert.Contains(t, err.Error(), "test-tenant")

	// Test QuotaExceededError
	err = NewQuotaExceededError("test-tenant", "jobs_per_hour", 150, 100)
	assert.True(t, IsQuotaExceeded(err))
	assert.Contains(t, err.Error(), "150")
	assert.Contains(t, err.Error(), "100")

	// Test EncryptionError
	origErr := assert.AnError
	err = NewEncryptionError("encrypt", "test-tenant", origErr)
	assert.True(t, IsEncryptionError(err))
	assert.Contains(t, err.Error(), "encrypt")
}

func TestPayloadEncryptor_Encryption(t *testing.T) {
	pe := NewPayloadEncryptor()

	config := &TenantConfig{
		ID: "test-tenant",
		Encryption: TenantEncryption{
			Enabled:     true,
			KEKProvider: "local",
			KEKKeyID:    "test-key-123",
			Algorithm:   "AES-256-GCM",
		},
	}

	payload := []byte("sensitive data to encrypt")

	// Test encryption
	encrypted, err := pe.EncryptPayload(payload, config)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEmpty(t, encrypted.EncryptedDEK)
	assert.NotEmpty(t, encrypted.EncryptedPayload)
	assert.NotEmpty(t, encrypted.Nonce)
	assert.NotEmpty(t, encrypted.AuthTag)
	assert.Equal(t, 1, encrypted.Version)

	// Test decryption
	decrypted, err := pe.DecryptPayload(encrypted, config)
	assert.NoError(t, err)
	assert.Equal(t, payload, decrypted)

	// Test encryption disabled
	config.Encryption.Enabled = false
	_, err = pe.EncryptPayload(payload, config)
	assert.Error(t, err)
	assert.Equal(t, ErrEncryptionNotEnabled, err)
}

func TestRateLimitChecker(t *testing.T) {
	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	rlc := NewRateLimitChecker(redisClient)

	config := &TenantConfig{
		ID: "test-tenant",
		RateLimiting: TenantRateLimiting{
			Enabled:        true,
			WindowDuration: time.Second,
		},
		Quotas: TenantQuotas{
			EnqueueRateLimit: 5, // 5 per second
		},
	}

	// Test rate limiting within limits
	for i := 0; i < 5; i++ {
		err := rlc.CheckRateLimit("test-tenant", "enqueue", config)
		assert.NoError(t, err)
	}

	// Test rate limiting exceeding limits
	err := rlc.CheckRateLimit("test-tenant", "enqueue", config)
	assert.Error(t, err)
	assert.True(t, IsRateLimited(err))

	// Test rate limiting disabled
	config.RateLimiting.Enabled = false
	err = rlc.CheckRateLimit("test-tenant", "enqueue", config)
	assert.NoError(t, err)
}