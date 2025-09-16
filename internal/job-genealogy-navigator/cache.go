// Copyright 2025 James Ross
package genealogy

import (
	"time"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// GetTree retrieves a cached genealogy tree
func (gc *GenealogyCache) GetTree(jobID string) *JobGenealogy {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	entry, exists := gc.trees[jobID]
	if !exists || time.Now().After(entry.UpdatedAt.Add(gc.ttl)) {
		return nil
	}

	return entry
}

// SetTree caches a genealogy tree
func (gc *GenealogyCache) SetTree(jobID string, tree *JobGenealogy) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Enforce cache size limit
	if len(gc.trees) >= gc.maxSize {
		gc.evictOldestTree()
	}

	// Update timestamp and cache
	tree.UpdatedAt = time.Now()
	gc.trees[jobID] = tree
}

// DeleteTree removes a tree from cache
func (gc *GenealogyCache) DeleteTree(jobID string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	delete(gc.trees, jobID)
}

// GetLayout retrieves a cached layout
func (gc *GenealogyCache) GetLayout(key string) *TreeLayout {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	layout, exists := gc.layouts[key]
	if !exists || time.Now().After(layout.ComputedAt.Add(gc.ttl)) {
		return nil
	}

	return layout
}

// SetLayout caches a tree layout
func (gc *GenealogyCache) SetLayout(key string, layout *TreeLayout) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Enforce cache size limit
	if len(gc.layouts) >= gc.maxSize {
		gc.evictOldestLayout()
	}

	layout.ComputedAt = time.Now()
	gc.layouts[key] = layout
}

// GetRelationships retrieves cached relationships
func (gc *GenealogyCache) GetRelationships(jobID string) []JobRelationship {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return gc.relationships[jobID]
}

// SetRelationships caches relationships for a job
func (gc *GenealogyCache) SetRelationships(jobID string, relationships []JobRelationship) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Enforce cache size limit
	if len(gc.relationships) >= gc.maxSize {
		gc.evictOldestRelationships()
	}

	gc.relationships[jobID] = relationships
}

// Cleanup removes expired entries from all caches
func (gc *GenealogyCache) Cleanup() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-gc.ttl)

	// Clean trees
	for id, tree := range gc.trees {
		if tree.UpdatedAt.Before(cutoff) {
			delete(gc.trees, id)
		}
	}

	// Clean layouts
	for key, layout := range gc.layouts {
		if layout.ComputedAt.Before(cutoff) {
			delete(gc.layouts, key)
		}
	}

	// Clean relationships (they don't have timestamps, so just clear if cache is too big)
	if len(gc.relationships) > gc.maxSize {
		// Keep only most recently accessed (approximate)
		if len(gc.relationships) > gc.maxSize/2 {
			for id := range gc.relationships {
				delete(gc.relationships, id)
				if len(gc.relationships) <= gc.maxSize/2 {
					break
				}
			}
		}
	}
}

// Clear removes all cached data
func (gc *GenealogyCache) Clear() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.trees = make(map[string]*JobGenealogy)
	gc.layouts = make(map[string]*TreeLayout)
	gc.relationships = make(map[string][]JobRelationship)
}

// Stats returns cache statistics
func (gc *GenealogyCache) Stats() map[string]interface{} {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return map[string]interface{}{
		"trees_count":         len(gc.trees),
		"layouts_count":       len(gc.layouts),
		"relationships_count": len(gc.relationships),
		"max_size":           gc.maxSize,
		"ttl_seconds":        gc.ttl.Seconds(),
	}
}

// evictOldestTree removes the oldest tree entry
func (gc *GenealogyCache) evictOldestTree() {
	var oldestID string
	var oldestTime time.Time

	for id, tree := range gc.trees {
		if oldestTime.IsZero() || tree.UpdatedAt.Before(oldestTime) {
			oldestTime = tree.UpdatedAt
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(gc.trees, oldestID)
	}
}

// evictOldestLayout removes the oldest layout entry
func (gc *GenealogyCache) evictOldestLayout() {
	var oldestKey string
	var oldestTime time.Time

	for key, layout := range gc.layouts {
		if oldestTime.IsZero() || layout.ComputedAt.Before(oldestTime) {
			oldestTime = layout.ComputedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(gc.layouts, oldestKey)
	}
}

// evictOldestRelationships removes the oldest relationships entry
func (gc *GenealogyCache) evictOldestRelationships() {
	// Since relationships don't have timestamps, just remove a random entry
	for id := range gc.relationships {
		delete(gc.relationships, id)
		break
	}
}