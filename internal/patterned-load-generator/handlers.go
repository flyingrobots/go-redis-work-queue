// Copyright 2025 James Ross
package patternedloadgenerator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandlers provides HTTP handlers for the patterned load generator
type HTTPHandlers struct {
	generator *LoadGenerator
	logger    *zap.Logger
}

// NewHTTPHandlers creates new HTTP handlers
func NewHTTPHandlers(generator *LoadGenerator, logger *zap.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		generator: generator,
		logger:    logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandlers) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/patterned-load-generator").Subrouter()

	// Profile management
	api.HandleFunc("/profiles", h.handleListProfiles).Methods("GET")
	api.HandleFunc("/profiles", h.handleCreateProfile).Methods("POST")
	api.HandleFunc("/profiles/{id}", h.handleGetProfile).Methods("GET")
	api.HandleFunc("/profiles/{id}", h.handleUpdateProfile).Methods("PUT")
	api.HandleFunc("/profiles/{id}", h.handleDeleteProfile).Methods("DELETE")

	// Load generation control
	api.HandleFunc("/start/{profileId}", h.handleStartProfile).Methods("POST")
	api.HandleFunc("/start", h.handleStartPattern).Methods("POST")
	api.HandleFunc("/stop", h.handleStop).Methods("POST")
	api.HandleFunc("/pause", h.handlePause).Methods("POST")
	api.HandleFunc("/resume", h.handleResume).Methods("POST")

	// Status and monitoring
	api.HandleFunc("/status", h.handleGetStatus).Methods("GET")
	api.HandleFunc("/metrics", h.handleGetMetrics).Methods("GET")

	// Pattern preview
	api.HandleFunc("/preview", h.handlePreviewPattern).Methods("POST")
}

// Profile management handlers

func (h *HTTPHandlers) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.generator.ListProfiles()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list profiles", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"profiles": profiles,
		"count":    len(profiles),
	})
}

func (h *HTTPHandlers) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var profile LoadProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.generator.SaveProfile(&profile); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to save profile", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, profile)
}

func (h *HTTPHandlers) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	profileID := mux.Vars(r)["id"]
	profile, err := h.generator.LoadProfile(profileID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Profile not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, profile)
}

func (h *HTTPHandlers) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	profileID := mux.Vars(r)["id"]
	var profile LoadProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	profile.ID = profileID
	if err := h.generator.SaveProfile(&profile); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	h.writeJSON(w, http.StatusOK, profile)
}

func (h *HTTPHandlers) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	profileID := mux.Vars(r)["id"]
	if err := h.generator.DeleteProfile(profileID); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete profile", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Load generation control handlers

func (h *HTTPHandlers) handleStartProfile(w http.ResponseWriter, r *http.Request) {
	profileID := mux.Vars(r)["profileId"]
	if err := h.generator.Start(profileID); err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to start load generation", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":    "started",
		"profile_id": profileID,
	})
}

func (h *HTTPHandlers) handleStartPattern(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pattern    LoadPattern `json:"pattern"`
		Guardrails *Guardrails `json:"guardrails,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.generator.StartPattern(&req.Pattern, req.Guardrails); err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to start pattern", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "started",
		"pattern": req.Pattern.Type,
	})
}

func (h *HTTPHandlers) handleStop(w http.ResponseWriter, r *http.Request) {
	if err := h.generator.Stop(); err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to stop", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *HTTPHandlers) handlePause(w http.ResponseWriter, r *http.Request) {
	if err := h.generator.Pause(); err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to pause", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (h *HTTPHandlers) handleResume(w http.ResponseWriter, r *http.Request) {
	if err := h.generator.Resume(); err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to resume", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}

// Status and monitoring handlers

func (h *HTTPHandlers) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.generator.GetStatus()
	h.writeJSON(w, http.StatusOK, status)
}

func (h *HTTPHandlers) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for duration
	duration := 5 * time.Minute // default
	if d := r.URL.Query().Get("duration"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			duration = parsed
		}
	}

	metrics := h.generator.GetMetrics(duration)
	h.writeJSON(w, http.StatusOK, metrics)
}

// Pattern preview handler

func (h *HTTPHandlers) handlePreviewPattern(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pattern    LoadPattern   `json:"pattern"`
		Duration   time.Duration `json:"duration"`
		Resolution time.Duration `json:"resolution,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if req.Resolution == 0 {
		req.Resolution = 1 * time.Second
	}

	if req.Duration == 0 {
		req.Duration = req.Pattern.Duration
	}

	preview, err := h.generator.PreviewPattern(req.Pattern, req.Duration, req.Resolution)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to generate preview", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"pattern":    req.Pattern.Type,
		"duration":   req.Duration,
		"resolution": req.Resolution,
		"points":     preview,
	})
}

// Utility methods

func (h *HTTPHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *HTTPHandlers) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.Error(err))
	h.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	})
}

// PatternRequest represents a request to start a pattern
type PatternRequest struct {
	Type       PatternType            `json:"type"`
	Duration   time.Duration          `json:"duration"`
	Parameters map[string]interface{} `json:"parameters"`
	Guardrails *Guardrails            `json:"guardrails,omitempty"`
	QueueName  string                 `json:"queue_name,omitempty"`
}

// ToLoadPattern converts a PatternRequest to a LoadPattern
func (pr *PatternRequest) ToLoadPattern() LoadPattern {
	return LoadPattern{
		Type:        pr.Type,
		Name:        fmt.Sprintf("API Pattern: %s", pr.Type),
		Description: "Pattern started via API",
		Duration:    pr.Duration,
		Parameters:  pr.Parameters,
	}
}