// Copyright 2025 James Ross
package genealogy

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisGraphStore implements GraphStore using Redis as the backend
type RedisGraphStore struct {
	client redis.Cmdable
	config GenealogyConfig
	logger *zap.Logger
}

// NewRedisGraphStore creates a new Redis-backed graph store
func NewRedisGraphStore(client redis.Cmdable, config GenealogyConfig, logger *zap.Logger) *RedisGraphStore {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RedisGraphStore{
		client: client,
		config: config,
		logger: logger,
	}
}

// AddRelationship stores a parent-child relationship
func (r *RedisGraphStore) AddRelationship(ctx context.Context, rel JobRelationship) error {
	pipe := r.client.Pipeline()

	// Store in parent->children index
	childrenKey := r.getChildrenKey(rel.ParentID)
	timestamp := float64(rel.Timestamp.Unix())
	pipe.ZAdd(ctx, childrenKey, &redis.Z{
		Score:  timestamp,
		Member: rel.ChildID,
	})

	// Store in child->parents index
	parentsKey := r.getParentsKey(rel.ChildID)
	pipe.ZAdd(ctx, parentsKey, &redis.Z{
		Score:  timestamp,
		Member: rel.ParentID,
	})

	// Store relationship metadata
	relationKey := r.getRelationshipKey(rel.ParentID, rel.ChildID)
	relationData := map[string]interface{}{
		"type":         string(rel.Type),
		"spawn_reason": rel.SpawnReason,
		"timestamp":    rel.Timestamp.Unix(),
	}

	if rel.Metadata != nil {
		metadataJSON, err := json.Marshal(rel.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal relationship metadata: %w", err)
		}
		relationData["metadata"] = string(metadataJSON)
	}

	pipe.HMSet(ctx, relationKey, relationData)

	// Set TTLs
	if r.config.RelationshipTTL > 0 {
		pipe.Expire(ctx, childrenKey, r.config.RelationshipTTL)
		pipe.Expire(ctx, parentsKey, r.config.RelationshipTTL)
		pipe.Expire(ctx, relationKey, r.config.RelationshipTTL)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add relationship %s->%s: %w", rel.ParentID, rel.ChildID, err)
	}

	r.logger.Debug("Added job relationship",
		zap.String("parent_id", rel.ParentID),
		zap.String("child_id", rel.ChildID),
		zap.String("type", string(rel.Type)))

	return nil
}

// GetRelationships retrieves all relationships for a job (both as parent and child)
func (r *RedisGraphStore) GetRelationships(ctx context.Context, jobID string) ([]JobRelationship, error) {
	relationships := make([]JobRelationship, 0)

	// Get relationships where job is parent
	parentRels, err := r.getRelationshipsAsParent(ctx, jobID)
	if err != nil {
		return nil, err
	}
	relationships = append(relationships, parentRels...)

	// Get relationships where job is child
	childRels, err := r.getRelationshipsAsChild(ctx, jobID)
	if err != nil {
		return nil, err
	}
	relationships = append(relationships, childRels...)

	return relationships, nil
}

// GetParents retrieves direct parents of a job
func (r *RedisGraphStore) GetParents(ctx context.Context, jobID string) ([]string, error) {
	parentsKey := r.getParentsKey(jobID)

	members, err := r.client.ZRevRange(ctx, parentsKey, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get parents for job %s: %w", jobID, err)
	}

	return members, nil
}

// GetChildren retrieves direct children of a job
func (r *RedisGraphStore) GetChildren(ctx context.Context, jobID string) ([]string, error) {
	childrenKey := r.getChildrenKey(jobID)

	members, err := r.client.ZRange(ctx, childrenKey, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get children for job %s: %w", jobID, err)
	}

	return members, nil
}

// GetAncestors retrieves all ancestors of a job (recursive parents)
func (r *RedisGraphStore) GetAncestors(ctx context.Context, jobID string) ([]string, error) {
	ancestors := make([]string, 0)
	visited := make(map[string]bool)
	queue := []string{jobID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		parents, err := r.GetParents(ctx, currentID)
		if err != nil {
			return nil, err
		}

		for _, parentID := range parents {
			if !visited[parentID] {
				ancestors = append(ancestors, parentID)
				queue = append(queue, parentID)
			}
		}
	}

	return ancestors, nil
}

// GetDescendants retrieves all descendants of a job (recursive children)
func (r *RedisGraphStore) GetDescendants(ctx context.Context, jobID string) ([]string, error) {
	descendants := make([]string, 0)
	visited := make(map[string]bool)
	queue := []string{jobID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		children, err := r.GetChildren(ctx, currentID)
		if err != nil {
			return nil, err
		}

		for _, childID := range children {
			if !visited[childID] {
				descendants = append(descendants, childID)
				queue = append(queue, childID)
			}
		}
	}

	return descendants, nil
}

// BuildGenealogy constructs a complete family tree for a job
func (r *RedisGraphStore) BuildGenealogy(ctx context.Context, rootID string) (*JobGenealogy, error) {
	// Find the true root by traversing ancestors
	trueRootID, err := r.findTrueRoot(ctx, rootID)
	if err != nil {
		return nil, err
	}

	// Get all descendants from true root
	descendants, err := r.GetDescendants(ctx, trueRootID)
	if err != nil {
		return nil, err
	}

	// Include the root itself
	allJobIDs := append([]string{trueRootID}, descendants...)

	// Build nodes map (job details will be populated by caller)
	nodes := make(map[string]*JobNode)
	for _, jobID := range allJobIDs {
		nodes[jobID] = &JobNode{
			ID:       jobID,
			ChildIDs: make([]string, 0),
		}
	}

	// Get all relationships for these jobs
	relationships := make([]JobRelationship, 0)
	for _, jobID := range allJobIDs {
		jobRels, err := r.GetRelationships(ctx, jobID)
		if err != nil {
			r.logger.Warn("Failed to get relationships for job",
				zap.String("job_id", jobID),
				zap.Error(err))
			continue
		}
		relationships = append(relationships, jobRels...)
	}

	// Build parent-child structure and generation map
	generationMap := make(map[int][]string)
	maxDepth := 0

	// Set generations using BFS
	visited := make(map[string]bool)
	queue := []struct {
		jobID      string
		generation int
	}{{trueRootID, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.jobID] {
			continue
		}
		visited[current.jobID] = true

		// Update node generation
		if node, exists := nodes[current.jobID]; exists {
			node.Generation = current.generation
		}

		// Add to generation map
		if generationMap[current.generation] == nil {
			generationMap[current.generation] = make([]string, 0)
		}
		generationMap[current.generation] = append(generationMap[current.generation], current.jobID)

		if current.generation > maxDepth {
			maxDepth = current.generation
		}

		// Add children to queue
		children, err := r.GetChildren(ctx, current.jobID)
		if err != nil {
			continue
		}

		if node, exists := nodes[current.jobID]; exists {
			node.ChildIDs = children
		}

		for _, childID := range children {
			if !visited[childID] {
				queue = append(queue, struct {
					jobID      string
					generation int
				}{childID, current.generation + 1})

				// Set parent reference
				if childNode, exists := nodes[childID]; exists {
					childNode.ParentID = current.jobID
				}
			}
		}
	}

	genealogy := &JobGenealogy{
		RootID:        trueRootID,
		Nodes:         nodes,
		Relationships: r.deduplicateRelationships(relationships),
		GenerationMap: generationMap,
		MaxDepth:      maxDepth,
		TotalJobs:     len(nodes),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	r.logger.Debug("Built genealogy",
		zap.String("root_id", trueRootID),
		zap.Int("total_jobs", genealogy.TotalJobs),
		zap.Int("max_depth", maxDepth),
		zap.Int("relationships", len(genealogy.Relationships)))

	return genealogy, nil
}

// RemoveRelationships removes all relationships for a job
func (r *RedisGraphStore) RemoveRelationships(ctx context.Context, jobID string) error {
	pipe := r.client.Pipeline()

	// Get all relationships to remove metadata
	relationships, err := r.GetRelationships(ctx, jobID)
	if err != nil {
		return err
	}

	// Remove relationship metadata
	for _, rel := range relationships {
		relationKey := r.getRelationshipKey(rel.ParentID, rel.ChildID)
		pipe.Del(ctx, relationKey)
	}

	// Remove from children indexes
	childrenKey := r.getChildrenKey(jobID)
	pipe.Del(ctx, childrenKey)

	// Remove from parents indexes
	parentsKey := r.getParentsKey(jobID)
	pipe.Del(ctx, parentsKey)

	// Remove job from other jobs' children sets
	for _, rel := range relationships {
		if rel.ParentID == jobID {
			// Remove from parent's children set
			otherChildrenKey := r.getChildrenKey(rel.ChildID)
			pipe.ZRem(ctx, otherChildrenKey, jobID)
		} else if rel.ChildID == jobID {
			// Remove from child's parents set
			otherParentsKey := r.getParentsKey(rel.ParentID)
			pipe.ZRem(ctx, otherParentsKey, jobID)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove relationships for job %s: %w", jobID, err)
	}

	r.logger.Debug("Removed job relationships", zap.String("job_id", jobID))
	return nil
}

// Cleanup removes relationships older than the specified time
func (r *RedisGraphStore) Cleanup(ctx context.Context, olderThan time.Time) error {
	// This is a simplified cleanup - in production you'd want more sophisticated cleanup
	// that scans for expired relationship keys

	cutoffScore := float64(olderThan.Unix())

	// Use SCAN to find all relationship keys
	iter := r.client.Scan(ctx, 0, r.config.RedisKeyPrefix+":*", 100).Iterator()
	deletedCount := 0

	for iter.Next(ctx) {
		key := iter.Val()

		// Check if it's a sorted set (children or parents index)
		if r.client.Type(ctx, key).Val() == "zset" {
			// Remove members older than cutoff
			removed, err := r.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("(%f", cutoffScore)).Result()
			if err != nil {
				r.logger.Warn("Failed to cleanup key", zap.String("key", key), zap.Error(err))
				continue
			}
			deletedCount += int(removed)

			// If set is empty, delete the key
			count, err := r.client.ZCard(ctx, key).Result()
			if err == nil && count == 0 {
				r.client.Del(ctx, key)
			}
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error during cleanup scan: %w", err)
	}

	r.logger.Info("Completed relationship cleanup",
		zap.Time("older_than", olderThan),
		zap.Int("deleted_count", deletedCount))

	return nil
}

// Helper methods

func (r *RedisGraphStore) getChildrenKey(parentID string) string {
	return fmt.Sprintf("%s:children:%s", r.config.RedisKeyPrefix, parentID)
}

func (r *RedisGraphStore) getParentsKey(childID string) string {
	return fmt.Sprintf("%s:parents:%s", r.config.RedisKeyPrefix, childID)
}

func (r *RedisGraphStore) getRelationshipKey(parentID, childID string) string {
	return fmt.Sprintf("%s:relations:%s:%s", r.config.RedisKeyPrefix, parentID, childID)
}

func (r *RedisGraphStore) getRelationshipsAsParent(ctx context.Context, jobID string) ([]JobRelationship, error) {
	relationships := make([]JobRelationship, 0)

	children, err := r.GetChildren(ctx, jobID)
	if err != nil {
		return nil, err
	}

	for _, childID := range children {
		rel, err := r.getRelationshipMetadata(ctx, jobID, childID)
		if err != nil {
			r.logger.Warn("Failed to get relationship metadata",
				zap.String("parent_id", jobID),
				zap.String("child_id", childID),
				zap.Error(err))
			continue
		}
		if rel != nil {
			relationships = append(relationships, *rel)
		}
	}

	return relationships, nil
}

func (r *RedisGraphStore) getRelationshipsAsChild(ctx context.Context, jobID string) ([]JobRelationship, error) {
	relationships := make([]JobRelationship, 0)

	parents, err := r.GetParents(ctx, jobID)
	if err != nil {
		return nil, err
	}

	for _, parentID := range parents {
		rel, err := r.getRelationshipMetadata(ctx, parentID, jobID)
		if err != nil {
			r.logger.Warn("Failed to get relationship metadata",
				zap.String("parent_id", parentID),
				zap.String("child_id", jobID),
				zap.Error(err))
			continue
		}
		if rel != nil {
			relationships = append(relationships, *rel)
		}
	}

	return relationships, nil
}

func (r *RedisGraphStore) getRelationshipMetadata(ctx context.Context, parentID, childID string) (*JobRelationship, error) {
	relationKey := r.getRelationshipKey(parentID, childID)

	data, err := r.client.HGetAll(ctx, relationKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	rel := &JobRelationship{
		ParentID: parentID,
		ChildID:  childID,
	}

	if relType, exists := data["type"]; exists {
		rel.Type = RelationshipType(relType)
	}

	if spawnReason, exists := data["spawn_reason"]; exists {
		rel.SpawnReason = spawnReason
	}

	if timestampStr, exists := data["timestamp"]; exists {
		if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
			rel.Timestamp = time.Unix(timestamp, 0)
		}
	}

	if metadataStr, exists := data["metadata"]; exists && metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			rel.Metadata = metadata
		}
	}

	return rel, nil
}

func (r *RedisGraphStore) findTrueRoot(ctx context.Context, jobID string) (string, error) {
	currentID := jobID
	visited := make(map[string]bool)

	for {
		if visited[currentID] {
			// Cycle detected, return current
			return currentID, nil
		}
		visited[currentID] = true

		parents, err := r.GetParents(ctx, currentID)
		if err != nil {
			return "", err
		}

		if len(parents) == 0 {
			// Found root
			return currentID, nil
		}

		// Move to first parent (in case of multiple parents, this gives a deterministic result)
		currentID = parents[0]
	}
}

func (r *RedisGraphStore) deduplicateRelationships(relationships []JobRelationship) []JobRelationship {
	seen := make(map[string]bool)
	deduplicated := make([]JobRelationship, 0)

	for _, rel := range relationships {
		key := fmt.Sprintf("%s->%s", rel.ParentID, rel.ChildID)
		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, rel)
		}
	}

	return deduplicated
}