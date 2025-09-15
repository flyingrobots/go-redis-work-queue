// Copyright 2025 James Ross
package deduplication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

// ZstdCompressor implements compression using the Zstandard algorithm
type ZstdCompressor struct {
	encoder     *zstd.Encoder
	decoder     *zstd.Decoder
	dictionary  []byte
	level       zstd.EncoderLevel
	dictSize    int
	useDictionary bool
	mu          sync.RWMutex
	stats       CompressionStats
}

// CompressionStats tracks compression performance metrics
type CompressionStats struct {
	TotalCompressed   int64   `json:"total_compressed"`
	TotalDecompressed int64   `json:"total_decompressed"`
	BytesIn           int64   `json:"bytes_in"`
	BytesOut          int64   `json:"bytes_out"`
	CompressionRatio  float64 `json:"compression_ratio"`
	AvgCompressionTime time.Duration `json:"avg_compression_time"`
	AvgDecompressionTime time.Duration `json:"avg_decompression_time"`
	DictionarySize    int     `json:"dictionary_size"`
	LastUpdated       time.Time `json:"last_updated"`
}

// NewZstdCompressor creates a new Zstd-based compressor
func NewZstdCompressor(config *CompressionConfig) (*ZstdCompressor, error) {
	if !config.Enabled {
		return &ZstdCompressor{
			useDictionary: false,
			level:         zstd.SpeedDefault,
		}, nil
	}

	// Map compression level
	var level zstd.EncoderLevel
	switch {
	case config.Level <= 3:
		level = zstd.SpeedFastest
	case config.Level <= 6:
		level = zstd.SpeedDefault
	case config.Level <= 9:
		level = zstd.SpeedBetterCompression
	default:
		level = zstd.SpeedBestCompression
	}

	compressor := &ZstdCompressor{
		level:         level,
		dictSize:      config.DictionarySize,
		useDictionary: config.UseDictionary,
		stats: CompressionStats{
			LastUpdated: time.Now(),
		},
	}

	// Create encoder and decoder
	if err := compressor.initializeCodecs(); err != nil {
		return nil, err
	}

	return compressor, nil
}

// initializeCodecs creates the encoder and decoder
func (zc *ZstdCompressor) initializeCodecs() error {
	var err error

	// Create encoder options
	encoderOpts := []zstd.EOption{
		zstd.WithEncoderLevel(zc.level),
		zstd.WithEncoderConcurrency(1), // Single-threaded for consistency
	}

	// Create decoder options
	decoderOpts := []zstd.DOption{
		zstd.WithDecoderConcurrency(1), // Single-threaded for consistency
	}

	// Add dictionary if available
	if zc.useDictionary && len(zc.dictionary) > 0 {
		encoderOpts = append(encoderOpts, zstd.WithEncoderDict(zc.dictionary))
		decoderOpts = append(decoderOpts, zstd.WithDecoderDicts(zc.dictionary))
	}

	// Create encoder
	zc.encoder, err = zstd.NewWriter(nil, encoderOpts...)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	// Create decoder
	zc.decoder, err = zstd.NewReader(nil, decoderOpts...)
	if err != nil {
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}

	return nil
}

// Compress compresses data using Zstd
func (zc *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	start := time.Now()

	zc.mu.RLock()
	encoder := zc.encoder
	zc.mu.RUnlock()

	if encoder == nil {
		// Fallback: return uncompressed data if encoder not available
		return data, nil
	}

	// Compress the data
	compressed := encoder.EncodeAll(data, nil)

	// Update statistics
	duration := time.Since(start)
	zc.updateCompressionStats(len(data), len(compressed), duration)

	return compressed, nil
}

// Decompress decompresses data using Zstd
func (zc *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	start := time.Now()

	zc.mu.RLock()
	decoder := zc.decoder
	zc.mu.RUnlock()

	if decoder == nil {
		// Fallback: return data as-is if decoder not available
		return data, nil
	}

	// Decompress the data
	decompressed, err := decoder.DecodeAll(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Update statistics
	duration := time.Since(start)
	zc.updateDecompressionStats(len(data), len(decompressed), duration)

	return decompressed, nil
}

// BuildDictionary builds a compression dictionary from sample data
func (zc *ZstdCompressor) BuildDictionary(samples [][]byte) error {
	if !zc.useDictionary || len(samples) == 0 {
		return nil
	}

	// Combine all samples into a single buffer
	var totalSize int
	for _, sample := range samples {
		totalSize += len(sample)
	}

	combinedSamples := make([]byte, 0, totalSize)
	for _, sample := range samples {
		combinedSamples = append(combinedSamples, sample...)
	}

	// TODO: Build dictionary (API not available in current zstd version)
	// For now, just use the combined samples as a simple dictionary
	dictionary := combinedSamples
	if len(dictionary) > zc.dictSize {
		dictionary = dictionary[:zc.dictSize]
	}

	zc.mu.Lock()
	zc.dictionary = dictionary
	zc.stats.DictionarySize = len(dictionary)
	zc.mu.Unlock()

	// Reinitialize codecs with new dictionary
	return zc.initializeCodecs()
}

// GetCompressionRatio returns the current compression ratio
func (zc *ZstdCompressor) GetCompressionRatio() float64 {
	zc.mu.RLock()
	defer zc.mu.RUnlock()
	return zc.stats.CompressionRatio
}

// GetStats returns current compression statistics
func (zc *ZstdCompressor) GetStats() CompressionStats {
	zc.mu.RLock()
	defer zc.mu.RUnlock()
	return zc.stats
}

// updateCompressionStats updates compression statistics
func (zc *ZstdCompressor) updateCompressionStats(inputSize, outputSize int, duration time.Duration) {
	zc.mu.Lock()
	defer zc.mu.Unlock()

	zc.stats.TotalCompressed++
	zc.stats.BytesIn += int64(inputSize)
	zc.stats.BytesOut += int64(outputSize)

	// Update compression ratio
	if zc.stats.BytesIn > 0 {
		zc.stats.CompressionRatio = float64(zc.stats.BytesOut) / float64(zc.stats.BytesIn)
	}

	// Update average compression time (exponential moving average)
	if zc.stats.TotalCompressed == 1 {
		zc.stats.AvgCompressionTime = duration
	} else {
		alpha := 0.1 // Smoothing factor
		zc.stats.AvgCompressionTime = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(zc.stats.AvgCompressionTime),
		)
	}

	zc.stats.LastUpdated = time.Now()
}

// updateDecompressionStats updates decompression statistics
func (zc *ZstdCompressor) updateDecompressionStats(inputSize, outputSize int, duration time.Duration) {
	zc.mu.Lock()
	defer zc.mu.Unlock()

	zc.stats.TotalDecompressed++

	// Update average decompression time (exponential moving average)
	if zc.stats.TotalDecompressed == 1 {
		zc.stats.AvgDecompressionTime = duration
	} else {
		alpha := 0.1 // Smoothing factor
		zc.stats.AvgDecompressionTime = time.Duration(
			alpha*float64(duration) + (1-alpha)*float64(zc.stats.AvgDecompressionTime),
		)
	}

	zc.stats.LastUpdated = time.Now()
}

// Close closes the compressor and releases resources
func (zc *ZstdCompressor) Close() error {
	zc.mu.Lock()
	defer zc.mu.Unlock()

	if zc.encoder != nil {
		zc.encoder.Close()
		zc.encoder = nil
	}

	if zc.decoder != nil {
		zc.decoder.Close()
		zc.decoder = nil
	}

	return nil
}

// DictionaryBuilder helps build compression dictionaries from sample data
type DictionaryBuilder struct {
	samples    [][]byte
	maxSamples int
	maxSize    int
	totalSize  int
	mu         sync.Mutex
}

// NewDictionaryBuilder creates a new dictionary builder
func NewDictionaryBuilder(maxSamples, maxSize int) *DictionaryBuilder {
	return &DictionaryBuilder{
		samples:    make([][]byte, 0, maxSamples),
		maxSamples: maxSamples,
		maxSize:    maxSize,
	}
}

// AddSample adds a sample to the dictionary builder
func (db *DictionaryBuilder) AddSample(data []byte) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if we need to make room
	if len(db.samples) >= db.maxSamples {
		// Remove oldest sample (FIFO)
		removed := db.samples[0]
		db.samples = db.samples[1:]
		db.totalSize -= len(removed)
	}

	// Add new sample
	sample := make([]byte, len(data))
	copy(sample, data)
	db.samples = append(db.samples, sample)
	db.totalSize += len(sample)

	// Trim samples if total size exceeds limit
	for db.totalSize > db.maxSize && len(db.samples) > 0 {
		removed := db.samples[0]
		db.samples = db.samples[1:]
		db.totalSize -= len(removed)
	}
}

// BuildDictionary creates a compression dictionary from collected samples
func (db *DictionaryBuilder) BuildDictionary(dictSize int) ([]byte, error) {
	db.mu.Lock()
	samples := make([][]byte, len(db.samples))
	copy(samples, db.samples)
	db.mu.Unlock()

	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples available for dictionary building")
	}

	// Combine all samples
	var totalSize int
	for _, sample := range samples {
		totalSize += len(sample)
	}

	combinedData := make([]byte, 0, totalSize)
	for _, sample := range samples {
		combinedData = append(combinedData, sample...)
	}

	// TODO: Build dictionary (API not available in current zstd version)
	// For now, just use the combined data as a simple dictionary
	dictionary := combinedData
	if len(dictionary) > dictSize {
		dictionary = dictionary[:dictSize]
	}

	return dictionary, nil
}

// GetSampleCount returns the number of collected samples
func (db *DictionaryBuilder) GetSampleCount() int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return len(db.samples)
}

// GetTotalSize returns the total size of collected samples
func (db *DictionaryBuilder) GetTotalSize() int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.totalSize
}

// Clear removes all collected samples
func (db *DictionaryBuilder) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.samples = db.samples[:0]
	db.totalSize = 0
}

// CompressionBenchmark provides benchmarking tools for compression
type CompressionBenchmark struct {
	compressor Compressor
	samples    [][]byte
	results    []BenchmarkResult
}

// BenchmarkResult represents the result of a compression benchmark
type BenchmarkResult struct {
	SampleSize       int           `json:"sample_size"`
	OriginalSize     int           `json:"original_size"`
	CompressedSize   int           `json:"compressed_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	CompressionTime  time.Duration `json:"compression_time"`
	DecompressionTime time.Duration `json:"decompression_time"`
	Throughput       float64       `json:"throughput_mbps"`
}

// NewCompressionBenchmark creates a new compression benchmark
func NewCompressionBenchmark(compressor Compressor) *CompressionBenchmark {
	return &CompressionBenchmark{
		compressor: compressor,
		samples:    make([][]byte, 0),
		results:    make([]BenchmarkResult, 0),
	}
}

// AddSample adds a sample for benchmarking
func (cb *CompressionBenchmark) AddSample(data []byte) {
	sample := make([]byte, len(data))
	copy(sample, data)
	cb.samples = append(cb.samples, sample)
}

// RunBenchmark runs compression benchmarks on all samples
func (cb *CompressionBenchmark) RunBenchmark() error {
	cb.results = cb.results[:0]

	for i, sample := range cb.samples {
		// Benchmark compression
		start := time.Now()
		compressed, err := cb.compressor.Compress(sample)
		compressionTime := time.Since(start)

		if err != nil {
			return fmt.Errorf("compression failed for sample %d: %w", i, err)
		}

		// Benchmark decompression
		start = time.Now()
		decompressed, err := cb.compressor.Decompress(compressed)
		decompressionTime := time.Since(start)

		if err != nil {
			return fmt.Errorf("decompression failed for sample %d: %w", i, err)
		}

		// Verify correctness
		if !bytes.Equal(sample, decompressed) {
			return fmt.Errorf("decompression mismatch for sample %d", i)
		}

		// Calculate metrics
		compressionRatio := float64(len(compressed)) / float64(len(sample))
		throughput := float64(len(sample)) / (1024 * 1024) / compressionTime.Seconds()

		result := BenchmarkResult{
			SampleSize:        i,
			OriginalSize:      len(sample),
			CompressedSize:    len(compressed),
			CompressionRatio:  compressionRatio,
			CompressionTime:   compressionTime,
			DecompressionTime: decompressionTime,
			Throughput:        throughput,
		}

		cb.results = append(cb.results, result)
	}

	return nil
}

// GetResults returns benchmark results
func (cb *CompressionBenchmark) GetResults() []BenchmarkResult {
	return cb.results
}

// GetSummary returns a summary of benchmark results
func (cb *CompressionBenchmark) GetSummary() BenchmarkSummary {
	if len(cb.results) == 0 {
		return BenchmarkSummary{}
	}

	var totalOriginal, totalCompressed int64
	var totalCompressionTime, totalDecompressionTime time.Duration
	var minRatio, maxRatio float64 = 1.0, 0.0

	for _, result := range cb.results {
		totalOriginal += int64(result.OriginalSize)
		totalCompressed += int64(result.CompressedSize)
		totalCompressionTime += result.CompressionTime
		totalDecompressionTime += result.DecompressionTime

		if result.CompressionRatio < minRatio {
			minRatio = result.CompressionRatio
		}
		if result.CompressionRatio > maxRatio {
			maxRatio = result.CompressionRatio
		}
	}

	avgRatio := float64(totalCompressed) / float64(totalOriginal)
	avgCompressionTime := totalCompressionTime / time.Duration(len(cb.results))
	avgDecompressionTime := totalDecompressionTime / time.Duration(len(cb.results))

	return BenchmarkSummary{
		SampleCount:           len(cb.results),
		TotalOriginalBytes:    totalOriginal,
		TotalCompressedBytes:  totalCompressed,
		OverallCompressionRatio: avgRatio,
		MinCompressionRatio:   minRatio,
		MaxCompressionRatio:   maxRatio,
		AvgCompressionTime:    avgCompressionTime,
		AvgDecompressionTime:  avgDecompressionTime,
		SpaceSavings:          1.0 - avgRatio,
	}
}

// BenchmarkSummary provides a summary of benchmark results
type BenchmarkSummary struct {
	SampleCount             int           `json:"sample_count"`
	TotalOriginalBytes      int64         `json:"total_original_bytes"`
	TotalCompressedBytes    int64         `json:"total_compressed_bytes"`
	OverallCompressionRatio float64       `json:"overall_compression_ratio"`
	MinCompressionRatio     float64       `json:"min_compression_ratio"`
	MaxCompressionRatio     float64       `json:"max_compression_ratio"`
	AvgCompressionTime      time.Duration `json:"avg_compression_time"`
	AvgDecompressionTime    time.Duration `json:"avg_decompression_time"`
	SpaceSavings            float64       `json:"space_savings"`
}

// ToJSON converts benchmark summary to JSON
func (bs *BenchmarkSummary) ToJSON() ([]byte, error) {
	return json.MarshalIndent(bs, "", "  ")
}