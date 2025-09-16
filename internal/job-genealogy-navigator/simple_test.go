// Copyright 2025 James Ross
package genealogy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test enum parsing functions (no dependencies)
func TestParseViewMode(t *testing.T) {
	tests := []struct {
		input   string
		want    ViewMode
		wantErr bool
	}{
		{"full", ViewModeFull, false},
		{"ancestors", ViewModeAncestors, false},
		{"descendants", ViewModeDescendants, false},
		{"blame", ViewModeBlamePath, false},
		{"impact", ViewModeImpactZone, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseViewMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestParseLayoutMode(t *testing.T) {
	tests := []struct {
		input   string
		want    LayoutMode
		wantErr bool
	}{
		{"topdown", LayoutModeTopDown, false},
		{"timeline", LayoutModeTimeline, false},
		{"radial", LayoutModeRadial, false},
		{"compact", LayoutModeCompact, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseLayoutMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestParseRelationshipType(t *testing.T) {
	tests := []struct {
		input   string
		want    RelationshipType
		wantErr bool
	}{
		{"retry", RelationshipRetry, false},
		{"spawn", RelationshipSpawn, false},
		{"fork", RelationshipFork, false},
		{"callback", RelationshipCallback, false},
		{"compensation", RelationshipCompensation, false},
		{"continuation", RelationshipContinuation, false},
		{"batch_member", RelationshipBatchMember, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseRelationshipType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestParseJobStatus(t *testing.T) {
	tests := []struct {
		input   string
		want    JobStatus
		wantErr bool
	}{
		{"pending", JobStatusPending, false},
		{"processing", JobStatusProcessing, false},
		{"success", JobStatusSuccess, false},
		{"failed", JobStatusFailed, false},
		{"retry", JobStatusRetry, false},
		{"cancelled", JobStatusCancelled, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseJobStatus(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

// Test validation functions
func TestValidationError(t *testing.T) {
	err := NewValidationError("test_field", "invalid_value", "test message")
	assert.Contains(t, err.Error(), "test_field")
	assert.Contains(t, err.Error(), "invalid_value")
	assert.Contains(t, err.Error(), "test message")
}

func TestIsValidViewMode(t *testing.T) {
	assert.True(t, IsValidViewMode(ViewModeFull))
	assert.True(t, IsValidViewMode(ViewModeAncestors))
	assert.True(t, IsValidViewMode(ViewModeDescendants))
	assert.True(t, IsValidViewMode(ViewModeBlamePath))
	assert.True(t, IsValidViewMode(ViewModeImpactZone))
	assert.False(t, IsValidViewMode("invalid"))
}

func TestIsValidLayoutMode(t *testing.T) {
	assert.True(t, IsValidLayoutMode(LayoutModeTopDown))
	assert.True(t, IsValidLayoutMode(LayoutModeTimeline))
	assert.True(t, IsValidLayoutMode(LayoutModeRadial))
	assert.True(t, IsValidLayoutMode(LayoutModeCompact))
	assert.False(t, IsValidLayoutMode("invalid"))
}

func TestIsValidRelationshipType(t *testing.T) {
	assert.True(t, IsValidRelationshipType(RelationshipRetry))
	assert.True(t, IsValidRelationshipType(RelationshipSpawn))
	assert.True(t, IsValidRelationshipType(RelationshipFork))
	assert.True(t, IsValidRelationshipType(RelationshipCallback))
	assert.True(t, IsValidRelationshipType(RelationshipCompensation))
	assert.True(t, IsValidRelationshipType(RelationshipContinuation))
	assert.True(t, IsValidRelationshipType(RelationshipBatchMember))
	assert.False(t, IsValidRelationshipType("invalid"))
}

func TestIsValidJobStatus(t *testing.T) {
	assert.True(t, IsValidJobStatus(JobStatusPending))
	assert.True(t, IsValidJobStatus(JobStatusProcessing))
	assert.True(t, IsValidJobStatus(JobStatusSuccess))
	assert.True(t, IsValidJobStatus(JobStatusFailed))
	assert.True(t, IsValidJobStatus(JobStatusRetry))
	assert.True(t, IsValidJobStatus(JobStatusCancelled))
	assert.False(t, IsValidJobStatus("invalid"))
}

// Test configuration validation
func TestGenealogyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  GenealogyConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultGenealogyConfig(),
			wantErr: false,
		},
		{
			name: "empty redis prefix",
			config: GenealogyConfig{
				RedisKeyPrefix: "",
			},
			wantErr: true,
		},
		{
			name: "zero cache TTL",
			config: GenealogyConfig{
				RedisKeyPrefix: "test",
				CacheTTL:       0,
			},
			wantErr: true,
		},
		{
			name: "excessive max tree size",
			config: GenealogyConfig{
				RedisKeyPrefix: "test",
				CacheTTL:       time.Minute,
				MaxTreeSize:    200000,
			},
			wantErr: true,
		},
		{
			name: "invalid view mode",
			config: GenealogyConfig{
				RedisKeyPrefix:    "test",
				CacheTTL:          time.Minute,
				MaxTreeSize:       1000,
				MaxGenerations:    100,
				DefaultViewMode:   "invalid",
				DefaultLayoutMode: LayoutModeTopDown,
				NodeWidth:         100,
				NodeHeight:        20,
				RefreshInterval:   time.Second,
				RelationshipTTL:   time.Hour,
				TreeCacheTTL:      time.Minute,
				CleanupInterval:   time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenealogyConfig_SetDefaults(t *testing.T) {
	config := GenealogyConfig{}
	config.SetDefaults()

	defaults := DefaultGenealogyConfig()
	assert.Equal(t, defaults.RedisKeyPrefix, config.RedisKeyPrefix)
	assert.Equal(t, defaults.CacheTTL, config.CacheTTL)
	assert.Equal(t, defaults.MaxTreeSize, config.MaxTreeSize)
}

func TestGenealogyConfig_Clone(t *testing.T) {
	original := DefaultGenealogyConfig()
	clone := original.Clone()

	assert.Equal(t, original.RedisKeyPrefix, clone.RedisKeyPrefix)
	assert.Equal(t, original.CacheTTL, clone.CacheTTL)

	// Modify clone to ensure independence
	clone.RedisKeyPrefix = "modified"
	assert.NotEqual(t, original.RedisKeyPrefix, clone.RedisKeyPrefix)
}

// Test error classification
func TestErrorClassification(t *testing.T) {
	validationErr := NewValidationError("field", "value", "message")
	assert.True(t, IsValidationError(validationErr))
	assert.False(t, IsRetryable(validationErr))

	assert.False(t, IsRetryable(ErrJobNotFound))
	assert.False(t, IsRetryable(ErrCyclicRelationship))

	storageErr := NewStorageError("get", "key", "message", assert.AnError)
	assert.True(t, IsStorageError(storageErr))
	assert.True(t, IsRetryable(storageErr))
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{nil, "OK"},
		{ErrJobNotFound, "JOB_NOT_FOUND"},
		{ErrCyclicRelationship, "CYCLIC_RELATIONSHIP"},
		{NewValidationError("field", "value", "message"), "VALIDATION_ERROR"},
		{assert.AnError, "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.code, GetErrorCode(tt.err))
		})
	}
}

// Test cache operations
func TestGenealogyCache_TreeOperations(t *testing.T) {
	cache := NewGenealogyCache(100, 5*time.Minute)

	jobID := "test-job"
	tree := &JobGenealogy{
		RootID:    jobID,
		TotalJobs: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test setting and getting
	cache.SetTree(jobID, tree)
	retrieved := cache.GetTree(jobID)
	assert.NotNil(t, retrieved)
	assert.Equal(t, tree.RootID, retrieved.RootID)

	// Test deletion
	cache.DeleteTree(jobID)
	retrieved = cache.GetTree(jobID)
	assert.Nil(t, retrieved)
}

func TestGenealogyCache_TTLExpiration(t *testing.T) {
	cache := NewGenealogyCache(100, 1*time.Millisecond)

	jobID := "test-job"
	tree := &JobGenealogy{
		RootID:    jobID,
		TotalJobs: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	cache.SetTree(jobID, tree)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	retrieved := cache.GetTree(jobID)
	assert.Nil(t, retrieved)
}

func TestGenealogyCache_Stats(t *testing.T) {
	cache := NewGenealogyCache(100, 5*time.Minute)

	jobID := "test-job"
	tree := &JobGenealogy{
		RootID:    jobID,
		TotalJobs: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	cache.SetTree(jobID, tree)

	relationships := []JobRelationship{
		{
			ParentID:  "parent",
			ChildID:   "child",
			Type:      RelationshipSpawn,
			Timestamp: time.Now(),
		},
	}
	cache.SetRelationships("parent", relationships)

	stats := cache.Stats()
	assert.Equal(t, 1, stats["trees_count"].(int))
	assert.Equal(t, 1, stats["relationships_count"].(int))
	assert.Equal(t, 100, stats["max_size"].(int))
}