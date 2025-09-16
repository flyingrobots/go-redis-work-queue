// Copyright 2025 James Ross
package deduplication

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/crc32"
	"math/bits"
	"time"
)

// RabinChunker implements content-based chunking using Rabin fingerprinting
type RabinChunker struct {
	config     *ChunkingConfig
	polynomial uint64
	mask       uint64
	modTable   [256]uint64
	detector   SimilarityDetector
}

// NewRabinChunker creates a new Rabin-based chunker
func NewRabinChunker(config *ChunkingConfig, detector SimilarityDetector) *RabinChunker {
	chunker := &RabinChunker{
		config:   config,
		polynomial: config.Polynomial,
		detector: detector,
	}

	// Calculate mask for boundary detection
	bits := 0
	target := config.AvgChunkSize
	for (1 << bits) < target {
		bits++
	}
	chunker.mask = (1 << bits) - 1

	// Precompute modular reduction table for efficiency
	chunker.buildModTable()

	return chunker
}

// buildModTable precomputes the modular reduction table for fast rolling hash
func (rc *RabinChunker) buildModTable() {
	for i := 0; i < 256; i++ {
		val := uint64(i)
		for j := 0; j < 8; j++ {
			if val&1 != 0 {
				val = (val >> 1) ^ rc.polynomial
			} else {
				val >>= 1
			}
		}
		rc.modTable[i] = val
	}
}

// ChunkPayload splits payload into content-defined chunks
func (rc *RabinChunker) ChunkPayload(data []byte) ([]Chunk, error) {
	if len(data) == 0 {
		return []Chunk{}, nil
	}

	chunks := make([]Chunk, 0)
	rollingHash := NewRollingHash(rc.config.WindowSize, rc.polynomial)

	start := 0

	// Initialize rolling hash window
	windowFilled := 0
	for i := 0; i < len(data) && windowFilled < rc.config.WindowSize; i++ {
		rollingHash.Roll(data[i])
		windowFilled++
	}

	for i := rc.config.WindowSize; i < len(data); i++ {
		rollingHash.Roll(data[i])

		// Check for chunk boundary
		if rc.isChunkBoundary(rollingHash.Sum(), i-start) {
			if i-start >= rc.config.MinChunkSize {
				chunk, err := rc.createChunk(data[start:i+1])
				if err != nil {
					return nil, err
				}
				chunks = append(chunks, chunk)
				start = i + 1
			}
		}

		// Force chunk boundary at max size
		if i-start >= rc.config.MaxChunkSize {
			chunk, err := rc.createChunk(data[start:i+1])
			if err != nil {
				return nil, err
			}
			chunks = append(chunks, chunk)
			start = i + 1
		}
	}

	// Handle final chunk
	if start < len(data) {
		chunk, err := rc.createChunk(data[start:])
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// isChunkBoundary determines if the current position is a chunk boundary
func (rc *RabinChunker) isChunkBoundary(hash uint64, currentSize int) bool {
	// Ensure minimum chunk size
	if currentSize < rc.config.MinChunkSize {
		return false
	}

	// Check if hash matches boundary pattern
	return (hash & rc.mask) == 0
}

// createChunk creates a chunk from data with hash and metadata
func (rc *RabinChunker) createChunk(data []byte) (Chunk, error) {
	hash := sha256.Sum256(data)

	chunk := Chunk{
		Hash:      hash[:],
		Data:      make([]byte, len(data)),
		Size:      len(data),
		CompSize:  len(data), // Will be updated after compression
		RefCount:  0,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	copy(chunk.Data, data)

	return chunk, nil
}

// FindSimilarPayloads finds payloads similar to the given data
func (rc *RabinChunker) FindSimilarPayloads(data []byte) ([]string, error) {
	if rc.detector == nil {
		return []string{}, nil
	}

	signature, err := rc.detector.ComputeSignature(data)
	if err != nil {
		return nil, err
	}

	return rc.detector.FindSimilar(signature, rc.config.SimilarityThreshold)
}

// ComputeChecksum computes a checksum for integrity verification
func (rc *RabinChunker) ComputeChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ValidateChecksum validates data against a checksum
func (rc *RabinChunker) ValidateChecksum(data []byte, checksum []byte) bool {
	computed := rc.ComputeChecksum(data)
	if len(computed) != len(checksum) {
		return false
	}

	for i := range computed {
		if computed[i] != checksum[i] {
			return false
		}
	}

	return true
}

// NewRollingHash creates a new rolling hash instance
func NewRollingHash(windowSize int, polynomial uint64) *RollingHash {
	return &RollingHash{
		hash:       0,
		window:     make([]byte, windowSize),
		windowPos:  0,
		polynomial: polynomial,
		windowSize: windowSize,
	}
}

// Roll updates the rolling hash with a new byte
func (rh *RollingHash) Roll(b byte) {
	// Remove the byte that's falling out of the window
	oldByte := rh.window[rh.windowPos]

	// Add the new byte to the window
	rh.window[rh.windowPos] = b
	rh.windowPos = (rh.windowPos + 1) % rh.windowSize

	// Update the hash
	rh.hash = rh.hash<<1 ^ uint64(b)

	// Remove contribution of the old byte
	oldContrib := uint64(oldByte) << uint64(rh.windowSize)
	rh.hash ^= oldContrib

	// Apply polynomial modular reduction
	rh.hash = rh.reduceHash(rh.hash)
}

// Sum returns the current hash value
func (rh *RollingHash) Sum() uint64 {
	return rh.hash
}

// Reset clears the rolling hash state
func (rh *RollingHash) Reset() {
	rh.hash = 0
	rh.windowPos = 0
	for i := range rh.window {
		rh.window[i] = 0
	}
}

// reduceHash applies polynomial modular reduction
func (rh *RollingHash) reduceHash(hash uint64) uint64 {
	// Simple polynomial reduction using XOR
	deg := bits.Len64(rh.polynomial) - 1

	for bits.Len64(hash) > deg {
		shift := bits.Len64(hash) - deg - 1
		hash ^= rh.polynomial << shift
	}

	return hash
}

// MinHashSimilarityDetector implements similarity detection using MinHash LSH
type MinHashSimilarityDetector struct {
	numHashes   int
	bands       int
	rows        int
	hashFuncs   []HashFunction
	lshIndex    map[string][]string
}

// HashFunction represents a hash function for MinHash
type HashFunction struct {
	a, b, c uint64
}

// NewMinHashSimilarityDetector creates a new MinHash-based similarity detector
func NewMinHashSimilarityDetector(numHashes, bands int) *MinHashSimilarityDetector {
	rows := numHashes / bands

	detector := &MinHashSimilarityDetector{
		numHashes: numHashes,
		bands:     bands,
		rows:      rows,
		hashFuncs: make([]HashFunction, numHashes),
		lshIndex:  make(map[string][]string),
	}

	// Initialize hash functions with random parameters
	for i := 0; i < numHashes; i++ {
		detector.hashFuncs[i] = HashFunction{
			a: uint64(i*2 + 1),
			b: uint64(i*3 + 2),
			c: uint64(1<<32 - 1),
		}
	}

	return detector
}

// ComputeSignature computes MinHash signature for data
func (mh *MinHashSimilarityDetector) ComputeSignature(data []byte) ([]uint64, error) {
	// Extract shingles (k-grams) from data
	shingles := extractShingles(data, 3)

	signature := make([]uint64, mh.numHashes)

	for i := 0; i < mh.numHashes; i++ {
		signature[i] = ^uint64(0) // Initialize to max value

		for _, shingle := range shingles {
			hashVal := mh.hashFuncs[i].Hash(shingle)
			if hashVal < signature[i] {
				signature[i] = hashVal
			}
		}
	}

	return signature, nil
}

// FindSimilar finds similar items using LSH
func (mh *MinHashSimilarityDetector) FindSimilar(signature []uint64, threshold float64) ([]string, error) {
	similar := make(map[string]bool)

	// Check each LSH band
	for band := 0; band < mh.bands; band++ {
		bandSig := signature[band*mh.rows : (band+1)*mh.rows]
		bandKey := computeBandKey(bandSig)

		// Find candidates in this band
		if candidates, exists := mh.lshIndex[bandKey]; exists {
			for _, candidate := range candidates {
				similar[candidate] = true
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(similar))
	for item := range similar {
		result = append(result, item)
	}

	return result, nil
}

// AddToIndex adds an item to the similarity index
func (mh *MinHashSimilarityDetector) AddToIndex(id string, signature []uint64) error {
	// Add to each LSH band
	for band := 0; band < mh.bands; band++ {
		bandSig := signature[band*mh.rows : (band+1)*mh.rows]
		bandKey := computeBandKey(bandSig)

		if _, exists := mh.lshIndex[bandKey]; !exists {
			mh.lshIndex[bandKey] = make([]string, 0)
		}

		mh.lshIndex[bandKey] = append(mh.lshIndex[bandKey], id)
	}

	return nil
}

// RemoveFromIndex removes an item from the similarity index
func (mh *MinHashSimilarityDetector) RemoveFromIndex(id string) error {
	// Remove from all bands (would need to track which bands the item is in)
	for bandKey, candidates := range mh.lshIndex {
		filtered := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			if candidate != id {
				filtered = append(filtered, candidate)
			}
		}

		if len(filtered) == 0 {
			delete(mh.lshIndex, bandKey)
		} else {
			mh.lshIndex[bandKey] = filtered
		}
	}

	return nil
}

// Hash computes hash value for a shingle
func (hf *HashFunction) Hash(shingle uint64) uint64 {
	return ((hf.a*shingle + hf.b) % hf.c)
}

// extractShingles extracts k-shingles from data
func extractShingles(data []byte, k int) []uint64 {
	if len(data) < k {
		return []uint64{uint64(crc32.ChecksumIEEE(data))}
	}

	shingles := make([]uint64, 0, len(data)-k+1)

	for i := 0; i <= len(data)-k; i++ {
		shingle := crc32.ChecksumIEEE(data[i : i+k])
		shingles = append(shingles, uint64(shingle))
	}

	return shingles
}

// computeBandKey computes a key for a band signature
func computeBandKey(bandSig []uint64) string {
	hash := sha256.New()
	for _, val := range bandSig {
		hash.Write([]byte{
			byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
			byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
		})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// ChunkingAnalyzer provides analysis of chunking effectiveness
type ChunkingAnalyzer struct {
	chunkSizes []int
	boundaries []int
	duplicates map[string]int
}

// NewChunkingAnalyzer creates a new chunking analyzer
func NewChunkingAnalyzer() *ChunkingAnalyzer {
	return &ChunkingAnalyzer{
		chunkSizes: make([]int, 0),
		boundaries: make([]int, 0),
		duplicates: make(map[string]int),
	}
}

// AnalyzeChunking analyzes the effectiveness of chunking for given data
func (ca *ChunkingAnalyzer) AnalyzeChunking(chunks []Chunk) *ChunkingAnalysis {
	ca.chunkSizes = ca.chunkSizes[:0]
	ca.boundaries = ca.boundaries[:0]
	ca.duplicates = make(map[string]int)

	totalSize := 0
	uniqueChunks := 0

	for _, chunk := range chunks {
		ca.chunkSizes = append(ca.chunkSizes, chunk.Size)
		totalSize += chunk.Size

		hashStr := hex.EncodeToString(chunk.Hash)
		ca.duplicates[hashStr]++

		if ca.duplicates[hashStr] == 1 {
			uniqueChunks++
		}
	}

	return &ChunkingAnalysis{
		TotalChunks:    len(chunks),
		UniqueChunks:   uniqueChunks,
		DuplicateChunks: len(chunks) - uniqueChunks,
		TotalSize:      totalSize,
		AvgChunkSize:   float64(totalSize) / float64(len(chunks)),
		MinChunkSize:   minInt(ca.chunkSizes),
		MaxChunkSize:   maxInt(ca.chunkSizes),
		DeduplicationRatio: float64(uniqueChunks) / float64(len(chunks)),
	}
}

// ChunkingAnalysis represents the results of chunking analysis
type ChunkingAnalysis struct {
	TotalChunks        int     `json:"total_chunks"`
	UniqueChunks       int     `json:"unique_chunks"`
	DuplicateChunks    int     `json:"duplicate_chunks"`
	TotalSize          int     `json:"total_size"`
	AvgChunkSize       float64 `json:"avg_chunk_size"`
	MinChunkSize       int     `json:"min_chunk_size"`
	MaxChunkSize       int     `json:"max_chunk_size"`
	DeduplicationRatio float64 `json:"deduplication_ratio"`
}

// Helper functions
func minInt(slice []int) int {
	if len(slice) == 0 {
		return 0
	}

	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxInt(slice []int) int {
	if len(slice) == 0 {
		return 0
	}

	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max
}