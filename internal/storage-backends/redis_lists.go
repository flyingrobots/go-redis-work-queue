package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisListsConfig configures the Redis Lists backend
type RedisListsConfig struct {
	URL            string            `json:"url" yaml:"url"`
	Database       int               `json:"database" yaml:"database"`
	Password       string            `json:"password" yaml:"password"`
	KeyPrefix      string            `json:"key_prefix" yaml:"key_prefix"`
	MaxConnections int               `json:"max_connections" yaml:"max_connections"`
	ConnTimeout    time.Duration     `json:"conn_timeout" yaml:"conn_timeout"`
	ReadTimeout    time.Duration     `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration     `json:"write_timeout" yaml:"write_timeout"`
	PoolTimeout    time.Duration     `json:"pool_timeout" yaml:"pool_timeout"`
	IdleTimeout    time.Duration     `json:"idle_timeout" yaml:"idle_timeout"`
	MaxRetries     int               `json:"max_retries" yaml:"max_retries"`
	ClusterMode    bool              `json:"cluster_mode" yaml:"cluster_mode"`
	ClusterAddrs   []string          `json:"cluster_addrs" yaml:"cluster_addrs"`
	TLS            bool              `json:"tls" yaml:"tls"`
	Options        map[string]string `json:"options" yaml:"options"`
}

// RedisListsBackend implements QueueBackend using Redis Lists
type RedisListsBackend struct {
	client    redis.Cmdable
	config    RedisListsConfig
	queueName string
	keyPrefix string

	// Metrics tracking
	stats     *BackendStats
	statsLock sync.RWMutex
}

// RedisListsFactory creates Redis Lists backend instances
type RedisListsFactory struct{}

// Create creates a new Redis Lists backend
func (f *RedisListsFactory) Create(config interface{}) (QueueBackend, error) {
	cfg, ok := config.(RedisListsConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type for redis-lists backend")
	}

	return NewRedisListsBackend(cfg)
}

// Validate validates Redis Lists backend configuration
func (f *RedisListsFactory) Validate(config interface{}) error {
	cfg, ok := config.(RedisListsConfig)
	if !ok {
		return fmt.Errorf("invalid config type for redis-lists backend")
	}

	if cfg.URL == "" && len(cfg.ClusterAddrs) == 0 {
		return fmt.Errorf("either URL or cluster addresses must be provided")
	}

	return nil
}

// NewRedisListsBackend creates a new Redis Lists backend
func NewRedisListsBackend(config RedisListsConfig) (*RedisListsBackend, error) {
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

	keyPrefix := config.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "queue:"
	}

	return &RedisListsBackend{
		client:    client,
		config:    config,
		keyPrefix: keyPrefix,
		stats: &BackendStats{
			EnqueueRate: 0.0,
			DequeueRate: 0.0,
			ErrorRate:   0.0,
			QueueDepth:  0,
		},
	}, nil
}

// Enqueue adds a job to the queue
func (r *RedisListsBackend) Enqueue(ctx context.Context, job *Job) error {
	r.updateStats("enqueue", time.Now(), nil)

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		r.updateStats("enqueue", time.Now(), err)
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Push to list
	queueKey := r.getQueueKey(job.Queue)
	err = r.client.LPush(ctx, queueKey, jobData).Err()
	if err != nil {
		r.updateStats("enqueue", time.Now(), err)
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

// Dequeue removes and returns a job from the queue
func (r *RedisListsBackend) Dequeue(ctx context.Context, opts DequeueOptions) (*Job, error) {
	r.updateStats("dequeue", time.Now(), nil)

	queueKey := r.getQueueKey(r.queueName)

	var result []string
	var err error

	if opts.Timeout > 0 {
		// Blocking pop
		result, err = r.client.BRPop(ctx, opts.Timeout, queueKey).Result()
	} else {
		// Non-blocking pop
		jobData, err := r.client.RPop(ctx, queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				return nil, nil // No jobs available
			}
			r.updateStats("dequeue", time.Now(), err)
			return nil, fmt.Errorf("failed to dequeue job: %w", err)
		}
		result = []string{queueKey, jobData}
	}

	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout or no jobs
		}
		r.updateStats("dequeue", time.Now(), err)
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	// Deserialize job
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		r.updateStats("dequeue", time.Now(), err)
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	return &job, nil
}

// Ack acknowledges successful processing of a job
func (r *RedisListsBackend) Ack(ctx context.Context, jobID string) error {
	// Redis Lists don't have native acknowledgment
	// In a full implementation, we'd track processing jobs separately
	return nil
}

// Nack negatively acknowledges a job and optionally requeues it
func (r *RedisListsBackend) Nack(ctx context.Context, jobID string, requeue bool) error {
	// Redis Lists don't have native negative acknowledgment
	// In a full implementation, we'd handle requeuing logic here
	return nil
}

// Length returns the number of jobs in the queue
func (r *RedisListsBackend) Length(ctx context.Context) (int64, error) {
	queueKey := r.getQueueKey(r.queueName)
	length, err := r.client.LLen(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	r.statsLock.Lock()
	r.stats.QueueDepth = length
	r.statsLock.Unlock()

	return length, nil
}

// Peek looks at a job without removing it
func (r *RedisListsBackend) Peek(ctx context.Context, offset int64) (*Job, error) {
	queueKey := r.getQueueKey(r.queueName)

	// Use LINDEX to peek at specific position
	jobData, err := r.client.LIndex(ctx, queueKey, offset).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No job at that position
		}
		return nil, fmt.Errorf("failed to peek at job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	return &job, nil
}

// Move transfers a job to another queue
func (r *RedisListsBackend) Move(ctx context.Context, jobID string, targetQueue string) error {
	// This is a simplified implementation
	// In practice, we'd need to find the job by ID and move it atomically
	return fmt.Errorf("move operation not implemented for Redis Lists backend")
}

// Iter returns an iterator for jobs in the queue
func (r *RedisListsBackend) Iter(ctx context.Context, opts IterOptions) (Iterator, error) {
	queueKey := r.getQueueKey(r.queueName)

	// Get all jobs from the list
	var jobs []string
	var err error

	if opts.Count > 0 {
		jobs, err = r.client.LRange(ctx, queueKey, 0, opts.Count-1).Result()
	} else {
		jobs, err = r.client.LRange(ctx, queueKey, 0, -1).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for iteration: %w", err)
	}

	// Convert to Job structs
	jobList := make([]*Job, 0, len(jobs))
	for _, jobData := range jobs {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue // Skip malformed jobs
		}
		jobList = append(jobList, &job)
	}

	return NewJobIterator(jobList), nil
}

// Capabilities returns the capabilities of Redis Lists backend
func (r *RedisListsBackend) Capabilities() BackendCapabilities {
	return BackendCapabilities{
		AtomicAck:         false, // Best effort with Lua scripts
		ConsumerGroups:    false,
		Replay:            false,
		IdempotentEnqueue: false,
		Transactions:      true, // Via Lua scripts
		Persistence:       true,
		Clustering:        r.config.ClusterMode,
		TimeToLive:        false,
		Prioritization:    false, // Via separate queues
		BatchOperations:   true,
	}
}

// Stats returns backend performance statistics
func (r *RedisListsBackend) Stats(ctx context.Context) (*BackendStats, error) {
	r.statsLock.RLock()
	defer r.statsLock.RUnlock()

	// Get current queue length
	length, _ := r.Length(ctx)

	stats := *r.stats
	stats.QueueDepth = length

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

// Health performs a health check on the Redis connection
func (r *RedisListsBackend) Health(ctx context.Context) HealthStatus {
	status := HealthStatus{
		CheckedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Ping Redis
	start := time.Now()
	err := r.client.Ping(ctx).Err()
	latency := time.Since(start)

	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Error = err
		status.Message = "Redis ping failed"
		return status
	}

	status.Metadata["ping_latency"] = latency.String()

	// Check queue length
	if r.queueName != "" {
		length, err := r.Length(ctx)
		if err != nil {
			status.Status = HealthStatusDegraded
			status.Message = "Cannot get queue length"
		} else {
			status.Metadata["queue_length"] = fmt.Sprintf("%d", length)
			if length > 10000 {
				status.Status = HealthStatusDegraded
				status.Message = "Queue length is very high"
			} else {
				status.Status = HealthStatusHealthy
			}
		}
	} else {
		status.Status = HealthStatusHealthy
	}

	return status
}

// Close closes the Redis connection
func (r *RedisListsBackend) Close() error {
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// SetQueueName sets the queue name for this backend instance
func (r *RedisListsBackend) SetQueueName(queueName string) {
	r.queueName = queueName
}

// Helper methods

func (r *RedisListsBackend) getQueueKey(queueName string) string {
	if queueName == "" {
		queueName = r.queueName
	}
	return r.keyPrefix + queueName
}

func (r *RedisListsBackend) updateStats(operation string, timestamp time.Time, err error) {
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

// Initialize Redis Lists backend in the registry
func init() {
	RegisterBackend(BackendTypeRedisLists, &RedisListsFactory{})
}
