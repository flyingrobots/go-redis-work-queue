// Copyright 2025 James Ross
package genealogy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// Mock implementations for testing

type MockGraphStore struct {
	mock.Mock
}

func (m *MockGraphStore) AddRelationship(ctx context.Context, rel JobRelationship) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func (m *MockGraphStore) GetRelationships(ctx context.Context, jobID string) ([]JobRelationship, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]JobRelationship), args.Error(1)
}

func (m *MockGraphStore) GetParents(ctx context.Context, jobID string) ([]string, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGraphStore) GetChildren(ctx context.Context, jobID string) ([]string, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGraphStore) GetAncestors(ctx context.Context, jobID string) ([]string, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGraphStore) GetDescendants(ctx context.Context, jobID string) ([]string, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGraphStore) BuildGenealogy(ctx context.Context, rootID string) (*JobGenealogy, error) {
	args := m.Called(ctx, rootID)
	return args.Get(0).(*JobGenealogy), args.Error(1)
}

func (m *MockGraphStore) RemoveRelationships(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockGraphStore) Cleanup(ctx context.Context, olderThan time.Time) error {
	args := m.Called(ctx, olderThan)
	return args.Error(0)
}

type MockJobProvider struct {
	mock.Mock
}

func (m *MockJobProvider) GetJobDetails(ctx context.Context, jobID string) (*JobDetails, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*JobDetails), args.Error(1)
}

func (m *MockJobProvider) GetBulkJobDetails(ctx context.Context, jobIDs []string) (map[string]*JobDetails, error) {
	args := m.Called(ctx, jobIDs)
	return args.Get(0).(map[string]*JobDetails), args.Error(1)
}

type MockTreeRenderer struct {
	mock.Mock
}

func (m *MockTreeRenderer) RenderTree(genealogy *JobGenealogy, state *NavigationState) (*TreeLayout, error) {
	args := m.Called(genealogy, state)
	return args.Get(0).(*TreeLayout), args.Error(1)
}

// Test data helpers

func createTestJobDetails(id string, status JobStatus) *JobDetails {
	return &JobDetails{
		ID:        id,
		Type:      "test_job",
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  map[string]interface{}{"test": true},
	}
}

func createTestRelationship(parentID, childID string, relType RelationshipType) JobRelationship {
	return JobRelationship{
		ParentID:    parentID,
		ChildID:     childID,
		Type:        relType,
		SpawnReason: "test_reason",
		Timestamp:   time.Now(),
		Metadata:    map[string]interface{}{"test": true},
	}
}

func createTestGenealogy(rootID string) *JobGenealogy {
	nodes := map[string]*JobNode{
		rootID: {
			ID:         rootID,
			ParentID:   "",
			ChildIDs:   []string{"child1", "child2"},
			Generation: 0,
		},
		"child1": {
			ID:         "child1",
			ParentID:   rootID,
			ChildIDs:   []string{},
			Generation: 1,
		},
		"child2": {
			ID:         "child2",
			ParentID:   rootID,
			ChildIDs:   []string{},
			Generation: 1,
		},
	}

	return &JobGenealogy{
		RootID:        rootID,
		Nodes:         nodes,
		Relationships: []JobRelationship{},
		GenerationMap: map[int][]string{
			0: {rootID},
			1: {"child1", "child2"},
		},
		MaxDepth:  1,
		TotalJobs: 3,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Navigator Tests

func TestNewNavigator(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	assert.NotNil(t, nav)
	assert.Equal(t, store, nav.store)
	assert.Equal(t, provider, nav.jobProvider)
	assert.Equal(t, renderer, nav.renderer)
	assert.NotNil(t, nav.cache)
}

func TestNavigator_GetGenealogy(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()
	jobID := "test-job-1"
	expectedGenealogy := createTestGenealogy(jobID)

	// Mock store to return genealogy
	store.On("BuildGenealogy", ctx, jobID).Return(expectedGenealogy, nil)

	// Mock provider to return job details
	jobDetails := map[string]*JobDetails{
		jobID:    createTestJobDetails(jobID, JobStatusSuccess),
		"child1": createTestJobDetails("child1", JobStatusSuccess),
		"child2": createTestJobDetails("child2", JobStatusFailed),
	}
	provider.On("GetBulkJobDetails", ctx, []string{jobID, "child1", "child2"}).Return(jobDetails, nil)

	result, err := nav.GetGenealogy(ctx, jobID)

	require.NoError(t, err)
	assert.Equal(t, expectedGenealogy.RootID, result.RootID)
	assert.Equal(t, expectedGenealogy.TotalJobs, result.TotalJobs)
	assert.Len(t, result.Nodes, 3)

	// Verify job details were populated
	for id, node := range result.Nodes {
		assert.NotNil(t, node.JobDetails)
		assert.Equal(t, id, node.JobDetails.ID)
	}

	store.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestNavigator_GetGenealogy_Cached(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()
	jobID := "test-job-1"
	cachedGenealogy := createTestGenealogy(jobID)

	// Pre-populate cache
	nav.cache.SetTree(jobID, cachedGenealogy)

	result, err := nav.GetGenealogy(ctx, jobID)

	require.NoError(t, err)
	assert.Equal(t, cachedGenealogy.RootID, result.RootID)

	// Store and provider should not be called for cached result
	store.AssertNotCalled(t, "BuildGenealogy")
	provider.AssertNotCalled(t, "GetBulkJobDetails")
}

func TestNavigator_GetImpactAnalysis(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()
	jobID := "test-job-1"

	// Mock descendants
	descendants := []string{"child1", "child2", "grandchild1"}
	store.On("GetDescendants", ctx, jobID).Return(descendants, nil)

	// Mock job details
	jobDetails := map[string]*JobDetails{
		jobID:         createTestJobDetails(jobID, JobStatusFailed),
		"child1":      createTestJobDetails("child1", JobStatusFailed),
		"child2":      createTestJobDetails("child2", JobStatusSuccess),
		"grandchild1": createTestJobDetails("grandchild1", JobStatusFailed),
	}
	allIDs := append([]string{jobID}, descendants...)
	provider.On("GetBulkJobDetails", ctx, allIDs).Return(jobDetails, nil)

	result, err := nav.GetImpactAnalysis(ctx, jobID)

	require.NoError(t, err)
	assert.Equal(t, jobID, result.SourceJobID)
	assert.Len(t, result.AffectedJobs, 4)
	assert.Equal(t, 3, result.FailedJobsCount)
	assert.Equal(t, 1, result.SuccessfulJobsCount)

	store.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestNavigator_GetBlameAnalysis(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()
	failedJobID := "failed-job"

	// Mock ancestors
	ancestors := []string{"root", "parent1", "parent2"}
	store.On("GetAncestors", ctx, failedJobID).Return(ancestors, nil)

	// Mock job details
	jobDetails := map[string]*JobDetails{
		failedJobID: createTestJobDetails(failedJobID, JobStatusFailed),
		"root":      createTestJobDetails("root", JobStatusSuccess),
		"parent1":   createTestJobDetails("parent1", JobStatusFailed),
		"parent2":   createTestJobDetails("parent2", JobStatusSuccess),
	}
	allIDs := append([]string{failedJobID}, ancestors...)
	provider.On("GetBulkJobDetails", ctx, allIDs).Return(jobDetails, nil)

	result, err := nav.GetBlameAnalysis(ctx, failedJobID)

	require.NoError(t, err)
	assert.Equal(t, failedJobID, result.FailedJobID)
	assert.Len(t, result.BlamePath, 4)

	// Find likely root cause (first failed job in ancestry)
	var rootCause *JobDetails
	for _, job := range result.BlamePath {
		if job.Status == JobStatusFailed {
			rootCause = job
			break
		}
	}
	assert.NotNil(t, rootCause)
	assert.Equal(t, result.LikelyRootCause, rootCause)

	store.AssertExpectations(t)
	provider.AssertExpectations(t)
}

func TestNavigator_SetNavigationMode(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	// Test setting navigation mode
	err := nav.SetNavigationMode(ViewModeAncestors, LayoutModeTimeline)
	require.NoError(t, err)

	state := nav.GetNavigationState()
	assert.Equal(t, ViewModeAncestors, state.ViewMode)
	assert.Equal(t, LayoutModeTimeline, state.LayoutMode)
}

func TestNavigator_SetNavigationMode_Invalid(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	// Test invalid view mode
	err := nav.SetNavigationMode("invalid", LayoutModeTopDown)
	assert.Error(t, err)
	assert.True(t, IsValidationError(err))

	// Test invalid layout mode
	err = nav.SetNavigationMode(ViewModeFull, "invalid")
	assert.Error(t, err)
	assert.True(t, IsValidationError(err))
}

func TestNavigator_NavigateToNode(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()
	jobID := "test-job-1"
	genealogy := createTestGenealogy(jobID)

	// Pre-populate cache with genealogy
	nav.cache.SetTree(jobID, genealogy)

	err := nav.NavigateToNode(ctx, "child1")
	require.NoError(t, err)

	state := nav.GetNavigationState()
	assert.Equal(t, "child1", state.FocusedNodeID)
}

func TestNavigator_NavigateToNode_NotFound(t *testing.T) {
	store := &MockGraphStore{}
	provider := &MockJobProvider{}
	renderer := &MockTreeRenderer{}
	config := TestingConfig()
	logger := zaptest.NewLogger(t)

	nav := NewNavigator(store, provider, renderer, config, logger)

	ctx := context.Background()

	err := nav.NavigateToNode(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tree not loaded")
}

// Cache Tests

func TestGenealogyCache_TreeOperations(t *testing.T) {
	cache := NewGenealogyCache(100, 5*time.Minute)

	jobID := "test-job"
	tree := createTestGenealogy(jobID)

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
	tree := createTestGenealogy(jobID)

	cache.SetTree(jobID, tree)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	retrieved := cache.GetTree(jobID)
	assert.Nil(t, retrieved)
}

func TestGenealogyCache_SizeLimit(t *testing.T) {
	cache := NewGenealogyCache(2, 5*time.Minute)

	// Add trees up to limit
	for i := 0; i < 3; i++ {
		jobID := string(rune('a' + i))
		tree := createTestGenealogy(jobID)
		cache.SetTree(jobID, tree)
	}

	stats := cache.Stats()
	assert.Equal(t, 2, stats["trees_count"].(int))
}

func TestGenealogyCache_Stats(t *testing.T) {
	cache := NewGenealogyCache(100, 5*time.Minute)

	jobID := "test-job"
	tree := createTestGenealogy(jobID)
	cache.SetTree(jobID, tree)

	relationships := []JobRelationship{
		createTestRelationship("parent", "child", RelationshipSpawn),
	}
	cache.SetRelationships("parent", relationships)

	stats := cache.Stats()
	assert.Equal(t, 1, stats["trees_count"].(int))
	assert.Equal(t, 1, stats["relationships_count"].(int))
	assert.Equal(t, 100, stats["max_size"].(int))
}

// Configuration Tests

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

// Error Tests

func TestValidationError(t *testing.T) {
	err := NewValidationError("test_field", "invalid_value", "test message")
	assert.Contains(t, err.Error(), "test_field")
	assert.Contains(t, err.Error(), "invalid_value")
	assert.Contains(t, err.Error(), "test message")
}

func TestGenealogyError(t *testing.T) {
	underlyingErr := assert.AnError
	err := NewGenealogyError("test_op", "job-123", "test message", underlyingErr)

	assert.Contains(t, err.Error(), "test_op")
	assert.Contains(t, err.Error(), "job-123")
	assert.Contains(t, err.Error(), "test message")
	assert.Equal(t, underlyingErr, err.Unwrap())
}

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
		assert.Equal(t, tt.code, GetErrorCode(tt.err))
	}
}

// Enum parsing tests

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
		result, err := ParseViewMode(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		}
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
		result, err := ParseLayoutMode(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		}
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
		result, err := ParseRelationshipType(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		}
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
		result, err := ParseJobStatus(tt.input)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		}
	}
}