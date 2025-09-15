// Copyright 2025 James Ross
package deduplication

import (
	"time"
)

// Chunk represents a deduplicated data chunk
type Chunk struct {
	Hash      []byte    `json:"hash"`       // SHA-256 hash of the chunk data
	Data      []byte    `json:"data"`       // Compressed chunk data
	Size      int       `json:"size"`       // Original size before compression
	CompSize  int       `json:"comp_size"`  // Compressed size
	RefCount  int64     `json:"ref_count"`  // Reference count for GC
	CreatedAt time.Time `json:"created_at"` // Creation timestamp
	LastUsed  time.Time `json:"last_used"`  // Last access timestamp
}

// ChunkReference represents a reference to a chunk in a payload
type ChunkReference struct {
	Hash   []byte `json:"hash"`   // Hash of the referenced chunk
	Offset int    `json:"offset"` // Offset in the original payload
	Size   int    `json:"size"`   // Size of this chunk in original payload
}

// PayloadMap represents a deduplicated payload as chunk references
type PayloadMap struct {
	JobID      string           `json:"job_id"`      // Original job ID
	OrigSize   int              `json:"orig_size"`   // Original payload size
	ChunkRefs  []ChunkReference `json:"chunk_refs"`  // References to chunks
	Checksum   []byte           `json:"checksum"`    // Integrity checksum
	CreatedAt  time.Time        `json:"created_at"`  // Creation timestamp
	Compressed bool             `json:"compressed"`  // Whether chunks are compressed
}

// DeduplicationStats represents system-wide deduplication statistics
type DeduplicationStats struct {
	TotalPayloads    int64   `json:"total_payloads"`    // Total payloads processed
	TotalChunks      int64   `json:"total_chunks"`      // Total unique chunks
	TotalBytes       int64   `json:"total_bytes"`       // Total original bytes
	DeduplicatedBytes int64   `json:"deduplicated_bytes"` // Bytes after deduplication
	CompressionRatio float64 `json:"compression_ratio"` // Overall compression ratio
	DeduplicationRatio float64 `json:"deduplication_ratio"` // Deduplication effectiveness
	MemorySavings    int64   `json:"memory_savings"`    // Bytes saved
	SavingsPercent   float64 `json:"savings_percent"`   // Percentage saved
	ChunkHitRate     float64 `json:"chunk_hit_rate"`    // Cache hit rate for chunks
	AvgChunkSize     float64 `json:"avg_chunk_size"`    // Average chunk size
	PopularChunks    []ChunkStats `json:"popular_chunks"` // Most referenced chunks
	LastUpdated      time.Time `json:"last_updated"`     // Stats timestamp
}

// ChunkStats represents statistics for a specific chunk
type ChunkStats struct {
	Hash      string    `json:"hash"`       // Hex-encoded chunk hash
	RefCount  int64     `json:"ref_count"`  // Current reference count
	Size      int       `json:"size"`       // Chunk size
	CreatedAt time.Time `json:"created_at"` // Creation time
	LastUsed  time.Time `json:"last_used"`  // Last access time
	HitCount  int64     `json:"hit_count"`  // Number of times referenced
}

// RollingHash represents a rolling hash state for content-based chunking
type RollingHash struct {
	hash      uint64
	window    []byte
	windowPos int
	polynomial uint64
	windowSize int
}

// ChunkingConfig represents configuration for the chunking algorithm
type ChunkingConfig struct {
	MinChunkSize    int     `json:"min_chunk_size"`    // Minimum chunk size (default: 1KB)
	MaxChunkSize    int     `json:"max_chunk_size"`    // Maximum chunk size (default: 64KB)
	AvgChunkSize    int     `json:"avg_chunk_size"`    // Target average chunk size (default: 8KB)
	WindowSize      int     `json:"window_size"`       // Rolling hash window size
	Polynomial      uint64  `json:"polynomial"`        // Rolling hash polynomial
	SimilarityThreshold float64 `json:"similarity_threshold"` // Similarity detection threshold
}

// CompressionConfig represents compression settings
type CompressionConfig struct {
	Enabled      bool   `json:"enabled"`       // Enable compression
	Level        int    `json:"level"`         // Compression level (1-19 for zstd)
	DictionarySize int   `json:"dictionary_size"` // Compression dictionary size
	UseDictionary bool   `json:"use_dictionary"`  // Use learned compression dictionary
}

// GarbageCollectionConfig represents GC settings
type GarbageCollectionConfig struct {
	Enabled          bool          `json:"enabled"`           // Enable automatic GC
	Interval         time.Duration `json:"interval"`          // GC run interval
	OrphanThreshold  time.Duration `json:"orphan_threshold"`  // Time before chunks are considered orphaned
	BatchSize        int           `json:"batch_size"`        // GC batch processing size
	ConcurrentWorkers int          `json:"concurrent_workers"` // Number of GC workers
}

// Config represents the overall deduplication system configuration
type Config struct {
	Enabled         bool                     `json:"enabled"`          // Enable deduplication
	RedisKeyPrefix  string                   `json:"redis_key_prefix"` // Redis key prefix
	RedisDB         int                      `json:"redis_db"`         // Redis database number
	Chunking        ChunkingConfig           `json:"chunking"`         // Chunking configuration
	Compression     CompressionConfig        `json:"compression"`      // Compression settings
	GarbageCollection GarbageCollectionConfig `json:"garbage_collection"` // GC settings
	SafetyMode      bool                     `json:"safety_mode"`      // Fallback on errors
	MigrationRatio  float64                  `json:"migration_ratio"`  // Gradual migration percentage
	MaxMemoryMB     int64                    `json:"max_memory_mb"`    // Maximum memory usage
	StatsInterval   time.Duration            `json:"stats_interval"`   // Statistics update interval
}

// DeduplicationManager defines the main interface for the deduplication system
type DeduplicationManager interface {
	// Core operations
	DeduplicatePayload(jobID string, payload []byte) (*PayloadMap, error)
	ReconstructPayload(payloadMap *PayloadMap) ([]byte, error)

	// Chunk management
	StoreChunk(chunk *Chunk) error
	GetChunk(hash []byte) (*Chunk, error)
	DeleteChunk(hash []byte) error

	// Reference counting
	AddReference(chunkHash []byte) error
	RemoveReference(chunkHash []byte) error
	GetReferenceCount(chunkHash []byte) (int64, error)

	// Statistics and monitoring
	GetStats() (*DeduplicationStats, error)
	GetChunkStats(hash []byte) (*ChunkStats, error)
	GetPopularChunks(limit int) ([]ChunkStats, error)

	// Garbage collection
	RunGarbageCollection() error
	GetOrphanedChunks() ([][]byte, error)

	// Health and diagnostics
	ValidateIntegrity() error
	AuditReferences() error
	GetHealth() map[string]interface{}

	// Configuration
	UpdateConfig(config *Config) error
	GetConfig() *Config
}

// ChunkStore defines the interface for chunk storage operations
type ChunkStore interface {
	Store(chunk *Chunk) error
	Get(hash []byte) (*Chunk, error)
	Delete(hash []byte) error
	Exists(hash []byte) (bool, error)
	List(prefix []byte, limit int) ([][]byte, error)
	UpdateLastUsed(hash []byte) error
}

// ReferenceCounter defines the interface for reference counting operations
type ReferenceCounter interface {
	Add(chunkHash []byte) (int64, error)
	Remove(chunkHash []byte) (int64, error)
	Get(chunkHash []byte) (int64, error)
	GetAll() (map[string]int64, error)
	AuditAndRepair() error
}

// Chunker defines the interface for payload chunking operations
type Chunker interface {
	ChunkPayload(data []byte) ([]Chunk, error)
	FindSimilarPayloads(data []byte) ([]string, error)
	ComputeChecksum(data []byte) []byte
	ValidateChecksum(data []byte, checksum []byte) bool
}

// Compressor defines the interface for compression operations
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	BuildDictionary(samples [][]byte) error
	GetCompressionRatio() float64
}

// SimilarityDetector defines the interface for similarity detection
type SimilarityDetector interface {
	ComputeSignature(data []byte) ([]uint64, error)
	FindSimilar(signature []uint64, threshold float64) ([]string, error)
	AddToIndex(id string, signature []uint64) error
	RemoveFromIndex(id string) error
}

// PayloadEvent represents events in the deduplication system
type PayloadEvent struct {
	Type      EventType              `json:"type"`
	JobID     string                 `json:"job_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
}

// EventType represents the type of deduplication event
type EventType string

const (
	EventDeduplicationStarted   EventType = "deduplication_started"
	EventDeduplicationCompleted EventType = "deduplication_completed"
	EventReconstructionStarted  EventType = "reconstruction_started"
	EventReconstructionCompleted EventType = "reconstruction_completed"
	EventChunkCreated           EventType = "chunk_created"
	EventChunkReused            EventType = "chunk_reused"
	EventChunkDeleted           EventType = "chunk_deleted"
	EventGarbageCollectionRun   EventType = "garbage_collection_run"
	EventErrorOccurred          EventType = "error_occurred"
	EventConfigUpdated          EventType = "config_updated"
)

// Default configuration values
const (
	DefaultMinChunkSize     = 1024      // 1KB
	DefaultMaxChunkSize     = 65536     // 64KB
	DefaultAvgChunkSize     = 8192      // 8KB
	DefaultWindowSize       = 64        // 64 bytes
	DefaultPolynomial       = 0x82f63b78 // Rabin polynomial
	DefaultCompressionLevel = 3         // zstd level 3
	DefaultDictionarySize   = 65536     // 64KB dictionary
	DefaultGCInterval       = time.Hour // 1 hour GC interval
	DefaultOrphanThreshold  = 24 * time.Hour // 24 hours orphan threshold
	DefaultBatchSize        = 1000      // 1000 chunks per GC batch
	DefaultConcurrentWorkers = 4        // 4 GC workers
	DefaultStatsInterval    = 5 * time.Minute // 5 minute stats update
	DefaultMaxMemoryMB      = 1024      // 1GB max memory
)

// Redis key prefixes
const (
	ChunkKeyPrefix     = "dedup:chunk:"
	RefCountKeyPrefix  = "dedup:refs:"
	PayloadKeyPrefix   = "dedup:payload:"
	StatsKeyPrefix     = "dedup:stats:"
	IndexKeyPrefix     = "dedup:index:"
	ConfigKeyPrefix    = "dedup:config:"
	EventKeyPrefix     = "dedup:event:"
)

// Error types
type DeduplicationError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DeduplicationError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Common error codes
const (
	ErrCodeChunkNotFound       = "CHUNK_NOT_FOUND"
	ErrCodePayloadNotFound     = "PAYLOAD_NOT_FOUND"
	ErrCodeChecksumMismatch    = "CHECKSUM_MISMATCH"
	ErrCodeCompressionFailed   = "COMPRESSION_FAILED"
	ErrCodeDecompressionFailed = "DECOMPRESSION_FAILED"
	ErrCodeInvalidConfig       = "INVALID_CONFIG"
	ErrCodeStorageFull         = "STORAGE_FULL"
	ErrCodeReferenceCorruption = "REFERENCE_CORRUPTION"
	ErrCodeGCFailed            = "GC_FAILED"
)

// Helper functions for error creation
func NewDeduplicationError(code, message string, cause error) *DeduplicationError {
	return &DeduplicationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}