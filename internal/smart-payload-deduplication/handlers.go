// Copyright 2025 James Ross
package deduplication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// DeduplicationService provides HTTP handlers for deduplication operations
type DeduplicationService struct {
	manager DeduplicationManager
	logger  *zap.Logger
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService(manager DeduplicationManager, logger *zap.Logger) *DeduplicationService {
	return &DeduplicationService{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers HTTP routes for the deduplication service
func (ds *DeduplicationService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/dedup/payload", ds.handlePayloadDeduplication)
	mux.HandleFunc("/api/v1/dedup/reconstruct", ds.handlePayloadReconstruction)
	mux.HandleFunc("/api/v1/dedup/stats", ds.handleStats)
	mux.HandleFunc("/api/v1/dedup/chunks", ds.handleChunks)
	mux.HandleFunc("/api/v1/dedup/gc", ds.handleGarbageCollection)
	mux.HandleFunc("/api/v1/dedup/health", ds.handleHealth)
	mux.HandleFunc("/api/v1/dedup/config", ds.handleConfig)
}

// Payload deduplication endpoint
func (ds *DeduplicationService) handlePayloadDeduplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request PayloadDeduplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ds.logger.Error("Failed to decode deduplication request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.JobID == "" {
		http.Error(w, "job_id is required", http.StatusBadRequest)
		return
	}

	if len(request.Payload) == 0 {
		http.Error(w, "payload is required", http.StatusBadRequest)
		return
	}

	start := time.Now()

	payloadMap, err := ds.manager.DeduplicatePayload(request.JobID, request.Payload)
	if err != nil {
		ds.logger.Error("Payload deduplication failed",
			zap.String("job_id", request.JobID),
			zap.Error(err))

		if IsRecoverable(err) {
			http.Error(w, "Deduplication temporarily unavailable", http.StatusServiceUnavailable)
		} else {
			http.Error(w, "Deduplication failed", http.StatusInternalServerError)
		}
		return
	}

	response := PayloadDeduplicationResponse{
		JobID:        request.JobID,
		PayloadMap:   payloadMap,
		OriginalSize: len(request.Payload),
		ChunkCount:   len(payloadMap.ChunkRefs),
		ProcessingTime: time.Since(start),
		Compressed:   payloadMap.Compressed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ds.logger.Debug("Payload deduplication completed",
		zap.String("job_id", request.JobID),
		zap.Int("original_size", len(request.Payload)),
		zap.Int("chunk_count", len(payloadMap.ChunkRefs)),
		zap.Duration("processing_time", time.Since(start)))
}

// Payload reconstruction endpoint
func (ds *DeduplicationService) handlePayloadReconstruction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request PayloadReconstructionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ds.logger.Error("Failed to decode reconstruction request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.PayloadMap == nil {
		http.Error(w, "payload_map is required", http.StatusBadRequest)
		return
	}

	start := time.Now()

	payload, err := ds.manager.ReconstructPayload(request.PayloadMap)
	if err != nil {
		ds.logger.Error("Payload reconstruction failed",
			zap.String("job_id", request.PayloadMap.JobID),
			zap.Error(err))

		if IsRecoverable(err) {
			http.Error(w, "Reconstruction temporarily unavailable", http.StatusServiceUnavailable)
		} else {
			http.Error(w, "Reconstruction failed", http.StatusInternalServerError)
		}
		return
	}

	response := PayloadReconstructionResponse{
		JobID:           request.PayloadMap.JobID,
		Payload:         payload,
		ReconstructedSize: len(payload),
		ProcessingTime:  time.Since(start),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ds.logger.Debug("Payload reconstruction completed",
		zap.String("job_id", request.PayloadMap.JobID),
		zap.Int("reconstructed_size", len(payload)),
		zap.Duration("processing_time", time.Since(start)))
}

// Statistics endpoint
func (ds *DeduplicationService) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := ds.manager.GetStats()
	if err != nil {
		ds.logger.Error("Failed to get statistics", zap.Error(err))
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Chunks management endpoint
func (ds *DeduplicationService) handleChunks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ds.handleGetChunks(w, r)
	case http.MethodDelete:
		ds.handleDeleteChunk(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ds *DeduplicationService) handleGetChunks(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	popularChunks, err := ds.manager.GetPopularChunks(limit)
	if err != nil {
		ds.logger.Error("Failed to get popular chunks", zap.Error(err))
		http.Error(w, "Failed to get chunks", http.StatusInternalServerError)
		return
	}

	response := ChunksResponse{
		Chunks: popularChunks,
		Total:  len(popularChunks),
		Limit:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ds *DeduplicationService) handleDeleteChunk(w http.ResponseWriter, r *http.Request) {
	hashParam := r.URL.Query().Get("hash")
	if hashParam == "" {
		http.Error(w, "hash parameter is required", http.StatusBadRequest)
		return
	}

	// Decode hex hash
	hash, err := decodeHexHash(hashParam)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	err = ds.manager.DeleteChunk(hash)
	if err != nil {
		ds.logger.Error("Failed to delete chunk",
			zap.String("hash", hashParam),
			zap.Error(err))
		http.Error(w, "Failed to delete chunk", http.StatusInternalServerError)
		return
	}

	response := DeleteChunkResponse{
		Hash:    hashParam,
		Deleted: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Garbage collection endpoint
func (ds *DeduplicationService) handleGarbageCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()

	err := ds.manager.RunGarbageCollection()
	if err != nil {
		ds.logger.Error("Garbage collection failed", zap.Error(err))
		http.Error(w, "Garbage collection failed", http.StatusInternalServerError)
		return
	}

	response := GarbageCollectionResponse{
		Started:   start,
		Completed: time.Now(),
		Duration:  time.Since(start),
		Success:   true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ds.logger.Info("Manual garbage collection completed",
		zap.Duration("duration", time.Since(start)))
}

// Health check endpoint
func (ds *DeduplicationService) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := ds.manager.GetHealth()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Configuration endpoint
func (ds *DeduplicationService) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ds.handleGetConfig(w, r)
	case http.MethodPut:
		ds.handleUpdateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ds *DeduplicationService) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config := ds.manager.GetConfig()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (ds *DeduplicationService) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig Config
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		ds.logger.Error("Failed to decode config update", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := ds.manager.UpdateConfig(&newConfig)
	if err != nil {
		ds.logger.Error("Failed to update configuration", zap.Error(err))
		http.Error(w, fmt.Sprintf("Configuration update failed: %v", err), http.StatusBadRequest)
		return
	}

	response := ConfigUpdateResponse{
		Updated: true,
		Config:  &newConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ds.logger.Info("Configuration updated successfully")
}

// Request/Response types for HTTP API

type PayloadDeduplicationRequest struct {
	JobID   string `json:"job_id"`
	Payload []byte `json:"payload"`
}

type PayloadDeduplicationResponse struct {
	JobID          string           `json:"job_id"`
	PayloadMap     *PayloadMap      `json:"payload_map"`
	OriginalSize   int              `json:"original_size"`
	ChunkCount     int              `json:"chunk_count"`
	ProcessingTime time.Duration    `json:"processing_time"`
	Compressed     bool             `json:"compressed"`
}

type PayloadReconstructionRequest struct {
	PayloadMap *PayloadMap `json:"payload_map"`
}

type PayloadReconstructionResponse struct {
	JobID             string        `json:"job_id"`
	Payload           []byte        `json:"payload"`
	ReconstructedSize int           `json:"reconstructed_size"`
	ProcessingTime    time.Duration `json:"processing_time"`
}

type ChunksResponse struct {
	Chunks []ChunkStats `json:"chunks"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
}

type DeleteChunkResponse struct {
	Hash    string `json:"hash"`
	Deleted bool   `json:"deleted"`
}

type GarbageCollectionResponse struct {
	Started   time.Time     `json:"started"`
	Completed time.Time     `json:"completed"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
}

type ConfigUpdateResponse struct {
	Updated bool    `json:"updated"`
	Config  *Config `json:"config"`
}

// Producer/Worker Integration

// ProducerIntegration provides integration with the producer system
type ProducerIntegration struct {
	manager       DeduplicationManager
	logger        *zap.Logger
	enabled       bool
	fallbackMode  bool
	migrationRate float64
}

// NewProducerIntegration creates a new producer integration
func NewProducerIntegration(manager DeduplicationManager, config *Config, logger *zap.Logger) *ProducerIntegration {
	return &ProducerIntegration{
		manager:       manager,
		logger:        logger,
		enabled:       config.Enabled,
		fallbackMode:  config.SafetyMode,
		migrationRate: config.MigrationRatio,
	}
}

// EnqueueJob handles job enqueuing with optional deduplication
func (pi *ProducerIntegration) EnqueueJob(jobID string, payload []byte) (interface{}, error) {
	if !pi.enabled || !pi.shouldDeduplicate(jobID) {
		return payload, nil // Return original payload
	}

	// Attempt deduplication
	payloadMap, err := pi.manager.DeduplicatePayload(jobID, payload)
	if err != nil {
		pi.logger.Warn("Deduplication failed, falling back to original payload",
			zap.String("job_id", jobID),
			zap.Error(err))

		if pi.fallbackMode {
			return payload, nil // Fallback to original payload
		}

		return nil, err
	}

	pi.logger.Debug("Job payload deduplicated",
		zap.String("job_id", jobID),
		zap.Int("original_size", len(payload)),
		zap.Int("chunk_count", len(payloadMap.ChunkRefs)))

	return payloadMap, nil
}

// shouldDeduplicate determines if a job should be deduplicated based on migration rate
func (pi *ProducerIntegration) shouldDeduplicate(jobID string) bool {
	if pi.migrationRate >= 1.0 {
		return true
	}

	if pi.migrationRate <= 0.0 {
		return false
	}

	// Use job ID hash to deterministically decide
	hash := computeJobHash(jobID)
	threshold := uint64(float64(^uint64(0)) * pi.migrationRate)

	return hash < threshold
}

// WorkerIntegration provides integration with the worker system
type WorkerIntegration struct {
	manager      DeduplicationManager
	logger       *zap.Logger
	enabled      bool
	fallbackMode bool
}

// NewWorkerIntegration creates a new worker integration
func NewWorkerIntegration(manager DeduplicationManager, config *Config, logger *zap.Logger) *WorkerIntegration {
	return &WorkerIntegration{
		manager:      manager,
		logger:       logger,
		enabled:      config.Enabled,
		fallbackMode: config.SafetyMode,
	}
}

// DequeueJob handles job dequeuing with payload reconstruction
func (wi *WorkerIntegration) DequeueJob(jobData interface{}) ([]byte, error) {
	// Check if data is a payload map (deduplicated) or raw payload
	if payloadMap, ok := jobData.(*PayloadMap); ok {
		// Reconstruct from payload map
		payload, err := wi.manager.ReconstructPayload(payloadMap)
		if err != nil {
			wi.logger.Error("Failed to reconstruct payload",
				zap.String("job_id", payloadMap.JobID),
				zap.Error(err))

			if wi.fallbackMode {
				return nil, fmt.Errorf("payload reconstruction failed and no fallback available")
			}

			return nil, err
		}

		wi.logger.Debug("Payload reconstructed",
			zap.String("job_id", payloadMap.JobID),
			zap.Int("reconstructed_size", len(payload)))

		return payload, nil
	}

	// Raw payload - return as-is
	if payload, ok := jobData.([]byte); ok {
		return payload, nil
	}

	return nil, fmt.Errorf("invalid job data type: %T", jobData)
}

// CleanupJob handles cleanup after job completion
func (wi *WorkerIntegration) CleanupJob(jobData interface{}, success bool) error {
	// If job failed, we might want to keep the payload map for retry
	if !success {
		return nil
	}

	// For successful jobs, clean up the payload map
	if payloadMap, ok := jobData.(*PayloadMap); ok {
		// Decrement chunk reference counts
		for _, chunkRef := range payloadMap.ChunkRefs {
			if err := wi.manager.RemoveReference(chunkRef.Hash); err != nil {
				wi.logger.Warn("Failed to remove chunk reference",
					zap.String("job_id", payloadMap.JobID),
					zap.String("chunk_hash", fmt.Sprintf("%x", chunkRef.Hash)),
					zap.Error(err))
			}
		}

		wi.logger.Debug("Job cleanup completed",
			zap.String("job_id", payloadMap.JobID),
			zap.Int("chunks_dereferenced", len(payloadMap.ChunkRefs)))
	}

	return nil
}

// Utility functions

func decodeHexHash(hashStr string) ([]byte, error) {
	// Remove potential 0x prefix
	if len(hashStr) > 2 && hashStr[:2] == "0x" {
		hashStr = hashStr[2:]
	}

	// Decode hex string
	hash := make([]byte, len(hashStr)/2)
	for i := 0; i < len(hash); i++ {
		var b byte
		_, err := fmt.Sscanf(hashStr[i*2:i*2+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		hash[i] = b
	}

	return hash, nil
}

func computeJobHash(jobID string) uint64 {
	hash := uint64(0)
	for _, b := range []byte(jobID) {
		hash = hash*31 + uint64(b)
	}
	return hash
}