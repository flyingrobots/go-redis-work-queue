// Copyright 2025 James Ross
package adminapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLogger handles audit log writing with rotation
type AuditLogger struct {
	mu          sync.Mutex
	file        *os.File
	path        string
	maxSize     int64
	maxBackups  int
	currentSize int64
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(path string, maxSize int64, maxBackups int) (*AuditLogger, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat audit log file: %w", err)
	}

	return &AuditLogger{
		file:        file,
		path:        path,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		currentSize: stat.Size(),
	}, nil
}

// Log writes an audit entry
func (l *AuditLogger) Log(entry AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	data = append(data, '\n')

	// Check if rotation is needed
	if l.currentSize+int64(len(data)) > l.maxSize {
		if err := l.rotate(); err != nil {
			return fmt.Errorf("failed to rotate audit log: %w", err)
		}
	}

	n, err := l.file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	l.currentSize += int64(n)
	return nil
}

// rotate performs log rotation
func (l *AuditLogger) rotate() error {
	l.file.Close()

	// Rename current file
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", l.path, timestamp)
	if err := os.Rename(l.path, newPath); err != nil {
		return err
	}

	// Clean up old backups
	if err := l.cleanupBackups(); err != nil {
		// Log but don't fail rotation
		fmt.Fprintf(os.Stderr, "Failed to cleanup old audit logs: %v\n", err)
	}

	// Open new file
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.currentSize = 0
	return nil
}

// cleanupBackups removes old backup files
func (l *AuditLogger) cleanupBackups() error {
	pattern := l.path + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(matches) <= l.maxBackups {
		return nil
	}

	// Sort by modification time and remove oldest
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			path:    match,
			modTime: stat.ModTime(),
		})
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files
	toRemove := len(files) - l.maxBackups
	for i := 0; i < toRemove && i < len(files); i++ {
		os.Remove(files[i].path)
	}

	return nil
}

// Close closes the audit logger
func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}