package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStreamsConfig configures the Redis Streams backend
type RedisStreamsConfig struct {
	URL              string        `json:"url" yaml:"url"`
	Database         int           `json:"database" yaml:"database"`
	Password         string        `json:"password" yaml:"password"`
	StreamName       string        `json:"stream_name" yaml:"stream_name"`
	ConsumerGroup    string        `json:"consumer_group" yaml:"consumer_group"`
	ConsumerName     string        `json:"consumer_name" yaml:"consumer_name"`
	MaxLength        int64         `json:"max_length" yaml:"max_length"`
	BlockTimeout     time.Duration `json:"block_timeout" yaml:"block_timeout"`
	ClaimMinIdle     time.Duration `json:"claim_min_idle" yaml:"claim_min_idle"`
	ClaimCount       int64         `json:"claim_count" yaml:"claim_count"`
	MaxConnections   int           `json:"max_connections" yaml:"max_connections"`
	ConnTimeout      time.Duration `json:"conn_timeout" yaml:"conn_timeout"`
	ReadTimeout      time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout" yaml:"write_timeout"`
	PoolTimeout      time.Duration `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxRetries       int           `json:"max_retries" yaml:"max_retries"`
	ClusterMode      bool          `json:"cluster_mode" yaml:"cluster_mode"`
	ClusterAddrs     []string      `json:"cluster_addrs" yaml:"cluster_addrs"`
	TLS              bool          `json:"tls" yaml:"tls"`
}

// RedisStreamsBackend implements QueueBackend using Redis Streams
type RedisStreamsBackend struct {
	client        redis.Cmdable
	config        RedisStreamsConfig
	streamName    string
	consumerGroup string
	consumerName  string

	// Processing tracking
	processingJobs map[string]*Job
	processingLock sync.RWMutex

	// Metrics tracking
	stats     *BackendStats
	statsLock sync.RWMutex
}

// RedisStreamsFactory creates Redis Streams backend instances
type RedisStreamsFactory struct{}

// Create creates a new Redis Streams backend
func (f *RedisStreamsFactory) Create(config interface{}) (QueueBackend, error) {
	cfg, ok := config.(RedisStreamsConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for redis-streams backend")
	}

	return NewRedisStreamsBackend(cfg)
}

// Validate validates Redis Streams backend configuration
func (f *RedisStreamsFactory) Validate(config interface{}) error {
	cfg, ok := config.(RedisStreamsConfig)
	if !ok {
		return fmt.Errorf("invalid config type for redis-streams backend")
	}

	if cfg.URL == "" && len(cfg.ClusterAddrs) == 0 {
		return fmt.Errorf("either URL or cluster addresses must be provided")
	}

	if cfg.StreamName == "" {
		return fmt.Errorf("stream name is required")
	}

	if cfg.ConsumerGroup == "" {
		return fmt.Errorf("consumer group is required")
	}

	if cfg.ConsumerName == "" {
		return fmt.Errorf("consumer name is required")
	}

	return nil
}

// NewRedisStreamsBackend creates a new Redis Streams backend
func NewRedisStreamsBackend(config RedisStreamsConfig) (*RedisStreamsBackend, error) {
	var client redis.Cmdable

	if config.ClusterMode && len(config.ClusterAddrs) > 0 {
		// Redis Cluster client
		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.ClusterAddrs,
			Password:     config.Password,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.ConnTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,
			IdleTimeout:  config.IdleTimeout,
		})
		client = clusterClient
	} else {
		// Single Redis client
		opt, err := redis.ParseURL(config.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid Redis URL: %w", err)
		}

		opt.DB = config.Database
		opt.Password = config.Password
		opt.MaxRetries = config.MaxRetries
		opt.DialTimeout = config.ConnTimeout
		opt.ReadTimeout = config.ReadTimeout
		opt.WriteTimeout = config.WriteTimeout
		opt.PoolTimeout = config.PoolTimeout
		opt.IdleTimeout = config.IdleTimeout

		if config.MaxConnections > 0 {
			opt.PoolSize = config.MaxConnections
		}

		client = redis.NewClient(opt)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	backend := &RedisStreamsBackend{
		client:         client,
		config:         config,
		streamName:     config.StreamName,
		consumerGroup:  config.ConsumerGroup,
		consumerName:   config.ConsumerName,
		processingJobs: make(map[string]*Job),
		stats: &BackendStats{
			EnqueueRate: 0.0,
			DequeueRate: 0.0,
			ErrorRate:   0.0,
			QueueDepth:  0,
		},
	}

	// Ensure consumer group exists
	if err := backend.ensureConsumerGroup(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure consumer group: %w", err)
	}

	return backend, nil
}

// ensureConsumerGroup creates the consumer group if it doesn't exist
func (r *RedisStreamsBackend) ensureConsumerGroup(ctx context.Context) error {
	// Try to create the consumer group
	err := r.client.XGroupCreate(ctx, r.streamName, r.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		// Create the stream first if it doesn't exist
		if err.Error() == "ERR The XGROUP subcommand requires the key to exist" {
			// Add a dummy entry to create the stream, then delete it
			id, addErr := r.client.XAdd(ctx, &redis.XAddArgs{
				Stream: r.streamName,
				ID:     "*",
				Values: map[string]interface{}{"init": "true"},
			}).Result()
			if addErr != nil {
				return fmt.Errorf("failed to create stream: %w", addErr)
			}

			// Delete the dummy entry
			r.client.XDel(ctx, r.streamName, id)

			// Now create the consumer group
			err = r.client.XGroupCreate(ctx, r.streamName, r.consumerGroup, "0").Err()
		}

		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create consumer group: %w", err)
		}
	}

	return nil
}

// Enqueue adds a job to the stream
func (r *RedisStreamsBackend) Enqueue(ctx context.Context, job *Job) error {
	r.updateStats("enqueue", time.Now(), nil)

	// Prepare stream entry values
	values := map[string]interface{}{
		"id":          job.ID,
		"type":        job.Type,
		"queue":       job.Queue,
		"priority":    job.Priority,
		"created_at":  job.CreatedAt.Unix(),
		"retry_count": job.RetryCount,
		"max_retries": job.MaxRetries,
	}

	// Serialize payload
	if job.Payload != nil {
		payloadData, err := json.Marshal(job.Payload)
		if err != nil {
			r.updateStats("enqueue", time.Now(), err)
			return fmt.Errorf("failed to serialize job payload: %w", err)
		}
		values["payload"] = string(payloadData)
	}

	// Serialize metadata
	if job.Metadata != nil {
		metadataData, err := json.Marshal(job.Metadata)
		if err != nil {
			r.updateStats("enqueue", time.Now(), err)
			return fmt.Errorf("failed to serialize job metadata: %w", err)
		}
		values["metadata"] = string(metadataData)
	}

	// Serialize tags
	if len(job.Tags) > 0 {
		tagsData, err := json.Marshal(job.Tags)
		if err != nil {
			r.updateStats("enqueue", time.Now(), err)
			return fmt.Errorf("failed to serialize job tags: %w", err)
		}
		values["tags"] = string(tagsData)
	}

	// Add to stream
	args := &redis.XAddArgs{
		Stream: r.streamName,
		ID:     "*",
		Values: values,
	}

	if r.config.MaxLength > 0 {
		args.MaxLen = r.config.MaxLength
		args.MaxLenApprox = 1
	}

	_, err := r.client.XAdd(ctx, args).Result()
	if err != nil {
		r.updateStats("enqueue", time.Now(), err)
		return fmt.Errorf("failed to add job to stream: %w", err)
	}

	return nil
}

// Dequeue reads a job from the stream
func (r *RedisStreamsBackend) Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error) {
	r.updateStats("dequeue", time.Now(), nil)

	consumerGroup := r.consumerGroup
	consumerName := r.consumerName

	if opts.ConsumerGroup != "" {
		consumerGroup = opts.ConsumerGroup
	}
	if opts.ConsumerID != "" {
		consumerName = opts.ConsumerID
	}

	timeout := r.config.BlockTimeout
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}

	count := int64(1)
	if opts.Count > 0 {
		count = int64(opts.Count)
	}

	// First, try to claim any pending messages
	if r.config.ClaimMinIdle > 0 {
		if err := r.claimPendingMessages(ctx, consumerGroup, consumerName); err != nil {
			// Log error but continue with normal reading
		}
	}

	// Read from the consumer group
	args := &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumerName,
		Streams:  []string{r.streamName, ">"},
		Count:    count,
		Block:    timeout,
	}

	streams, err := r.client.XReadGroup(ctx, args).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No messages available
		}
		r.updateStats("dequeue", time.Now(), err)
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}

	// Parse the first message
	msg := streams[0].Messages[0]
	job, err := r.parseStreamMessage(msg)
	if err != nil {
		r.updateStats("dequeue", time.Now(), err)
		return nil, fmt.Errorf("failed to parse stream message: %w", err)
	}

	// Track processing job
	r.processingLock.Lock()
	r.processingJobs[msg.ID] = job
	r.processingLock.Unlock()

	return job, nil
}

// Ack acknowledges a job by ID
func (r *RedisStreamsBackend) Ack(ctx context.Context, jobID string) error {
	// Find the stream message ID for this job
	messageID := r.findMessageID(jobID)
	if messageID == "" {
		return fmt.Errorf("job %s not found in processing jobs", jobID)
	}

	// Acknowledge the message
	err := r.client.XAck(ctx, r.streamName, r.consumerGroup, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	// Remove from processing jobs
	r.processingLock.Lock()
	delete(r.processingJobs, messageID)
	r.processingLock.Unlock()

	return nil
}

// Nack negatively acknowledges a job
func (r *RedisStreamsBackend) Nack(ctx context.Context, jobID string, requeue bool) error {
	messageID := r.findMessageID(jobID)
	if messageID == "" {
		return fmt.Errorf("job %s not found in processing jobs", jobID)
	}

	if requeue {
		// Re-add the job to the stream (simplified approach)
		// In a full implementation, we might modify retry count, etc.
		r.processingLock.RLock()
		job, exists := r.processingJobs[messageID]
		r.processingLock.RUnlock()

		if exists {
			job.RetryCount++
			if err := r.Enqueue(ctx, job); err != nil {
				return fmt.Errorf("failed to requeue job: %w", err)
			}
		}
	}

	// Acknowledge the original message to remove it from pending
	err := r.client.XAck(ctx, r.streamName, r.consumerGroup, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	// Remove from processing jobs
	r.processingLock.Lock()
	delete(r.processingJobs, messageID)
	r.processingLock.Unlock()

	return nil
}

// Length returns the stream length
func (r *RedisStreamsBackend) Length(ctx context.Context) (int64, error) {
	info, err := r.client.XInfoStream(ctx, r.streamName).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get stream info: %w", err)
	}

	r.statsLock.Lock()
	r.stats.QueueDepth = info.Length
	r.stats.StreamLength = &info.Length
	r.statsLock.Unlock()

	return info.Length, nil
}

// Peek looks at entries in the stream
func (r *RedisStreamsBackend) Peek(ctx context.Context, offset int64) (*Job, error) {
	// Use XRANGE to peek at stream entries
	messages, err := r.client.XRange(ctx, r.streamName, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read stream range: %w", err)
	}

	if offset >= int64(len(messages)) {
		return nil, nil
	}

	return r.parseStreamMessage(messages[offset])
}

// Move transfers a job to another stream/queue
func (r *RedisStreamsBackend) Move(ctx context.Context, jobID string, targetQueue string) error {
	// Find the job in processing jobs
	messageID := r.findMessageID(jobID)
	if messageID == "" {
		return fmt.Errorf("job %s not found in processing jobs", jobID)
	}

	r.processingLock.RLock()
	job, exists := r.processingJobs[messageID]
	r.processingLock.RUnlock()

	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	// Create a new backend for the target queue
	targetConfig := r.config
	targetConfig.StreamName = targetQueue

	targetBackend, err := NewRedisStreamsBackend(targetConfig)
	if err != nil {
		return fmt.Errorf("failed to create target backend: %w", err)
	}
	defer targetBackend.Close()

	// Enqueue to target
	job.Queue = targetQueue
	if err := targetBackend.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("failed to enqueue to target queue: %w", err)
	}

	// Acknowledge in source
	return r.Ack(ctx, jobID)
}

// Iter returns an iterator for stream entries
func (r *RedisStreamsBackend) Iter(ctx context.Context, opts IterOptions) (Iterator, error) {
	startID := "-"
	endID := "+"

	if opts.StartID != "" {
		startID = opts.StartID
	}
	if opts.EndID != "" {
		endID = opts.EndID
	}

	messages, err := r.client.XRange(ctx, r.streamName, startID, endID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read stream range: %w", err)
	}

	// Convert messages to jobs
	jobs := make([]*Job, 0, len(messages))
	for _, msg := range messages {
		job, err := r.parseStreamMessage(msg)
		if err != nil {
			continue // Skip malformed messages
		}
		jobs = append(jobs, job)
	}

	// Apply count limit
	if opts.Count > 0 && int64(len(jobs)) > opts.Count {
		jobs = jobs[:opts.Count]
	}

	// Apply reverse order
	if opts.Reverse {
		for i, j := 0, len(jobs)-1; i < j; i, j = i+1, j-1 {
			jobs[i], jobs[j] = jobs[j], jobs[i]
		}
	}

	return NewJobIterator(jobs), nil
}

// Capabilities returns the capabilities of Redis Streams backend
func (r *RedisStreamsBackend) Capabilities() BackendCapabilities {
	return BackendCapabilities{
		AtomicAck:          true,  // XACK guarantees
		ConsumerGroups:     true,  // Native XGROUP support
		Replay:             true,  // Historical XREAD
		IdempotentEnqueue:  false, // Application level
		Transactions:       true,
		Persistence:        true,
		Clustering:         r.config.ClusterMode,
		TimeToLive:         false,
		Prioritization:     false,
		BatchOperations:    true,
	}
}

// Stats returns backend performance statistics
func (r *RedisStreamsBackend) Stats(ctx context.Context) (*BackendStats, error) {
	r.statsLock.RLock()
	defer r.statsLock.RUnlock()

	stats := *r.stats

	// Get stream info
	info, err := r.client.XInfoStream(ctx, r.streamName).Result()
	if err == nil {
		stats.StreamLength = &info.Length
		stats.QueueDepth = info.Length
	}

	// Get consumer group info
	groups, err := r.client.XInfoGroups(ctx, r.streamName).Result()
	if err == nil {
		for _, group := range groups {
			if group.Name == r.consumerGroup {
				// Consumer lag calculation would need to be implemented
				// by comparing last delivered ID with stream length
				lag := int64(0) // Placeholder
				stats.ConsumerLag = &lag
				break
			}
		}
	}

	// Add connection pool stats if available
	if client, ok := r.client.(*redis.Client); ok {
		poolStats := client.PoolStats()
		stats.ConnectionPool = &PoolStats{
			Active:  int(poolStats.Hits),
			Idle:    int(poolStats.Misses),
			Total:   int(poolStats.Hits + poolStats.Misses),
			MaxOpen: 0, // Not available from go-redis
			MaxIdle: 0, // Not available from go-redis
		}
	}

	return &stats, nil
}

// Health performs a health check
func (r *RedisStreamsBackend) Health(ctx context.Context) HealthStatus {
	status := HealthStatus{
		CheckedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Ping Redis
	err := r.client.Ping(ctx).Err()
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Error = err
		status.Message = "Redis ping failed"
		return status
	}

	// Check stream existence
	info, err := r.client.XInfoStream(ctx, r.streamName).Result()
	if err != nil {
		status.Status = HealthStatusDegraded
		status.Message = "Cannot get stream info"
		return status
	}

	status.Metadata["stream_length"] = strconv.FormatInt(info.Length, 10)

	// Check consumer group lag
	groups, err := r.client.XInfoGroups(ctx, r.streamName).Result()
	if err == nil {
		for _, group := range groups {
			if group.Name == r.consumerGroup {
				// Consumer lag would need proper calculation
				lag := int64(0) // Placeholder
				status.Metadata["consumer_lag"] = strconv.FormatInt(lag, 10)
				if lag > 1000 {
					status.Status = HealthStatusDegraded
					status.Message = "High consumer lag"
				} else {
					status.Status = HealthStatusHealthy
				}
				break
			}
		}
	} else {
		status.Status = HealthStatusHealthy
	}

	return status
}

// Close closes the Redis connection
func (r *RedisStreamsBackend) Close() error {
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// Helper methods

func (r *RedisStreamsBackend) parseStreamMessage(msg redis.XMessage) (*Job, error) {
	job := &Job{}

	// Parse basic fields
	if id, ok := msg.Values["id"].(string); ok {
		job.ID = id
	}
	if jobType, ok := msg.Values["type"].(string); ok {
		job.Type = jobType
	}
	if queue, ok := msg.Values["queue"].(string); ok {
		job.Queue = queue
	}
	if priority, ok := msg.Values["priority"].(string); ok {
		if p, err := strconv.Atoi(priority); err == nil {
			job.Priority = p
		}
	}
	if createdAt, ok := msg.Values["created_at"].(string); ok {
		if timestamp, err := strconv.ParseInt(createdAt, 10, 64); err == nil {
			job.CreatedAt = time.Unix(timestamp, 0)
		}
	}
	if retryCount, ok := msg.Values["retry_count"].(string); ok {
		if count, err := strconv.Atoi(retryCount); err == nil {
			job.RetryCount = count
		}
	}
	if maxRetries, ok := msg.Values["max_retries"].(string); ok {
		if max, err := strconv.Atoi(maxRetries); err == nil {
			job.MaxRetries = max
		}
	}

	// Parse payload
	if payloadStr, ok := msg.Values["payload"].(string); ok {
		var payload interface{}
		if err := json.Unmarshal([]byte(payloadStr), &payload); err == nil {
			job.Payload = payload
		}
	}

	// Parse metadata
	if metadataStr, ok := msg.Values["metadata"].(string); ok {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			job.Metadata = metadata
		}
	}

	// Parse tags
	if tagsStr, ok := msg.Values["tags"].(string); ok {
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err == nil {
			job.Tags = tags
		}
	}

	return job, nil
}

func (r *RedisStreamsBackend) findMessageID(jobID string) string {
	r.processingLock.RLock()
	defer r.processingLock.RUnlock()

	for messageID, job := range r.processingJobs {
		if job.ID == jobID {
			return messageID
		}
	}
	return ""
}

func (r *RedisStreamsBackend) claimPendingMessages(ctx context.Context, consumerGroup, consumerName string) error {
	// Get pending messages
	pending, err := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   r.streamName,
		Group:    consumerGroup,
		Start:    "-",
		End:      "+",
		Count:    r.config.ClaimCount,
		Consumer: consumerName,
	}).Result()
	if err != nil {
		return err
	}

	// Claim messages that have been idle too long
	var messageIDs []string

	for _, p := range pending {
		// Check if message has been idle long enough to be claimed
		// Note: go-redis XPendingExt may not have LastDeliveredTime field
		// This would need to be implemented based on the actual redis library version
		if r.config.ClaimMinIdle > 0 {
			messageIDs = append(messageIDs, p.ID)
		}
	}

	if len(messageIDs) > 0 {
		_, err = r.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   r.streamName,
			Group:    consumerGroup,
			Consumer: consumerName,
			MinIdle:  r.config.ClaimMinIdle,
			Messages: messageIDs,
		}).Result()
		return err
	}

	return nil
}

func (r *RedisStreamsBackend) updateStats(operation string, timestamp time.Time, err error) {
	r.statsLock.Lock()
	defer r.statsLock.Unlock()

	if err != nil {
		r.stats.ErrorRate++
		r.stats.LastError = timestamp
	} else {
		switch operation {
		case "enqueue":
			r.stats.EnqueueRate++
			r.stats.LastEnqueue = timestamp
		case "dequeue":
			r.stats.DequeueRate++
			r.stats.LastDequeue = timestamp
		}
	}
}

// Initialize Redis Streams backend in the registry
func init() {
	RegisterBackend(BackendTypeRedisStreams, &RedisStreamsFactory{})
}