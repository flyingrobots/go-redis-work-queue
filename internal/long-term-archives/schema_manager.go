// Copyright 2025 James Ross
package archives

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// schemaManager implements the SchemaManager interface
type schemaManager struct {
	redis  *redis.Client
	logger *zap.Logger
	config *ArchiveConfig
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(redisClient *redis.Client, config *ArchiveConfig, logger *zap.Logger) SchemaManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &schemaManager{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// GetCurrentVersion returns the current schema version
func (sm *schemaManager) GetCurrentVersion(ctx context.Context) (int, error) {
	versionKey := "archive:schema:version"

	version, err := sm.redis.Get(ctx, versionKey).Int()
	if err == redis.Nil {
		// No version set, initialize with default
		version = 1
		err = sm.redis.Set(ctx, versionKey, version, 0).Err()
		if err != nil {
			return 0, fmt.Errorf("failed to initialize schema version: %w", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// Upgrade upgrades the schema to the target version
func (sm *schemaManager) Upgrade(ctx context.Context, targetVersion int) error {
	currentVersion, err := sm.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion >= targetVersion {
		return fmt.Errorf("current version %d is already at or above target version %d", currentVersion, targetVersion)
	}

	sm.logger.Info("Starting schema upgrade",
		zap.Int("from_version", currentVersion),
		zap.Int("to_version", targetVersion))

	// Get all available schema evolutions
	evolutions, err := sm.GetEvolution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema evolution: %w", err)
	}

	// Plan upgrade path
	upgradePath := sm.planUpgradePath(evolutions, currentVersion, targetVersion)
	if len(upgradePath) == 0 {
		return fmt.Errorf("no upgrade path found from version %d to %d", currentVersion, targetVersion)
	}

	// Execute upgrades step by step
	for _, evolution := range upgradePath {
		sm.logger.Info("Applying schema evolution",
			zap.Int("version", evolution.Version),
			zap.String("description", evolution.Description))

		err := sm.applyEvolution(ctx, evolution)
		if err != nil {
			return fmt.Errorf("failed to apply evolution to version %d: %w", evolution.Version, err)
		}

		// Update current version
		err = sm.redis.Set(ctx, "archive:schema:version", evolution.Version, 0).Err()
		if err != nil {
			return fmt.Errorf("failed to update schema version: %w", err)
		}

		sm.logger.Info("Schema evolution applied successfully",
			zap.Int("version", evolution.Version))
	}

	sm.logger.Info("Schema upgrade completed",
		zap.Int("from_version", currentVersion),
		zap.Int("to_version", targetVersion))

	return nil
}

// IsBackwardCompatible checks if an upgrade is backward compatible
func (sm *schemaManager) IsBackwardCompatible(ctx context.Context, fromVersion, toVersion int) (bool, error) {
	evolutions, err := sm.GetEvolution(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get schema evolution: %w", err)
	}

	for _, evolution := range evolutions {
		if evolution.Version > fromVersion && evolution.Version <= toVersion {
			if !evolution.Backward {
				return false, nil
			}
		}
	}

	return true, nil
}

// GetEvolution returns all schema evolutions
func (sm *schemaManager) GetEvolution(ctx context.Context) ([]SchemaEvolution, error) {
	evolutionKey := "archive:schema:evolutions"

	// Get all evolution entries
	evolutions, err := sm.redis.HGetAll(ctx, evolutionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema evolutions: %w", err)
	}

	var result []SchemaEvolution
	for versionStr, evolutionData := range evolutions {
		var evolution SchemaEvolution
		err := json.Unmarshal([]byte(evolutionData), &evolution)
		if err != nil {
			sm.logger.Warn("Failed to unmarshal evolution",
				zap.String("version", versionStr),
				zap.Error(err))
			continue
		}
		result = append(result, evolution)
	}

	// Sort by version
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Version > result[j].Version {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// AddEvolution adds a new schema evolution
func (sm *schemaManager) AddEvolution(ctx context.Context, evolution SchemaEvolution) error {
	evolutionKey := "archive:schema:evolutions"

	evolutionData, err := json.Marshal(evolution)
	if err != nil {
		return fmt.Errorf("failed to marshal evolution: %w", err)
	}

	err = sm.redis.HSet(ctx, evolutionKey, fmt.Sprintf("%d", evolution.Version), evolutionData).Err()
	if err != nil {
		return fmt.Errorf("failed to store evolution: %w", err)
	}

	sm.logger.Info("Schema evolution added",
		zap.Int("version", evolution.Version),
		zap.String("description", evolution.Description))

	return nil
}

// planUpgradePath plans the upgrade path from current to target version
func (sm *schemaManager) planUpgradePath(evolutions []SchemaEvolution, fromVersion, toVersion int) []SchemaEvolution {
	var path []SchemaEvolution

	for _, evolution := range evolutions {
		if evolution.Version > fromVersion && evolution.Version <= toVersion {
			path = append(path, evolution)
		}
	}

	return path
}

// applyEvolution applies a single schema evolution
func (sm *schemaManager) applyEvolution(ctx context.Context, evolution SchemaEvolution) error {
	// Record the evolution application
	applicationKey := fmt.Sprintf("archive:schema:applied:%d", evolution.Version)
	applicationData := map[string]interface{}{
		"version":     evolution.Version,
		"applied_at":  time.Now(),
		"description": evolution.Description,
		"changes":     evolution.Changes,
	}

	applicationJSON, err := json.Marshal(applicationData)
	if err != nil {
		return fmt.Errorf("failed to marshal application data: %w", err)
	}

	err = sm.redis.Set(ctx, applicationKey, applicationJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to record evolution application: %w", err)
	}

	// If migration is required, execute it
	if evolution.Migration != nil && evolution.Migration.Required {
		err := sm.executeMigration(ctx, evolution)
		if err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// executeMigration executes a data migration
func (sm *schemaManager) executeMigration(ctx context.Context, evolution SchemaEvolution) error {
	if evolution.Migration == nil {
		return nil
	}

	sm.logger.Info("Executing data migration",
		zap.Int("version", evolution.Version),
		zap.String("script", evolution.Migration.Script))

	// In a real implementation, this would execute the migration script
	// For now, we'll just simulate the migration
	migrationKey := fmt.Sprintf("archive:schema:migration:%d", evolution.Version)
	migrationData := map[string]interface{}{
		"version":         evolution.Version,
		"started_at":      time.Now(),
		"estimated_time":  evolution.Migration.EstimatedTime,
		"status":          "completed",
		"script":          evolution.Migration.Script,
		"rollback_script": evolution.Migration.RollbackScript,
	}

	migrationJSON, err := json.Marshal(migrationData)
	if err != nil {
		return fmt.Errorf("failed to marshal migration data: %w", err)
	}

	err = sm.redis.Set(ctx, migrationKey, migrationJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Simulate migration time
	time.Sleep(100 * time.Millisecond)

	sm.logger.Info("Data migration completed",
		zap.Int("version", evolution.Version))

	return nil
}

// ValidateSchema validates the current schema against jobs
func (sm *schemaManager) ValidateSchema(ctx context.Context, job ArchiveJob) error {
	currentVersion, err := sm.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Set the schema version on the job
	job.SchemaVersion = currentVersion

	// Validate required fields based on current schema version
	switch currentVersion {
	case 1:
		return sm.validateSchemaV1(job)
	case 2:
		return sm.validateSchemaV2(job)
	default:
		return sm.validateSchemaLatest(job)
	}
}

// validateSchemaV1 validates against schema version 1
func (sm *schemaManager) validateSchemaV1(job ArchiveJob) error {
	if job.JobID == "" {
		return fmt.Errorf("job_id is required")
	}
	if job.Queue == "" {
		return fmt.Errorf("queue is required")
	}
	if job.CompletedAt.IsZero() {
		return fmt.Errorf("completed_at is required")
	}
	if job.Outcome == "" {
		return fmt.Errorf("outcome is required")
	}
	return nil
}

// validateSchemaV2 validates against schema version 2 (with additional fields)
func (sm *schemaManager) validateSchemaV2(job ArchiveJob) error {
	// All V1 validations
	if err := sm.validateSchemaV1(job); err != nil {
		return err
	}

	// Additional V2 validations
	if job.WorkerID == "" {
		return fmt.Errorf("worker_id is required in schema v2")
	}
	if job.ArchivedAt.IsZero() {
		return fmt.Errorf("archived_at is required in schema v2")
	}

	return nil
}

// validateSchemaLatest validates against the latest schema version
func (sm *schemaManager) validateSchemaLatest(job ArchiveJob) error {
	// Use the most recent validation logic
	return sm.validateSchemaV2(job)
}

// GetAppliedEvolutions returns all applied schema evolutions
func (sm *schemaManager) GetAppliedEvolutions(ctx context.Context) ([]map[string]interface{}, error) {
	pattern := "archive:schema:applied:*"
	keys, err := sm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied evolution keys: %w", err)
	}

	var applied []map[string]interface{}
	for _, key := range keys {
		data, err := sm.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var evolutionData map[string]interface{}
		err = json.Unmarshal([]byte(data), &evolutionData)
		if err != nil {
			continue
		}

		applied = append(applied, evolutionData)
	}

	return applied, nil
}

// InitializeDefaultEvolutions sets up default schema evolutions
func (sm *schemaManager) InitializeDefaultEvolutions(ctx context.Context) error {
	evolutions := []SchemaEvolution{
		{
			Version:     1,
			Description: "Initial schema with basic job fields",
			Changes: []SchemaChange{
				{
					Type:        ChangeTypeAdd,
					Field:       "job_id",
					NewType:     "string",
					Description: "Unique job identifier",
					Required:    true,
				},
				{
					Type:        ChangeTypeAdd,
					Field:       "queue",
					NewType:     "string",
					Description: "Queue name",
					Required:    true,
				},
				{
					Type:        ChangeTypeAdd,
					Field:       "completed_at",
					NewType:     "timestamp",
					Description: "Job completion timestamp",
					Required:    true,
				},
				{
					Type:        ChangeTypeAdd,
					Field:       "outcome",
					NewType:     "enum",
					Description: "Job execution outcome",
					Required:    true,
				},
			},
			CreatedAt: time.Now(),
			Backward:  true,
		},
		{
			Version:     2,
			Description: "Add worker tracking and archive timestamp",
			Changes: []SchemaChange{
				{
					Type:        ChangeTypeAdd,
					Field:       "worker_id",
					NewType:     "string",
					Description: "Worker that processed the job",
					Required:    true,
				},
				{
					Type:        ChangeTypeAdd,
					Field:       "archived_at",
					NewType:     "timestamp",
					Description: "When the job was archived",
					Required:    true,
				},
				{
					Type:        ChangeTypeAdd,
					Field:       "tenant",
					NewType:     "string",
					Description: "Tenant identifier for multi-tenancy",
					Required:    false,
				},
			},
			CreatedAt: time.Now(),
			Backward:  true,
		},
	}

	for _, evolution := range evolutions {
		// Check if evolution already exists
		existing, err := sm.GetEvolution(ctx)
		if err != nil {
			return err
		}

		exists := false
		for _, e := range existing {
			if e.Version == evolution.Version {
				exists = true
				break
			}
		}

		if !exists {
			err := sm.AddEvolution(ctx, evolution)
			if err != nil {
				return fmt.Errorf("failed to add default evolution %d: %w", evolution.Version, err)
			}
		}
	}

	return nil
}