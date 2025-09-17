// Copyright 2025 James Ross
package deduplication

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisChunkStore implements chunk storage using Redis as the backend
type RedisChunkStore struct {
	redis      redis.Cmdable
	keyPrefix  string
	ctx        context.Context
	compressor Compressor
}

// NewRedisChunkStore creates a new Redis-based chunk store
func NewRedisChunkStore(rdb redis.Cmdable, keyPrefix string, compressor Compressor) *RedisChunkStore {
	return &RedisChunkStore{
		redis:      rdb,
		keyPrefix:  keyPrefix,
		ctx:        context.Background(),
		compressor: compressor,
	}
}

// Store saves a chunk in Redis with compression
func (rcs *RedisChunkStore) Store(chunk *Chunk) error {
	hashStr := hex.EncodeToString(chunk.Hash)
	key := rcs.keyPrefix + ChunkKeyPrefix + hashStr

	// Check if chunk already exists
	exists, err := rcs.Exists(chunk.Hash)
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to check chunk existence", err)
	}

	if exists {
		// Update last used time
		return rcs.UpdateLastUsed(chunk.Hash)
	}

	// Compress chunk data if compressor is available
	data := chunk.Data
	if rcs.compressor != nil {
		compressed, err := rcs.compressor.Compress(chunk.Data)
		if err != nil {
			return NewDeduplicationError(ErrCodeCompressionFailed, "failed to compress chunk", err)
		}
		data = compressed
		chunk.CompSize = len(compressed)
	}

	// Create chunk metadata
	metadata := map[string]interface{}{
		"size":       chunk.Size,
		"comp_size":  len(data),
		"created_at": chunk.CreatedAt.Unix(),
		"last_used":  chunk.LastUsed.Unix(),
		"ref_count":  chunk.RefCount,
	}

	// Store using pipeline for atomicity
	pipe := rcs.redis.Pipeline()

	// Store compressed data
	pipe.Set(rcs.ctx, key, data, 0)

	// Store metadata
	metaKey := key + ":meta"
	pipe.HMSet(rcs.ctx, metaKey, metadata)

	// Execute pipeline
	_, err = pipe.Exec(rcs.ctx)
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to store chunk", err)
	}

	return nil
}

// Get retrieves a chunk from Redis and decompresses it
func (rcs *RedisChunkStore) Get(hash []byte) (*Chunk, error) {
	hashStr := hex.EncodeToString(hash)
	key := rcs.keyPrefix + ChunkKeyPrefix + hashStr
	metaKey := key + ":meta"

	// Get chunk data and metadata in parallel
	pipe := rcs.redis.Pipeline()
	dataCmd := pipe.Get(rcs.ctx, key)
	metaCmd := pipe.HMGet(rcs.ctx, metaKey, "size", "comp_size", "created_at", "last_used", "ref_count")

	_, err := pipe.Exec(rcs.ctx)
	if err != nil {
		if err == redis.Nil {
			return nil, NewDeduplicationError(ErrCodeChunkNotFound, "chunk not found", nil)
		}
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to get chunk", err)
	}

	// Get compressed data
	compressedData, err := dataCmd.Result()
	if err != nil {
		return nil, NewDeduplicationError(ErrCodeChunkNotFound, "chunk data not found", err)
	}

	// Parse metadata
	metaVals, err := metaCmd.Result()
	if err != nil {
		return nil, NewDeduplicationError(ErrCodeChunkNotFound, "chunk metadata not found", err)
	}

	chunk := &Chunk{
		Hash: make([]byte, len(hash)),
	}
	copy(chunk.Hash, hash)

	// Parse size
	if metaVals[0] != nil {
		if sizeStr, ok := metaVals[0].(string); ok {
			if size, err := strconv.Atoi(sizeStr); err == nil {
				chunk.Size = size
			}
		}
	}

	// Parse compressed size
	if metaVals[1] != nil {
		if compSizeStr, ok := metaVals[1].(string); ok {
			if compSize, err := strconv.Atoi(compSizeStr); err == nil {
				chunk.CompSize = compSize
			}
		}
	}

	// Parse created_at
	if metaVals[2] != nil {
		if createdStr, ok := metaVals[2].(string); ok {
			if createdUnix, err := strconv.ParseInt(createdStr, 10, 64); err == nil {
				chunk.CreatedAt = time.Unix(createdUnix, 0)
			}
		}
	}

	// Parse last_used
	if metaVals[3] != nil {
		if usedStr, ok := metaVals[3].(string); ok {
			if usedUnix, err := strconv.ParseInt(usedStr, 10, 64); err == nil {
				chunk.LastUsed = time.Unix(usedUnix, 0)
			}
		}
	}

	// Parse ref_count
	if metaVals[4] != nil {
		if refStr, ok := metaVals[4].(string); ok {
			if refCount, err := strconv.ParseInt(refStr, 10, 64); err == nil {
				chunk.RefCount = refCount
			}
		}
	}

	// Decompress data
	data := []byte(compressedData)
	if rcs.compressor != nil {
		decompressed, err := rcs.compressor.Decompress(data)
		if err != nil {
			return nil, NewDeduplicationError(ErrCodeDecompressionFailed, "failed to decompress chunk", err)
		}
		data = decompressed
	}

	chunk.Data = data

	// Update last used time asynchronously
	go rcs.UpdateLastUsed(hash)

	return chunk, nil
}

// Delete removes a chunk from Redis
func (rcs *RedisChunkStore) Delete(hash []byte) error {
	hashStr := hex.EncodeToString(hash)
	key := rcs.keyPrefix + ChunkKeyPrefix + hashStr
	metaKey := key + ":meta"

	// Delete both data and metadata
	pipe := rcs.redis.Pipeline()
	pipe.Del(rcs.ctx, key)
	pipe.Del(rcs.ctx, metaKey)

	_, err := pipe.Exec(rcs.ctx)
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to delete chunk", err)
	}

	return nil
}

// Exists checks if a chunk exists in storage
func (rcs *RedisChunkStore) Exists(hash []byte) (bool, error) {
	hashStr := hex.EncodeToString(hash)
	key := rcs.keyPrefix + ChunkKeyPrefix + hashStr

	exists, err := rcs.redis.Exists(rcs.ctx, key).Result()
	if err != nil {
		return false, NewDeduplicationError(ErrCodeStorageFull, "failed to check chunk existence", err)
	}

	return exists > 0, nil
}

// List returns chunk hashes with the given prefix
func (rcs *RedisChunkStore) List(prefix []byte, limit int) ([][]byte, error) {
	prefixStr := hex.EncodeToString(prefix)
	pattern := rcs.keyPrefix + ChunkKeyPrefix + prefixStr + "*"

	var keys []string
	var err error

	if limit > 0 {
		// Use SCAN for large result sets
		iter := rcs.redis.Scan(rcs.ctx, 0, pattern, int64(limit)).Iterator()
		for iter.Next(rcs.ctx) {
			key := iter.Val()
			// Extract just the chunk part (remove prefix and meta suffix)
			if len(key) > len(rcs.keyPrefix+ChunkKeyPrefix) && !endsWith(key, ":meta") {
				keys = append(keys, key)
			}
			if len(keys) >= limit {
				break
			}
		}
		err = iter.Err()
	} else {
		// Use KEYS for small result sets
		allKeys, keyErr := rcs.redis.Keys(rcs.ctx, pattern).Result()
		err = keyErr
		for _, key := range allKeys {
			if !endsWith(key, ":meta") {
				keys = append(keys, key)
			}
		}
	}

	if err != nil {
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to list chunks", err)
	}

	// Convert keys back to hashes
	hashes := make([][]byte, 0, len(keys))
	prefixLen := len(rcs.keyPrefix + ChunkKeyPrefix)

	for _, key := range keys {
		if len(key) > prefixLen {
			hashStr := key[prefixLen:]
			if hash, err := hex.DecodeString(hashStr); err == nil {
				hashes = append(hashes, hash)
			}
		}
	}

	return hashes, nil
}

// UpdateLastUsed updates the last used timestamp for a chunk
func (rcs *RedisChunkStore) UpdateLastUsed(hash []byte) error {
	hashStr := hex.EncodeToString(hash)
	metaKey := rcs.keyPrefix + ChunkKeyPrefix + hashStr + ":meta"

	err := rcs.redis.HSet(rcs.ctx, metaKey, "last_used", time.Now().Unix()).Err()
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to update last used time", err)
	}

	return nil
}

// RedisReferenceCounter implements reference counting using Redis
type RedisReferenceCounter struct {
	redis     redis.Cmdable
	keyPrefix string
	refKey    string
	ctx       context.Context
}

// NewRedisReferenceCounter creates a new Redis-based reference counter
func NewRedisReferenceCounter(rdb redis.Cmdable, keyPrefix string) *RedisReferenceCounter {
	return &RedisReferenceCounter{
		redis:     rdb,
		keyPrefix: keyPrefix,
		refKey:    keyPrefix + RefCountKeyPrefix + "global",
		ctx:       context.Background(),
	}
}

// Add increments the reference count for a chunk
func (rrc *RedisReferenceCounter) Add(chunkHash []byte) (int64, error) {
	hashStr := hex.EncodeToString(chunkHash)

	count, err := rrc.redis.HIncrBy(rrc.ctx, rrc.refKey, hashStr, 1).Result()
	if err != nil {
		return 0, NewDeduplicationError(ErrCodeReferenceCorruption, "failed to increment reference count", err)
	}

	// Refresh expiry on the reference hash
	rrc.redis.Expire(rrc.ctx, rrc.refKey, 7*24*time.Hour)

	return count, nil
}

// Remove decrements the reference count for a chunk
func (rrc *RedisReferenceCounter) Remove(chunkHash []byte) (int64, error) {
	hashStr := hex.EncodeToString(chunkHash)

	count, err := rrc.redis.HIncrBy(rrc.ctx, rrc.refKey, hashStr, -1).Result()
	if err != nil {
		return 0, NewDeduplicationError(ErrCodeReferenceCorruption, "failed to decrement reference count", err)
	}

	// Remove entry if count reaches zero or below
	if count <= 0 {
		rrc.redis.HDel(rrc.ctx, rrc.refKey, hashStr)
		return 0, nil
	}

	return count, nil
}

// Get returns the current reference count for a chunk
func (rrc *RedisReferenceCounter) Get(chunkHash []byte) (int64, error) {
	hashStr := hex.EncodeToString(chunkHash)

	countStr, err := rrc.redis.HGet(rrc.ctx, rrc.refKey, hashStr).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // No references exist
		}
		return 0, NewDeduplicationError(ErrCodeReferenceCorruption, "failed to get reference count", err)
	}

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, NewDeduplicationError(ErrCodeReferenceCorruption, "invalid reference count format", err)
	}

	return count, nil
}

// GetAll returns all reference counts
func (rrc *RedisReferenceCounter) GetAll() (map[string]int64, error) {
	allRefs, err := rrc.redis.HGetAll(rrc.ctx, rrc.refKey).Result()
	if err != nil {
		return nil, NewDeduplicationError(ErrCodeReferenceCorruption, "failed to get all reference counts", err)
	}

	result := make(map[string]int64, len(allRefs))

	for hashStr, countStr := range allRefs {
		if count, err := strconv.ParseInt(countStr, 10, 64); err == nil {
			result[hashStr] = count
		}
	}

	return result, nil
}

// AuditAndRepair checks reference count consistency and repairs if needed
func (rrc *RedisReferenceCounter) AuditAndRepair() error {
	// Get all reference counts
	allRefs, err := rrc.GetAll()
	if err != nil {
		return err
	}

	// Scan all payload maps to build actual reference counts
	actualRefs := make(map[string]int64)

	// Find all payload map keys
	payloadPattern := rrc.keyPrefix + PayloadKeyPrefix + "*"
	iter := rrc.redis.Scan(rrc.ctx, 0, payloadPattern, 1000).Iterator()

	for iter.Next(rrc.ctx) {
		payloadKey := iter.Val()

		// Get payload map
		payloadData, err := rrc.redis.Get(rrc.ctx, payloadKey).Result()
		if err != nil {
			continue // Skip if can't read
		}

		var payloadMap PayloadMap
		if err := json.Unmarshal([]byte(payloadData), &payloadMap); err != nil {
			continue // Skip if can't parse
		}

		// Count chunk references
		for _, chunkRef := range payloadMap.ChunkRefs {
			hashStr := hex.EncodeToString(chunkRef.Hash)
			actualRefs[hashStr]++
		}
	}

	if err := iter.Err(); err != nil {
		return NewDeduplicationError(ErrCodeReferenceCorruption, "failed to scan payload maps", err)
	}

	// Compare and repair mismatches
	repaired := 0

	for hashStr, actualCount := range actualRefs {
		storedCount := allRefs[hashStr]

		if storedCount != actualCount {
			// Fix the reference count
			err := rrc.redis.HSet(rrc.ctx, rrc.refKey, hashStr, actualCount).Err()
			if err != nil {
				return NewDeduplicationError(ErrCodeReferenceCorruption, "failed to repair reference count", err)
			}
			repaired++
		}
	}

	// Remove orphaned reference counts (no actual references)
	for hashStr, storedCount := range allRefs {
		if actualRefs[hashStr] == 0 && storedCount > 0 {
			err := rrc.redis.HDel(rrc.ctx, rrc.refKey, hashStr).Err()
			if err != nil {
				return NewDeduplicationError(ErrCodeReferenceCorruption, "failed to remove orphaned reference", err)
			}
			repaired++
		}
	}

	return nil
}

// PayloadMapStore handles storage of payload maps
type PayloadMapStore struct {
	redis     redis.Cmdable
	keyPrefix string
	ctx       context.Context
}

// NewPayloadMapStore creates a new payload map store
func NewPayloadMapStore(rdb redis.Cmdable, keyPrefix string) *PayloadMapStore {
	return &PayloadMapStore{
		redis:     rdb,
		keyPrefix: keyPrefix,
		ctx:       context.Background(),
	}
}

// Store saves a payload map
func (pms *PayloadMapStore) Store(payloadMap *PayloadMap) error {
	key := pms.keyPrefix + PayloadKeyPrefix + payloadMap.JobID

	data, err := json.Marshal(payloadMap)
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to marshal payload map", err)
	}

	err = pms.redis.Set(pms.ctx, key, data, 7*24*time.Hour).Err() // 7 day expiry
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to store payload map", err)
	}

	return nil
}

// Get retrieves a payload map
func (pms *PayloadMapStore) Get(jobID string) (*PayloadMap, error) {
	key := pms.keyPrefix + PayloadKeyPrefix + jobID

	data, err := pms.redis.Get(pms.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NewDeduplicationError(ErrCodePayloadNotFound, "payload map not found", nil)
		}
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to get payload map", err)
	}

	var payloadMap PayloadMap
	err = json.Unmarshal([]byte(data), &payloadMap)
	if err != nil {
		return nil, NewDeduplicationError(ErrCodePayloadNotFound, "failed to unmarshal payload map", err)
	}

	return &payloadMap, nil
}

// Delete removes a payload map
func (pms *PayloadMapStore) Delete(jobID string) error {
	key := pms.keyPrefix + PayloadKeyPrefix + jobID

	err := pms.redis.Del(pms.ctx, key).Err()
	if err != nil {
		return NewDeduplicationError(ErrCodeStorageFull, "failed to delete payload map", err)
	}

	return nil
}

// List returns all payload map job IDs
func (pms *PayloadMapStore) List(limit int) ([]string, error) {
	pattern := pms.keyPrefix + PayloadKeyPrefix + "*"

	var keys []string
	iter := pms.redis.Scan(pms.ctx, 0, pattern, int64(limit)).Iterator()

	for iter.Next(pms.ctx) {
		key := iter.Val()
		// Extract job ID from key
		prefixLen := len(pms.keyPrefix + PayloadKeyPrefix)
		if len(key) > prefixLen {
			jobID := key[prefixLen:]
			keys = append(keys, jobID)
		}

		if limit > 0 && len(keys) >= limit {
			break
		}
	}

	if err := iter.Err(); err != nil {
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to list payload maps", err)
	}

	return keys, nil
}

// Helper function to check string suffix
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
