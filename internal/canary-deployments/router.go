package canary_deployments

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math/rand"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
)

// RedisRouter implements the Router interface using Redis for job routing
type RedisRouter struct {
	redis    *redis.Client
	logger   *slog.Logger
	splitters map[string]*QueueSplitter
	mu       sync.RWMutex
}

// NewRedisRouter creates a new Redis-based router
func NewRedisRouter(redis *redis.Client, logger *slog.Logger) *RedisRouter {
	return &RedisRouter{
		redis:     redis,
		logger:    logger,
		splitters: make(map[string]*QueueSplitter),
	}
}

// RouteJob routes a job to the appropriate queue based on canary configuration
func (r *RedisRouter) RouteJob(ctx context.Context, job *Job) (string, error) {
	r.mu.RLock()
	splitter, exists := r.splitters[job.Queue]
	r.mu.RUnlock()

	if !exists {
		// No canary configured for this queue, use original queue
		return job.Queue, nil
	}

	targetQueue := r.routeWithSplitter(job, splitter)

	r.logger.Debug("Routed job",
		"job_id", job.ID,
		"original_queue", job.Queue,
		"target_queue", targetQueue,
		"canary_percentage", splitter.Percentage)

	return targetQueue, nil
}

// UpdateRoutingPercentage updates the canary percentage for a queue
func (r *RedisRouter) UpdateRoutingPercentage(ctx context.Context, queue string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return NewInvalidPercentageError(percentage)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if percentage == 0 {
		// Remove canary routing
		delete(r.splitters, queue)
		r.logger.Info("Removed canary routing", "queue", queue)
	} else {
		// Update or create splitter
		splitter, exists := r.splitters[queue]
		if !exists {
			splitter = &QueueSplitter{
				StableQueue: queue,
				CanaryQueue: queue + "@canary",
				StickyHash:  true, // Default to sticky routing
			}
			r.splitters[queue] = splitter
		}

		splitter.Percentage = percentage
		r.logger.Info("Updated canary routing",
			"queue", queue,
			"percentage", percentage)
	}

	// Persist routing configuration to Redis
	if err := r.saveRoutingConfig(ctx, queue, percentage); err != nil {
		return fmt.Errorf("failed to save routing config: %w", err)
	}

	return nil
}

// GetRoutingStats returns routing statistics for a queue
func (r *RedisRouter) GetRoutingStats(ctx context.Context, queue string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get queue depths
	stableDepth, err := r.redis.LLen(ctx, queue).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stable queue depth: %w", err)
	}
	stats["stable_depth"] = stableDepth

	canaryQueue := queue + "@canary"
	canaryDepth, err := r.redis.LLen(ctx, canaryQueue).Result()
	if err != nil {
		// Canary queue might not exist
		canaryDepth = 0
	}
	stats["canary_depth"] = canaryDepth

	// Get routing counters from Redis
	stableKey := fmt.Sprintf("canary:stats:%s:stable", queue)
	canaryKey := fmt.Sprintf("canary:stats:%s:canary", queue)

	stableCount, err := r.redis.Get(ctx, stableKey).Int64()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get stable routing count: %w", err)
	}
	stats["stable_routed"] = stableCount

	canaryCount, err := r.redis.Get(ctx, canaryKey).Int64()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get canary routing count: %w", err)
	}
	stats["canary_routed"] = canaryCount

	// Calculate current percentage
	total := stableCount + canaryCount
	if total > 0 {
		stats["current_percentage"] = (canaryCount * 100) / total
	} else {
		stats["current_percentage"] = 0
	}

	return stats, nil
}

// SetStickyRouting enables or disables sticky routing for a queue
func (r *RedisRouter) SetStickyRouting(queue string, sticky bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	splitter, exists := r.splitters[queue]
	if exists {
		splitter.StickyHash = sticky
		r.logger.Info("Updated sticky routing",
			"queue", queue,
			"sticky", sticky)
	}
}

// LoadRoutingConfig loads routing configuration from Redis
func (r *RedisRouter) LoadRoutingConfig(ctx context.Context) error {
	pattern := "canary:routing:*"
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to list routing keys: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, key := range keys {
		// Extract queue name from key
		queue := key[len("canary:routing:"):]

		percentageStr, err := r.redis.Get(ctx, key).Result()
		if err != nil {
			r.logger.Warn("Failed to load routing config",
				"key", key,
				"error", err)
			continue
		}

		percentage, err := strconv.Atoi(percentageStr)
		if err != nil {
			r.logger.Warn("Invalid percentage in routing config",
				"key", key,
				"value", percentageStr)
			continue
		}

		if percentage > 0 {
			r.splitters[queue] = &QueueSplitter{
				StableQueue: queue,
				CanaryQueue: queue + "@canary",
				Percentage:  percentage,
				StickyHash:  true,
			}
		}
	}

	r.logger.Info("Loaded routing configuration",
		"splitters_count", len(r.splitters))

	return nil
}

// Private methods

func (r *RedisRouter) routeWithSplitter(job *Job, splitter *QueueSplitter) string {
	if splitter.Percentage <= 0 {
		return splitter.StableQueue
	}

	if splitter.Percentage >= 100 {
		return splitter.CanaryQueue
	}

	var shouldRouteToCanary bool

	if splitter.StickyHash {
		// Use consistent hashing based on job ID
		shouldRouteToCanary = r.hashBasedRouting(job.ID, splitter.Percentage)
	} else {
		// Use random routing
		shouldRouteToCanary = rand.Intn(100) < splitter.Percentage
	}

	if shouldRouteToCanary {
		// Update routing statistics
		r.updateRoutingStats(job.Queue, "canary")
		return splitter.CanaryQueue
	} else {
		// Update routing statistics
		r.updateRoutingStats(job.Queue, "stable")
		return splitter.StableQueue
	}
}

func (r *RedisRouter) hashBasedRouting(jobID string, percentage int) bool {
	hash := fnv.New32a()
	hash.Write([]byte(jobID))
	hashValue := hash.Sum32()

	// Use hash to determine routing (ensures consistency)
	return int(hashValue%100) < percentage
}

func (r *RedisRouter) updateRoutingStats(queue, destination string) {
	// Update routing statistics in Redis (async)
	go func() {
		ctx := context.Background()
		key := fmt.Sprintf("canary:stats:%s:%s", queue, destination)

		if err := r.redis.Incr(ctx, key).Err(); err != nil {
			r.logger.Debug("Failed to update routing stats",
				"queue", queue,
				"destination", destination,
				"error", err)
		}

		// Set expiration to prevent accumulation
		r.redis.Expire(ctx, key, 24*60*60) // 24 hours
	}()
}

func (r *RedisRouter) saveRoutingConfig(ctx context.Context, queue string, percentage int) error {
	key := fmt.Sprintf("canary:routing:%s", queue)

	if percentage == 0 {
		// Remove configuration
		return r.redis.Del(ctx, key).Err()
	}

	// Save configuration
	return r.redis.Set(ctx, key, percentage, 0).Err()
}

// ConsistentHashRouter provides consistent hash-based routing for multi-step workflows
type ConsistentHashRouter struct {
	router   *RedisRouter
	hashRing *HashRing
}

// HashRing implements a simple consistent hash ring
type HashRing struct {
	nodes   []HashNode
	mu      sync.RWMutex
}

type HashNode struct {
	Hash  uint32
	Value string
}

// NewConsistentHashRouter creates a router with consistent hashing
func NewConsistentHashRouter(redis *redis.Client, logger *slog.Logger) *ConsistentHashRouter {
	return &ConsistentHashRouter{
		router:   NewRedisRouter(redis, logger),
		hashRing: NewHashRing(),
	}
}

// NewHashRing creates a new hash ring
func NewHashRing() *HashRing {
	return &HashRing{
		nodes: make([]HashNode, 0),
	}
}

// UpdateNodes updates the hash ring with new node weights
func (hr *HashRing) UpdateNodes(stableWeight, canaryWeight int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.nodes = hr.nodes[:0] // Clear existing nodes

	// Add stable nodes
	for i := 0; i < stableWeight; i++ {
		hash := r.hash(fmt.Sprintf("stable-%d", i))
		hr.nodes = append(hr.nodes, HashNode{Hash: hash, Value: "stable"})
	}

	// Add canary nodes
	for i := 0; i < canaryWeight; i++ {
		hash := r.hash(fmt.Sprintf("canary-%d", i))
		hr.nodes = append(hr.nodes, HashNode{Hash: hash, Value: "canary"})
	}

	// Sort nodes by hash
	for i := 0; i < len(hr.nodes); i++ {
		for j := i + 1; j < len(hr.nodes); j++ {
			if hr.nodes[i].Hash > hr.nodes[j].Hash {
				hr.nodes[i], hr.nodes[j] = hr.nodes[j], hr.nodes[i]
			}
		}
	}
}

// GetNode returns the node responsible for a given key
func (hr *HashRing) GetNode(key string) string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.nodes) == 0 {
		return "stable" // Default fallback
	}

	hash := r.hash(key)

	// Find the first node with hash >= key hash
	for _, node := range hr.nodes {
		if node.Hash >= hash {
			return node.Value
		}
	}

	// Wrap around to the first node
	return hr.nodes[0].Value
}

func (hr *HashRing) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// StreamGroupRouter implements routing using Redis Streams consumer groups
type StreamGroupRouter struct {
	redis  *redis.Client
	logger *slog.Logger
	configs map[string]*StreamCanaryConfig
	mu     sync.RWMutex
}

// NewStreamGroupRouter creates a new stream group router
func NewStreamGroupRouter(redis *redis.Client, logger *slog.Logger) *StreamGroupRouter {
	return &StreamGroupRouter{
		redis:   redis,
		logger:  logger,
		configs: make(map[string]*StreamCanaryConfig),
	}
}

// RouteJob routes a job using stream groups
func (sgr *StreamGroupRouter) RouteJob(ctx context.Context, job *Job) (string, error) {
	sgr.mu.RLock()
	config, exists := sgr.configs[job.Queue]
	sgr.mu.RUnlock()

	if !exists {
		return job.Queue, nil
	}

	// Add to stream with appropriate routing metadata
	streamKey := config.StreamKey

	// Determine consumer group based on canary weight
	group := "stable"
	if rand.Float64() < config.CanaryWeight {
		group = "canary"
	}

	// Add routing metadata to job
	if job.Metadata == nil {
		job.Metadata = make(map[string]string)
	}
	job.Metadata["target_group"] = group

	sgr.logger.Debug("Routed job to stream group",
		"job_id", job.ID,
		"stream", streamKey,
		"group", group)

	return streamKey, nil
}

// UpdateRoutingPercentage updates the canary weight for stream routing
func (sgr *StreamGroupRouter) UpdateRoutingPercentage(ctx context.Context, queue string, percentage int) error {
	if percentage < 0 || percentage > 100 {
		return NewInvalidPercentageError(percentage)
	}

	weight := float64(percentage) / 100.0

	sgr.mu.Lock()
	defer sgr.mu.Unlock()

	if percentage == 0 {
		delete(sgr.configs, queue)
	} else {
		config, exists := sgr.configs[queue]
		if !exists {
			config = &StreamCanaryConfig{
				StreamKey:   queue,
				StableGroup: queue + ":stable",
				CanaryGroup: queue + ":canary",
			}
			sgr.configs[queue] = config
		}
		config.CanaryWeight = weight
	}

	sgr.logger.Info("Updated stream group routing",
		"queue", queue,
		"percentage", percentage,
		"weight", weight)

	return nil
}

// GetRoutingStats returns routing statistics for stream groups
func (sgr *StreamGroupRouter) GetRoutingStats(ctx context.Context, queue string) (map[string]int64, error) {
	sgr.mu.RLock()
	config, exists := sgr.configs[queue]
	sgr.mu.RUnlock()

	if !exists {
		return map[string]int64{}, nil
	}

	stats := make(map[string]int64)

	// Get stream length
	streamLen, err := sgr.redis.XLen(ctx, config.StreamKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream length: %w", err)
	}
	stats["stream_length"] = streamLen

	// Get pending messages for each group
	stablePending, err := sgr.getGroupPending(ctx, config.StreamKey, config.StableGroup)
	if err != nil {
		sgr.logger.Warn("Failed to get stable group pending", "error", err)
		stablePending = 0
	}
	stats["stable_pending"] = stablePending

	canaryPending, err := sgr.getGroupPending(ctx, config.StreamKey, config.CanaryGroup)
	if err != nil {
		sgr.logger.Warn("Failed to get canary group pending", "error", err)
		canaryPending = 0
	}
	stats["canary_pending"] = canaryPending

	// Calculate current weight percentage
	weight := config.CanaryWeight * 100
	stats["current_percentage"] = int64(weight)

	return stats, nil
}

func (sgr *StreamGroupRouter) getGroupPending(ctx context.Context, stream, group string) (int64, error) {
	info, err := sgr.redis.XInfoGroup(ctx, stream, group).Result()
	if err != nil {
		return 0, err
	}

	if len(info) > 0 {
		if pending, ok := info[0]["pending"].(int64); ok {
			return pending, nil
		}
	}

	return 0, nil
}