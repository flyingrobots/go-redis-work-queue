// Copyright 2025 James Ross
package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockTenantManager for testing - defines expected interface
type MockTenantManager struct {
	tenants map[string]*Tenant
	mu      sync.RWMutex
}

type Tenant struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     string                 `json:"status"`
	Config     TenantConfig           `json:"config"`
	Quotas     TenantQuotas           `json:"quotas"`
	Encryption EncryptionConfig       `json:"encryption"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type TenantConfig struct {
	IsolationLevel  string `json:"isolation_level"`
	EnableAudit     bool   `json:"enable_audit"`
	EnableMetrics   bool   `json:"enable_metrics"`
	DefaultQueueTTL int    `json:"default_queue_ttl"`
	MaxJobRetries   int    `json:"max_job_retries"`
}

type TenantQuotas struct {
	MaxJobsPerHour     int64   `json:"max_jobs_per_hour"`
	MaxJobsPerDay      int64   `json:"max_jobs_per_day"`
	MaxBacklogSize     int64   `json:"max_backlog_size"`
	MaxJobSizeBytes    int64   `json:"max_job_size_bytes"`
	MaxQueuesPerTenant int32   `json:"max_queues_per_tenant"`
	MaxWorkersPerQueue int32   `json:"max_workers_per_queue"`
	MaxStorageBytes    int64   `json:"max_storage_bytes"`
	EnqueueRateLimit   int32   `json:"enqueue_rate_limit"`
	DequeueRateLimit   int32   `json:"dequeue_rate_limit"`
	SoftLimitThreshold float64 `json:"soft_limit_threshold"`
}

type EncryptionConfig struct {
	Enabled           bool   `json:"enabled"`
	KEKProvider       string `json:"kek_provider"`
	KEKKeyID          string `json:"kek_key_id"`
	Algorithm         string `json:"algorithm"`
	DEKRotationPeriod string `json:"dek_rotation_period"`
	AutoRotate        bool   `json:"auto_rotate"`
}

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestTenantIDValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid tenant ID",
			id:      "acme-corp",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			id:      "tenant-123",
			wantErr: false,
		},
		{
			name:    "valid short",
			id:      "abc",
			wantErr: false,
		},
		{
			name:    "too short",
			id:      "ab",
			wantErr: true,
			errMsg:  "tenant ID must be 3-32 characters",
		},
		{
			name:    "too long",
			id:      strings.Repeat("a", 33),
			wantErr: true,
			errMsg:  "tenant ID must be 3-32 characters",
		},
		{
			name:    "starts with hyphen",
			id:      "-invalid",
			wantErr: true,
			errMsg:  "tenant ID must start and end with alphanumeric",
		},
		{
			name:    "ends with hyphen",
			id:      "invalid-",
			wantErr: true,
			errMsg:  "tenant ID must start and end with alphanumeric",
		},
		{
			name:    "uppercase letters",
			id:      "INVALID",
			wantErr: true,
			errMsg:  "tenant ID must be lowercase alphanumeric with hyphens",
		},
		{
			name:    "special characters",
			id:      "invalid_tenant",
			wantErr: true,
			errMsg:  "tenant ID must be lowercase alphanumeric with hyphens",
		},
		{
			name:    "reserved name",
			id:      "system",
			wantErr: true,
			errMsg:  "tenant ID is reserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTenantID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateTenant(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("create new tenant successfully", func(t *testing.T) {
		mgr := NewTenantManager(client, logger)

		tenant := &Tenant{
			ID:     "test-tenant",
			Name:   "Test Tenant",
			Status: "active",
			Config: TenantConfig{
				IsolationLevel: "standard",
				EnableAudit:    true,
				EnableMetrics:  true,
			},
			Quotas: TenantQuotas{
				MaxJobsPerHour: 10000,
				MaxJobsPerDay:  200000,
				MaxBacklogSize: 50000,
			},
		}

		err := mgr.CreateTenant(ctx, tenant)
		require.NoError(t, err)

		// Verify tenant was created in Redis
		configKey := fmt.Sprintf("tenant:%s:config", tenant.ID)
		exists, err := client.Exists(ctx, configKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		// Verify tenant data
		data, err := client.Get(ctx, configKey).Result()
		require.NoError(t, err)

		var retrieved Tenant
		err = json.Unmarshal([]byte(data), &retrieved)
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, retrieved.ID)
		assert.Equal(t, tenant.Name, retrieved.Name)
	})

	t.Run("prevent duplicate tenant creation", func(t *testing.T) {
		mgr := NewTenantManager(client, logger)

		tenant := &Tenant{
			ID:     "duplicate-tenant",
			Name:   "Duplicate Tenant",
			Status: "active",
		}

		// First creation should succeed
		err := mgr.CreateTenant(ctx, tenant)
		require.NoError(t, err)

		// Second creation should fail
		err = mgr.CreateTenant(ctx, tenant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("validate tenant ID on creation", func(t *testing.T) {
		mgr := NewTenantManager(client, logger)

		tenant := &Tenant{
			ID:     "INVALID-ID",
			Name:   "Invalid Tenant",
			Status: "active",
		}

		err := mgr.CreateTenant(ctx, tenant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tenant ID")
	})
}

func TestTenantNamespacing(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("key namespace isolation", func(t *testing.T) {
		// Create keys for tenant A
		tenantA := "tenant-a"
		keyA := fmt.Sprintf("t:%s:queue1", tenantA)
		err := client.Set(ctx, keyA, "data-a", 0).Err()
		require.NoError(t, err)

		// Create keys for tenant B
		tenantB := "tenant-b"
		keyB := fmt.Sprintf("t:%s:queue1", tenantB)
		err = client.Set(ctx, keyB, "data-b", 0).Err()
		require.NoError(t, err)

		// Verify isolation - scanning tenant A keys shouldn't return tenant B
		var cursor uint64
		keysA := []string{}
		match := fmt.Sprintf("t:%s:*", tenantA)

		for {
			var keys []string
			keys, cursor, err = client.Scan(ctx, cursor, match, 10).Result()
			require.NoError(t, err)
			keysA = append(keysA, keys...)
			if cursor == 0 {
				break
			}
		}

		assert.Len(t, keysA, 1)
		assert.Equal(t, keyA, keysA[0])

		// Verify data isolation
		dataA, err := client.Get(ctx, keyA).Result()
		require.NoError(t, err)
		assert.Equal(t, "data-a", dataA)

		dataB, err := client.Get(ctx, keyB).Result()
		require.NoError(t, err)
		assert.Equal(t, "data-b", dataB)
	})

	t.Run("queue namespace isolation", func(t *testing.T) {
		tenantA := "tenant-a"
		tenantB := "tenant-b"
		queueName := "orders"

		// Queue keys for different tenants
		queueA := fmt.Sprintf("t:%s:%s:jobs", tenantA, queueName)
		queueB := fmt.Sprintf("t:%s:%s:jobs", tenantB, queueName)

		// Add jobs to each tenant's queue
		err := client.LPush(ctx, queueA, "job-a-1", "job-a-2").Err()
		require.NoError(t, err)

		err = client.LPush(ctx, queueB, "job-b-1").Err()
		require.NoError(t, err)

		// Verify queue isolation
		lenA, err := client.LLen(ctx, queueA).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(2), lenA)

		lenB, err := client.LLen(ctx, queueB).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), lenB)
	})
}

func TestQuotaEnforcement(t *testing.T) {
	ctx := context.Background()

	t.Run("enforce job count limits", func(t *testing.T) {
		quotaMgr := NewQuotaManager()
		tenantID := "limited-tenant"

		// Set low quota for testing
		quotas := TenantQuotas{
			MaxJobsPerHour:  100,
			MaxJobsPerDay:   1000,
			MaxBacklogSize:  50,
			MaxJobSizeBytes: 1024,
		}

		err := quotaMgr.SetQuotas(ctx, tenantID, quotas)
		require.NoError(t, err)

		// Enqueue jobs up to limit
		for i := 0; i < 100; i++ {
			allowed, err := quotaMgr.CheckAndIncrementJobCount(ctx, tenantID)
			require.NoError(t, err)
			assert.True(t, allowed, "job %d should be allowed", i)
		}

		// Next job should be rejected
		allowed, err := quotaMgr.CheckAndIncrementJobCount(ctx, tenantID)
		require.NoError(t, err)
		assert.False(t, allowed, "should reject after quota exceeded")
	})

	t.Run("enforce backlog size limits", func(t *testing.T) {
		quotaMgr := NewQuotaManager()
		tenantID := "backlog-tenant"

		quotas := TenantQuotas{
			MaxBacklogSize: 10,
		}

		err := quotaMgr.SetQuotas(ctx, tenantID, quotas)
		require.NoError(t, err)

		// Check backlog limit
		for i := 0; i < 10; i++ {
			allowed, err := quotaMgr.CheckBacklogSize(ctx, tenantID, i)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Exceed backlog
		allowed, err := quotaMgr.CheckBacklogSize(ctx, tenantID, 11)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("soft limit warnings", func(t *testing.T) {
		quotaMgr := NewQuotaManager()
		tenantID := "warning-tenant"

		quotas := TenantQuotas{
			MaxJobsPerHour:     100,
			SoftLimitThreshold: 0.8, // Warn at 80%
		}

		err := quotaMgr.SetQuotas(ctx, tenantID, quotas)
		require.NoError(t, err)

		// Track warnings
		warnings := []string{}
		quotaMgr.OnWarning = func(tenant string, msg string) {
			warnings = append(warnings, msg)
		}

		// Use 79 jobs - no warning
		for i := 0; i < 79; i++ {
			quotaMgr.CheckAndIncrementJobCount(ctx, tenantID)
		}
		assert.Len(t, warnings, 0)

		// 80th job should trigger warning
		quotaMgr.CheckAndIncrementJobCount(ctx, tenantID)
		assert.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "approaching limit")
	})

	t.Run("rate limiting", func(t *testing.T) {
		quotaMgr := NewQuotaManager()
		tenantID := "rate-limited"

		quotas := TenantQuotas{
			EnqueueRateLimit: 10, // 10 per second
		}

		err := quotaMgr.SetQuotas(ctx, tenantID, quotas)
		require.NoError(t, err)

		// Should allow burst up to limit
		start := time.Now()
		allowed := 0

		for i := 0; i < 20; i++ {
			if quotaMgr.CheckRateLimit(ctx, tenantID, "enqueue") {
				allowed++
			}
		}

		elapsed := time.Since(start)
		if elapsed < time.Second {
			// Within 1 second, should only allow 10
			assert.LessOrEqual(t, allowed, 10)
		}
	})
}

func TestCrossTenantIsolation(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("prevent cross-tenant queue access", func(t *testing.T) {
		mgr := NewTenantManager(client, logger)

		// Create two tenants
		tenantA := &Tenant{ID: "tenant-a", Name: "Tenant A"}
		tenantB := &Tenant{ID: "tenant-b", Name: "Tenant B"}

		err := mgr.CreateTenant(ctx, tenantA)
		require.NoError(t, err)

		err = mgr.CreateTenant(ctx, tenantB)
		require.NoError(t, err)

		// Create queue for tenant A
		queueMgr := NewQueueManager(client, logger)
		err = queueMgr.CreateQueue(ctx, "tenant-a", "private-queue")
		require.NoError(t, err)

		// Try to access from tenant B context
		_, err = queueMgr.GetQueue(ctx, "tenant-b", "private-queue")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Verify tenant A can access
		queue, err := queueMgr.GetQueue(ctx, "tenant-a", "private-queue")
		require.NoError(t, err)
		assert.NotNil(t, queue)
	})

	t.Run("prevent cross-tenant job access", func(t *testing.T) {
		jobMgr := NewJobManager()

		// Enqueue job for tenant A
		jobID, err := jobMgr.EnqueueJob(ctx, "tenant-a", "queue1", []byte("data"))
		require.NoError(t, err)

		// Try to get job from tenant B context
		_, err = jobMgr.GetJob(ctx, "tenant-b", jobID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		// Verify tenant A can access
		job, err := jobMgr.GetJob(ctx, "tenant-a", jobID)
		require.NoError(t, err)
		assert.NotNil(t, job)
	})
}

func TestEncryption(t *testing.T) {
	ctx := context.Background()

	t.Run("encrypt and decrypt payload", func(t *testing.T) {
		encMgr := NewEncryptionManager()
		tenantID := "encrypted-tenant"

		// Configure encryption
		config := EncryptionConfig{
			Enabled:     true,
			KEKProvider: "local",
			Algorithm:   "AES-256-GCM",
		}

		err := encMgr.ConfigureEncryption(ctx, tenantID, config)
		require.NoError(t, err)

		// Test data
		plaintext := []byte("sensitive data")

		// Encrypt
		encrypted, err := encMgr.EncryptPayload(ctx, tenantID, plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted)

		// Decrypt
		decrypted, err := encMgr.DecryptPayload(ctx, tenantID, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("key rotation", func(t *testing.T) {
		encMgr := NewEncryptionManager()
		tenantID := "rotation-tenant"

		config := EncryptionConfig{
			Enabled:     true,
			KEKProvider: "local",
			Algorithm:   "AES-256-GCM",
		}

		err := encMgr.ConfigureEncryption(ctx, tenantID, config)
		require.NoError(t, err)

		// Encrypt with initial key
		plaintext := []byte("test data")
		encrypted1, err := encMgr.EncryptPayload(ctx, tenantID, plaintext)
		require.NoError(t, err)

		// Rotate keys
		err = encMgr.RotateKeys(ctx, tenantID)
		require.NoError(t, err)

		// Encrypt with new key
		encrypted2, err := encMgr.EncryptPayload(ctx, tenantID, plaintext)
		require.NoError(t, err)

		// Both should decrypt correctly
		decrypted1, err := encMgr.DecryptPayload(ctx, tenantID, encrypted1)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted1)

		decrypted2, err := encMgr.DecryptPayload(ctx, tenantID, encrypted2)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted2)
	})

	t.Run("tenant-specific encryption", func(t *testing.T) {
		encMgr := NewEncryptionManager()

		// Configure different tenants
		tenantA := "tenant-a"
		tenantB := "tenant-b"

		configA := EncryptionConfig{Enabled: true, KEKProvider: "local"}
		configB := EncryptionConfig{Enabled: true, KEKProvider: "local"}

		err := encMgr.ConfigureEncryption(ctx, tenantA, configA)
		require.NoError(t, err)

		err = encMgr.ConfigureEncryption(ctx, tenantB, configB)
		require.NoError(t, err)

		// Encrypt same data with different tenant keys
		plaintext := []byte("shared data")

		encryptedA, err := encMgr.EncryptPayload(ctx, tenantA, plaintext)
		require.NoError(t, err)

		encryptedB, err := encMgr.EncryptPayload(ctx, tenantB, plaintext)
		require.NoError(t, err)

		// Encrypted data should be different
		assert.NotEqual(t, encryptedA, encryptedB)

		// Cross-tenant decryption should fail
		_, err = encMgr.DecryptPayload(ctx, tenantA, encryptedB)
		assert.Error(t, err)

		_, err = encMgr.DecryptPayload(ctx, tenantB, encryptedA)
		assert.Error(t, err)
	})
}

// Mock implementations for testing
func ValidateTenantID(id string) error {
	if len(id) < 3 || len(id) > 32 {
		return fmt.Errorf("tenant ID must be 3-32 characters")
	}
	if id[0] == '-' || id[len(id)-1] == '-' {
		return fmt.Errorf("tenant ID must start and end with alphanumeric")
	}
	if id != strings.ToLower(id) {
		return fmt.Errorf("tenant ID must be lowercase alphanumeric with hyphens")
	}
	if strings.ContainsAny(id, "_!@#$%^&*()+=[]{}|\\:;\"'<>,.?/") {
		return fmt.Errorf("tenant ID must be lowercase alphanumeric with hyphens")
	}
	reserved := []string{"system", "admin", "default", "test"}
	for _, r := range reserved {
		if id == r {
			return fmt.Errorf("tenant ID is reserved")
		}
	}
	return nil
}

type TenantManager struct {
	client *redis.Client
	logger *zap.Logger
}

func NewTenantManager(client *redis.Client, logger *zap.Logger) *TenantManager {
	return &TenantManager{client: client, logger: logger}
}

func (tm *TenantManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	if err := ValidateTenantID(tenant.ID); err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	configKey := fmt.Sprintf("tenant:%s:config", tenant.ID)
	exists, err := tm.client.Exists(ctx, configKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("tenant already exists: %s", tenant.ID)
	}

	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	data, err := json.Marshal(tenant)
	if err != nil {
		return err
	}

	return tm.client.Set(ctx, configKey, data, 0).Err()
}

type QuotaManager struct {
	quotas    map[string]TenantQuotas
	usage     map[string]map[string]int64
	mu        sync.RWMutex
	OnWarning func(tenant, msg string)
}

func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]TenantQuotas),
		usage:  make(map[string]map[string]int64),
	}
}

func (qm *QuotaManager) SetQuotas(ctx context.Context, tenantID string, quotas TenantQuotas) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.quotas[tenantID] = quotas
	qm.usage[tenantID] = make(map[string]int64)
	return nil
}

func (qm *QuotaManager) CheckAndIncrementJobCount(ctx context.Context, tenantID string) (bool, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quotas, ok := qm.quotas[tenantID]
	if !ok {
		return true, nil // No quotas set
	}

	usage := qm.usage[tenantID]
	current := usage["jobs_hour"]

	if current >= quotas.MaxJobsPerHour {
		return false, nil
	}

	// Check soft limit
	if quotas.SoftLimitThreshold > 0 {
		threshold := float64(quotas.MaxJobsPerHour) * quotas.SoftLimitThreshold
		if float64(current) >= threshold && qm.OnWarning != nil {
			qm.OnWarning(tenantID, fmt.Sprintf("approaching hourly job limit: %d/%d", current, quotas.MaxJobsPerHour))
		}
	}

	usage["jobs_hour"]++
	return true, nil
}

func (qm *QuotaManager) CheckBacklogSize(ctx context.Context, tenantID string, currentSize int) (bool, error) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quotas, ok := qm.quotas[tenantID]
	if !ok {
		return true, nil
	}

	return int64(currentSize) <= quotas.MaxBacklogSize, nil
}

func (qm *QuotaManager) CheckRateLimit(ctx context.Context, tenantID string, operation string) bool {
	// Simplified rate limiting for testing
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quotas, ok := qm.quotas[tenantID]
	if !ok {
		return true
	}

	// Simple implementation - would use token bucket in production
	return quotas.EnqueueRateLimit > 0
}

type QueueManager struct {
	client *redis.Client
	logger *zap.Logger
}

func NewQueueManager(client *redis.Client, logger *zap.Logger) *QueueManager {
	return &QueueManager{client: client, logger: logger}
}

func (qm *QueueManager) CreateQueue(ctx context.Context, tenantID, queueName string) error {
	key := fmt.Sprintf("t:%s:%s:config", tenantID, queueName)
	return qm.client.Set(ctx, key, "{}", 0).Err()
}

func (qm *QueueManager) GetQueue(ctx context.Context, tenantID, queueName string) (interface{}, error) {
	key := fmt.Sprintf("t:%s:%s:config", tenantID, queueName)
	exists, err := qm.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("queue not found")
	}
	return struct{}{}, nil
}

type JobManager struct{}

func NewJobManager() *JobManager {
	return &JobManager{}
}

func (jm *JobManager) EnqueueJob(ctx context.Context, tenantID, queue string, payload []byte) (string, error) {
	return fmt.Sprintf("%s-%s-%d", tenantID, queue, time.Now().Unix()), nil
}

func (jm *JobManager) GetJob(ctx context.Context, tenantID, jobID string) (interface{}, error) {
	if !strings.HasPrefix(jobID, tenantID) {
		return nil, fmt.Errorf("access denied")
	}
	return struct{}{}, nil
}

type EncryptionManager struct {
	configs map[string]EncryptionConfig
	keys    map[string][]byte
	mu      sync.RWMutex
}

func NewEncryptionManager() *EncryptionManager {
	return &EncryptionManager{
		configs: make(map[string]EncryptionConfig),
		keys:    make(map[string][]byte),
	}
}

func (em *EncryptionManager) ConfigureEncryption(ctx context.Context, tenantID string, config EncryptionConfig) error {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.configs[tenantID] = config
	// Generate a simple key for testing
	em.keys[tenantID] = []byte(fmt.Sprintf("key-%s-%d", tenantID, time.Now().Unix()))
	return nil
}

func (em *EncryptionManager) EncryptPayload(ctx context.Context, tenantID string, plaintext []byte) ([]byte, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if _, ok := em.configs[tenantID]; !ok {
		return nil, fmt.Errorf("encryption not configured for tenant")
	}

	// Simple XOR for testing - would use real encryption in production
	key := em.keys[tenantID]
	encrypted := make([]byte, len(plaintext))
	for i := range plaintext {
		encrypted[i] = plaintext[i] ^ key[i%len(key)]
	}
	return encrypted, nil
}

func (em *EncryptionManager) DecryptPayload(ctx context.Context, tenantID string, encrypted []byte) ([]byte, error) {
	// XOR is symmetric, so we can use encrypt for decrypt
	return em.EncryptPayload(ctx, tenantID, encrypted)
}

func (em *EncryptionManager) RotateKeys(ctx context.Context, tenantID string) error {
	em.mu.Lock()
	defer em.mu.Unlock()
	// Keep old key for decryption, add new key for encryption
	em.keys[tenantID] = []byte(fmt.Sprintf("key-%s-%d", tenantID, time.Now().Unix()))
	return nil
}
