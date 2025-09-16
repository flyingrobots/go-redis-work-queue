// Copyright 2025 James Ross
package themeplayground

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// PlaygroundHandler provides HTTP endpoints for theme operations
type PlaygroundHandler struct {
	themeManager *ThemeManager
}

// NewPlaygroundHandler creates a new playground handler
func NewPlaygroundHandler(themeManager *ThemeManager) *PlaygroundHandler {
	return &PlaygroundHandler{
		themeManager: themeManager,
	}
}

// GetThemes returns all available themes
func (h *PlaygroundHandler) GetThemes(w http.ResponseWriter, r *http.Request) {
	themes := h.themeManager.ListThemes()

	response := map[string]interface{}{
		"themes": themes,
		"active": h.themeManager.GetActiveTheme().Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTheme returns a specific theme by name
func (h *PlaygroundHandler) GetTheme(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/themes/")
	if name == "" {
		http.Error(w, "Theme name required", http.StatusBadRequest)
		return
	}

	theme, err := h.themeManager.GetTheme(name)
	if err != nil {
		if err.(*ThemeError).Code == "THEME_NOT_FOUND" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(theme)
}

// SetActiveTheme sets the currently active theme
func (h *PlaygroundHandler) SetActiveTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Theme string `json:"theme"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Theme == "" {
		http.Error(w, "Theme name required", http.StatusBadRequest)
		return
	}

	if err := h.themeManager.SetActiveTheme(request.Theme); err != nil {
		if err.(*ThemeError).Code == "THEME_NOT_FOUND" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"success": true,
		"active":  request.Theme,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PreviewTheme returns theme preview information
func (h *PlaygroundHandler) PreviewTheme(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/themes/")
	name = strings.TrimSuffix(name, "/preview")

	if name == "" {
		http.Error(w, "Theme name required", http.StatusBadRequest)
		return
	}

	theme, err := h.themeManager.GetTheme(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Generate preview data
	preview := map[string]interface{}{
		"name":        theme.Name,
		"description": theme.Description,
		"category":    theme.Category,
		"palette": map[string]string{
			"background":   theme.Palette.Background.Hex,
			"primary":      theme.Palette.Primary.Hex,
			"secondary":    theme.Palette.Secondary.Hex,
			"accent":       theme.Palette.Accent.Hex,
			"text_primary": theme.Palette.TextPrimary.Hex,
			"success":      theme.Palette.Success.Hex,
			"warning":      theme.Palette.Warning.Hex,
			"error":        theme.Palette.Error.Hex,
		},
		"accessibility": theme.Accessibility,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// ValidateTheme validates a theme configuration
func (h *PlaygroundHandler) ValidateTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var theme Theme
	if err := json.NewDecoder(r.Body).Decode(&theme); err != nil {
		http.Error(w, "Invalid theme data", http.StatusBadRequest)
		return
	}

	if err := h.themeManager.ValidateTheme(&theme); err != nil {
		response := map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"valid":         true,
		"accessibility": theme.Accessibility,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SaveCustomTheme saves a custom theme
func (h *PlaygroundHandler) SaveCustomTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var theme Theme
	if err := json.NewDecoder(r.Body).Decode(&theme); err != nil {
		http.Error(w, "Invalid theme data", http.StatusBadRequest)
		return
	}

	// Set category to custom for user-created themes
	theme.Category = CategoryCustom

	if err := h.themeManager.SaveTheme(&theme); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"theme":   theme.Name,
		"saved":   true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPreferences returns user theme preferences
func (h *PlaygroundHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	prefs := h.themeManager.preferences
	if prefs == nil {
		http.Error(w, "Preferences not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// UpdatePreferences updates user theme preferences
func (h *PlaygroundHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	prefs := h.themeManager.preferences
	if prefs == nil {
		http.Error(w, "Preferences not available", http.StatusInternalServerError)
		return
	}

	// Update preferences based on provided fields
	for key, value := range updates {
		switch key {
		case "auto_detect_terminal":
			if v, ok := value.(bool); ok {
				prefs.AutoDetectTerminal = v
			}
		case "respect_no_color":
			if v, ok := value.(bool); ok {
				prefs.RespectNoColor = v
			}
		case "sync_with_system":
			if v, ok := value.(bool); ok {
				prefs.SyncWithSystem = v
			}
		case "accessibility_mode":
			if v, ok := value.(bool); ok {
				prefs.AccessibilityMode = v
			}
		case "motion_reduced":
			if v, ok := value.(bool); ok {
				prefs.MotionReduced = v
			}
		}
	}

	h.themeManager.savePreferences()

	response := map[string]interface{}{
		"success":     true,
		"preferences": prefs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportTheme exports a theme as JSON
func (h *PlaygroundHandler) ExportTheme(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/themes/")
	name = strings.TrimSuffix(name, "/export")

	if name == "" {
		http.Error(w, "Theme name required", http.StatusBadRequest)
		return
	}

	theme, err := h.themeManager.GetTheme(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("%s-theme.json", strings.ReplaceAll(theme.Name, " ", "-"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(theme)
}

// ImportTheme imports a theme from uploaded JSON
func (h *PlaygroundHandler) ImportTheme(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("theme")
	if err != nil {
		http.Error(w, "No theme file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if filepath.Ext(header.Filename) != ".json" {
		http.Error(w, "Invalid file type, expected .json", http.StatusBadRequest)
		return
	}

	// Decode theme
	var theme Theme
	if err := json.NewDecoder(file).Decode(&theme); err != nil {
		http.Error(w, "Invalid theme file", http.StatusBadRequest)
		return
	}

	// Validate and save theme
	if err := h.themeManager.SaveTheme(&theme); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"theme":   theme.Name,
		"message": "Theme imported successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConfigPath returns the current configuration directory path
func (h *PlaygroundHandler) GetConfigPath(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"config_dir":        h.themeManager.configDir,
		"themes_dir":        filepath.Join(h.themeManager.configDir, "themes"),
		"preferences_file":  filepath.Join(h.themeManager.configDir, "theme_preferences.json"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}