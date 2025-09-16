// Copyright 2025 James Ross
package rbacandtokens

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// AuditLogger handles audit log entries
type AuditLogger struct {
	writer   io.Writer
	file     *lumberjack.Logger
	mutex    sync.Mutex
	config   *AuditConfig
	filterFn func(*AuditEntry) *AuditEntry
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config *AuditConfig) (*AuditLogger, error) {
	if config == nil {
		return nil, fmt.Errorf("audit config is required")
	}

	if !config.Enabled {
		return &AuditLogger{
			config:   config,
			filterFn: defaultFilter,
		}, nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(config.LogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Configure rotating file writer
	fileWriter := &lumberjack.Logger{
		Filename:   config.LogPath,
		MaxSize:    int(config.RotateSize / (1024 * 1024)), // Convert to MB
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}

	logger := &AuditLogger{
		writer:   fileWriter,
		file:     fileWriter,
		config:   config,
		filterFn: defaultFilter,
	}

	if config.FilterSensitive {
		logger.filterFn = sensitiveFilter
	}

	return logger, nil
}

// Log writes an audit entry
func (a *AuditLogger) Log(entry AuditEntry) error {
	if !a.config.Enabled {
		return nil
	}

	// Apply filtering
	filteredEntry := a.filterFn(&entry)
	if filteredEntry == nil {
		return nil // Entry filtered out
	}

	// Ensure timestamp is set
	if filteredEntry.Timestamp.IsZero() {
		filteredEntry.Timestamp = time.Now()
	}

	// Serialize to JSON
	entryBytes, err := json.Marshal(filteredEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Write with newline
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.writer != nil {
		_, err = a.writer.Write(append(entryBytes, '\n'))
		if err != nil {
			return fmt.Errorf("failed to write audit entry: %w", err)
		}
	}

	return nil
}

// Query searches audit entries with filters
func (a *AuditLogger) Query(filter AuditFilter) ([]*AuditEntry, error) {
	if !a.config.Enabled {
		return []*AuditEntry{}, nil
	}

	// In a production system, this would query a database or search index
	// For now, we'll read from the log file
	file, err := os.Open(a.config.LogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*AuditEntry{}, nil
		}
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	var entries []*AuditEntry
	decoder := json.NewDecoder(file)

	for {
		var entry AuditEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			// Skip malformed entries
			continue
		}

		// Apply filters
		if matchesFilter(&entry, &filter) {
			entries = append(entries, &entry)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Apply limit
	if filter.Limit > 0 && len(entries) > filter.Limit {
		entries = entries[:filter.Limit]
	}

	return entries, nil
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	if a.file != nil {
		return a.file.Close()
	}
	return nil
}

// AuditFilter represents filters for querying audit entries
type AuditFilter struct {
	Subject   string    `json:"subject,omitempty"`
	Action    string    `json:"action,omitempty"`
	Resource  string    `json:"resource,omitempty"`
	Result    string    `json:"result,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	IP        string    `json:"ip,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// Helper functions

func defaultFilter(entry *AuditEntry) *AuditEntry {
	return entry // No filtering
}

func sensitiveFilter(entry *AuditEntry) *AuditEntry {
	if entry == nil {
		return nil
	}

	// Create a copy
	filtered := *entry

	// Remove sensitive details
	if filtered.Details != nil {
		filteredDetails := make(map[string]interface{})
		for k, v := range filtered.Details {
			// Filter out potentially sensitive keys
			switch k {
			case "token", "password", "secret", "key", "private_key":
				filteredDetails[k] = "[REDACTED]"
			default:
				filteredDetails[k] = v
			}
		}
		filtered.Details = filteredDetails
	}

	// Redact sensitive parts of user agent
	if len(filtered.UserAgent) > 100 {
		filtered.UserAgent = filtered.UserAgent[:100] + "..."
	}

	return &filtered
}

func matchesFilter(entry *AuditEntry, filter *AuditFilter) bool {
	if filter.Subject != "" && entry.Subject != filter.Subject {
		return false
	}

	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}

	if filter.Resource != "" && entry.Resource != filter.Resource {
		return false
	}

	if filter.Result != "" && entry.Result != filter.Result {
		return false
	}

	if filter.IP != "" && entry.IP != filter.IP {
		return false
	}

	if filter.RequestID != "" && entry.RequestID != filter.RequestID {
		return false
	}

	if !filter.StartTime.IsZero() && entry.Timestamp.Before(filter.StartTime) {
		return false
	}

	if !filter.EndTime.IsZero() && entry.Timestamp.After(filter.EndTime) {
		return false
	}

	return true
}