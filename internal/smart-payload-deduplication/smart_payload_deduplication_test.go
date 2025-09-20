//go:build smart_payload_dedup_tests
// +build smart_payload_dedup_tests

// Copyright 2025 James Ross
package deduplication

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestRabinChunker_ChunkPayload(t *testing.T) {
	config := &ChunkingConfig{
		MinChunkSize:        256,
		MaxChunkSize:        8192,
		AvgChunkSize:        1024,
		WindowSize:          64,
		Polynomial:          0x82f63b78,
		SimilarityThreshold: 0.8,
	}

	detector := NewMinHashSimilarityDetector(64, 8)
	chunker := NewRabinChunker(config, detector)

	tests := []struct {
		name      string
		data      []byte
		minChunks int
		maxChunks int
	}{
		{
			name:      "small payload",
			data:      []byte("hello world"),
			minChunks: 1,
			maxChunks: 1,
		},
		{
			name:      "medium payload",
			data:      bytes.Repeat([]byte("test data chunk content "), 100),
			minChunks: 1,
			maxChunks: 5,
		},
		{
			name:      "large payload",
			data:      bytes.Repeat([]byte("large chunk of data with repetitive content "), 1000),
			minChunks: 5,
			maxChunks: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := chunker.ChunkPayload(tt.data)
			if err != nil {
				t.Fatalf("ChunkPayload failed: %v", err)
			}

			if len(chunks) < tt.minChunks || len(chunks) > tt.maxChunks {
				t.Errorf("expected %d-%d chunks, got %d", tt.minChunks, tt.maxChunks, len(chunks))
			}

			// Verify chunk sizes
			for i, chunk := range chunks {
				if chunk.Size < config.MinChunkSize && i < len(chunks)-1 {
					t.Errorf("chunk %d size %d below minimum %d", i, chunk.Size, config.MinChunkSize)
				}

				if chunk.Size > config.MaxChunkSize {
					t.Errorf("chunk %d size %d above maximum %d", i, chunk.Size, config.MaxChunkSize)
				}

				if len(chunk.Data) != chunk.Size {
					t.Errorf("chunk %d data length %d != size %d", i, len(chunk.Data), chunk.Size)
				}

				if len(chunk.Hash) != 32 {
					t.Errorf("chunk %d hash length %d != 32", i, len(chunk.Hash))
				}
			}

			// Verify chunks can be reassembled
			reassembled := make([]byte, 0, len(tt.data))
			for _, chunk := range chunks {
				reassembled = append(reassembled, chunk.Data...)
			}

			if !bytes.Equal(tt.data, reassembled) {
				t.Error("reassembled data does not match original")
			}
		})
	}
}

func TestRollingHash_Consistency(t *testing.T) {
	windowSize := 64
	polynomial := uint64(0x82f63b78)

	data := []byte("this is a test string for rolling hash consistency")

	// Create two rolling hash instances
	rh1 := NewRollingHash(windowSize, polynomial)
	rh2 := NewRollingHash(windowSize, polynomial)

	// Feed same data to both
	for _, b := range data {
		rh1.Roll(b)
		rh2.Roll(b)
	}

	if rh1.Sum() != rh2.Sum() {
		t.Error("rolling hash instances produced different results for same input")
	}

	// Test reset functionality
	rh1.Reset()
	if rh1.Sum() != 0 {
		t.Error("rolling hash not reset to zero")
	}
}

func TestZstdCompressor_CompressDecompress(t *testing.T) {
	config := &CompressionConfig{
		Enabled:        true,
		Level:          3,
		DictionarySize: 1024,
		UseDictionary:  false, // Start without dictionary
	}

	compressor, err := NewZstdCompressor(config)
	if err != nil {
		t.Fatalf("Failed to create compressor: %v", err)
	}
	defer compressor.Close()

	testData := [][]byte{
		[]byte("hello world"),
		[]byte("this is a longer test string with more content to compress"),
		bytes.Repeat([]byte("repetitive content "), 100),
		make([]byte, 10000), // Zero-filled data
	}

	for i, data := range testData {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			// Compress
			compressed, err := compressor.Compress(data)
			if err != nil {
				t.Fatalf("Compression failed: %v", err)
			}

			// Decompress
			decompressed, err := compressor.Decompress(compressed)
			if err != nil {
				t.Fatalf("Decompression failed: %v", err)
			}

			// Verify
			if !bytes.Equal(data, decompressed) {
				t.Error("decompressed data does not match original")
			}

			t.Logf("Original: %d bytes, Compressed: %d bytes, Ratio: %.2f",
				len(data), len(compressed), float64(len(compressed))/float64(len(data)))
		})
	}
}

func TestZstdCompressor_WithDictionary(t *testing.T) {
	config := &CompressionConfig{
		Enabled:        true,
		Level:          3,
		DictionarySize: 1024,
		UseDictionary:  true,
	}

	compressor, err := NewZstdCompressor(config)
	if err != nil {
		t.Fatalf("Failed to create compressor: %v", err)
	}
	defer compressor.Close()

	// Build dictionary from samples
	samples := [][]byte{
		[]byte("common prefix with variable suffix 1"),
		[]byte("common prefix with variable suffix 2"),
		[]byte("common prefix with variable suffix 3"),
		[]byte("another common pattern in data 1"),
		[]byte("another common pattern in data 2"),
	}

	err = compressor.BuildDictionary(samples)
	if err != nil {
		t.Fatalf("Failed to build dictionary: %v", err)
	}

	// Test compression with dictionary
	testData := []byte("common prefix with variable suffix 999")

	compressed, err := compressor.Compress(testData)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(testData, decompressed) {
		t.Error("decompressed data does not match original")
	}

	stats := compressor.GetStats()
	if stats.DictionarySize == 0 {
		t.Error("dictionary size should be non-zero")
	}
}

func TestMinHashSimilarityDetector_ComputeSignature(t *testing.T) {
	detector := NewMinHashSimilarityDetector(128, 16)

	data1 := []byte("this is a test document with some content")
	data2 := []byte("this is a test document with different content")
	data3 := []byte("completely different document with no similarity")

	sig1, err := detector.ComputeSignature(data1)
	if err != nil {
		t.Fatalf("Failed to compute signature 1: %v", err)
	}

	sig2, err := detector.ComputeSignature(data2)
	if err != nil {
		t.Fatalf("Failed to compute signature 2: %v", err)
	}

	sig3, err := detector.ComputeSignature(data3)
	if err != nil {
		t.Fatalf("Failed to compute signature 3: %v", err)
	}

	if len(sig1) != 128 || len(sig2) != 128 || len(sig3) != 128 {
		t.Error("signatures should have length 128")
	}

	// Similar documents should have more matching elements
	matches12 := countMatchingElements(sig1, sig2)
	matches13 := countMatchingElements(sig1, sig3)

	if matches12 <= matches13 {
		t.Error("similar documents should have more matching signature elements")
	}

	t.Logf("Matches 1-2: %d, Matches 1-3: %d", matches12, matches13)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "default config",
			config:    DefaultConfig(),
			expectErr: false,
		},
		{
			name: "invalid chunk size",
			config: &Config{
				Chunking: ChunkingConfig{
					MinChunkSize: 0,
					MaxChunkSize: 1000,
					AvgChunkSize: 500,
				},
			},
			expectErr: true,
		},
		{
			name: "max less than min",
			config: &Config{
				Chunking: ChunkingConfig{
					MinChunkSize: 1000,
					MaxChunkSize: 500,
					AvgChunkSize: 750,
				},
			},
			expectErr: true,
		},
		{
			name: "invalid similarity threshold",
			config: &Config{
				Chunking: ChunkingConfig{
					MinChunkSize:        256,
					MaxChunkSize:        8192,
					AvgChunkSize:        1024,
					SimilarityThreshold: 1.5,
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error=%v, got error=%v", tt.expectErr, err != nil)
			}
		})
	}
}

func TestChunkingAnalyzer_AnalyzeChunking(t *testing.T) {
	analyzer := NewChunkingAnalyzer()

	// Create test chunks with some duplicates
	chunks := []Chunk{
		{Hash: []byte("hash1"), Size: 1024},
		{Hash: []byte("hash2"), Size: 2048},
		{Hash: []byte("hash1"), Size: 1024}, // Duplicate
		{Hash: []byte("hash3"), Size: 512},
		{Hash: []byte("hash2"), Size: 2048}, // Duplicate
	}

	analysis := analyzer.AnalyzeChunking(chunks)

	if analysis.TotalChunks != 5 {
		t.Errorf("expected 5 total chunks, got %d", analysis.TotalChunks)
	}

	if analysis.UniqueChunks != 3 {
		t.Errorf("expected 3 unique chunks, got %d", analysis.UniqueChunks)
	}

	if analysis.DuplicateChunks != 2 {
		t.Errorf("expected 2 duplicate chunks, got %d", analysis.DuplicateChunks)
	}

	expectedDedupeRatio := 3.0 / 5.0
	if abs(analysis.DeduplicationRatio-expectedDedupeRatio) > 0.01 {
		t.Errorf("expected dedup ratio %.2f, got %.2f", expectedDedupeRatio, analysis.DeduplicationRatio)
	}
}

// Integration tests (require Redis)

func TestRedisChunkStore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	rdb := setupTestRedis(t)
	defer rdb.Close()

	compressor, _ := NewZstdCompressor(&CompressionConfig{Enabled: false})
	store := NewRedisChunkStore(rdb, "test:", compressor)

	// Test data
	data := []byte("test chunk data for storage")
	hash := sha256.Sum256(data)

	chunk := &Chunk{
		Hash:      hash[:],
		Data:      data,
		Size:      len(data),
		CompSize:  len(data),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	// Store chunk
	err := store.Store(chunk)
	if err != nil {
		t.Fatalf("Failed to store chunk: %v", err)
	}

	// Retrieve chunk
	retrieved, err := store.Get(hash[:])
	if err != nil {
		t.Fatalf("Failed to get chunk: %v", err)
	}

	if !bytes.Equal(chunk.Data, retrieved.Data) {
		t.Error("retrieved chunk data does not match original")
	}

	if !bytes.Equal(chunk.Hash, retrieved.Hash) {
		t.Error("retrieved chunk hash does not match original")
	}

	// Test existence
	exists, err := store.Exists(hash[:])
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}

	if !exists {
		t.Error("chunk should exist")
	}

	// Delete chunk
	err = store.Delete(hash[:])
	if err != nil {
		t.Fatalf("Failed to delete chunk: %v", err)
	}

	// Verify deletion
	exists, err = store.Exists(hash[:])
	if err != nil {
		t.Fatalf("Failed to check existence after deletion: %v", err)
	}

	if exists {
		t.Error("chunk should not exist after deletion")
	}
}

func TestRedisReferenceCounter_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	rdb := setupTestRedis(t)
	defer rdb.Close()

	counter := NewRedisReferenceCounter(rdb, "test:")

	hash := []byte("test_chunk_hash")

	// Add references
	count, err := counter.Add(hash)
	if err != nil {
		t.Fatalf("Failed to add reference: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	count, err = counter.Add(hash)
	if err != nil {
		t.Fatalf("Failed to add second reference: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	// Get reference count
	count, err = counter.Get(hash)
	if err != nil {
		t.Fatalf("Failed to get reference count: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	// Remove reference
	count, err = counter.Remove(hash)
	if err != nil {
		t.Fatalf("Failed to remove reference: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Remove last reference
	count, err = counter.Remove(hash)
	if err != nil {
		t.Fatalf("Failed to remove last reference: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestManager_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	rdb := setupTestRedis(t)
	defer rdb.Close()

	config := DefaultConfig()
	config.RedisKeyPrefix = "test:"

	manager, err := NewManager(config, rdb, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	defer manager.Close()

	// Test payload with repetitive content
	payload := bytes.Repeat([]byte("test data block with some content "), 100)
	jobID := "test_job_1"

	// Deduplicate payload
	payloadMap, err := manager.DeduplicatePayload(jobID, payload)
	if err != nil {
		t.Fatalf("Failed to deduplicate payload: %v", err)
	}

	if payloadMap.JobID != jobID {
		t.Errorf("expected job ID %s, got %s", jobID, payloadMap.JobID)
	}

	if payloadMap.OrigSize != len(payload) {
		t.Errorf("expected original size %d, got %d", len(payload), payloadMap.OrigSize)
	}

	if len(payloadMap.ChunkRefs) == 0 {
		t.Error("expected at least one chunk reference")
	}

	// Reconstruct payload
	reconstructed, err := manager.ReconstructPayload(payloadMap)
	if err != nil {
		t.Fatalf("Failed to reconstruct payload: %v", err)
	}

	if !bytes.Equal(payload, reconstructed) {
		t.Error("reconstructed payload does not match original")
	}

	// Test with duplicate payload (should reuse chunks)
	payloadMap2, err := manager.DeduplicatePayload("test_job_2", payload)
	if err != nil {
		t.Fatalf("Failed to deduplicate duplicate payload: %v", err)
	}

	// Should have same number of chunks (chunks reused)
	if len(payloadMap2.ChunkRefs) != len(payloadMap.ChunkRefs) {
		t.Errorf("expected same chunk count for duplicate payload")
	}

	// Get statistics
	stats, err := manager.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TotalPayloads < 2 {
		t.Errorf("expected at least 2 payloads processed")
	}

	t.Logf("Processed %d payloads, %d chunks, %.2f%% savings",
		stats.TotalPayloads, stats.TotalChunks, stats.SavingsPercent*100)
}

// Benchmark tests

func BenchmarkRabinChunker_ChunkPayload(b *testing.B) {
	config := &ChunkingConfig{
		MinChunkSize: 1024,
		MaxChunkSize: 8192,
		AvgChunkSize: 4096,
		WindowSize:   64,
		Polynomial:   0x82f63b78,
	}

	detector := NewMinHashSimilarityDetector(64, 8)
	chunker := NewRabinChunker(config, detector)

	// Test data: 1MB of repetitive content
	data := bytes.Repeat([]byte("benchmark data with some repetitive content and variation "), 1000)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := chunker.ChunkPayload(data)
		if err != nil {
			b.Fatalf("ChunkPayload failed: %v", err)
		}
	}
}

func BenchmarkZstdCompressor_Compress(b *testing.B) {
	config := &CompressionConfig{
		Enabled: true,
		Level:   3,
	}

	compressor, err := NewZstdCompressor(config)
	if err != nil {
		b.Fatalf("Failed to create compressor: %v", err)
	}
	defer compressor.Close()

	data := bytes.Repeat([]byte("compression benchmark data "), 1000)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := compressor.Compress(data)
		if err != nil {
			b.Fatalf("Compression failed: %v", err)
		}
	}
}

// Helper functions

func setupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test database
	})

	// Clean up test database
	err := rdb.FlushDB(rdb.Context()).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	return rdb
}

func countMatchingElements(sig1, sig2 []uint64) int {
	matches := 0
	for i := 0; i < len(sig1) && i < len(sig2); i++ {
		if sig1[i] == sig2[i] {
			matches++
		}
	}
	return matches
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
