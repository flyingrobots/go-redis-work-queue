// Copyright 2025 James Ross
package genealogy

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Navigator implements the GenealogyNavigator interface
type Navigator struct {
	config      GenealogyConfig
	graphStore  GraphStore
	jobProvider JobProvider
	renderer    TreeRenderer
	cache       *GenealogyCache
	logger      *zap.Logger

	// Navigation state
	navState    *NavigationState
	currentTree *JobGenealogy
	currentLayout *TreeLayout
	mu          sync.RWMutex

	// Event handling
	eventCallbacks []func(NavigationEvent)
	eventMu        sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewNavigator creates a new genealogy navigator
func NewNavigator(config GenealogyConfig, graphStore GraphStore, jobProvider JobProvider, logger *zap.Logger) *Navigator {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Navigator{
		config:      config,
		graphStore:  graphStore,
		jobProvider: jobProvider,
		cache:       NewGenealogyCache(1000, config.CacheTTL),
		logger:      logger,
		navState: &NavigationState{
			ViewMode:      config.DefaultViewMode,
			LayoutMode:    config.DefaultLayoutMode,
			ExpandedNodes: make(map[string]bool),
			FocusPath:     make([]string, 0),
		},
		eventCallbacks: make([]func(NavigationEvent), 0),
	}
}

// Start initializes the navigator
func (n *Navigator) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ctx, n.cancel = context.WithCancel(ctx)

	// Start background refresh if enabled
	if n.config.BackgroundRefresh {
		n.wg.Add(1)
		go n.backgroundRefreshLoop()
	}

	// Start cleanup routine
	n.wg.Add(1)
	go n.cleanupLoop()

	n.logger.Info("Genealogy navigator started",
		zap.String("view_mode", string(n.navState.ViewMode)),
		zap.String("layout_mode", string(n.navState.LayoutMode)))

	return nil
}

// Stop shuts down the navigator
func (n *Navigator) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.cancel != nil {
		n.cancel()
	}

	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		n.logger.Info("Genealogy navigator stopped gracefully")
	case <-time.After(5 * time.Second):
		n.logger.Warn("Timeout waiting for genealogy navigator to stop")
	}

	return nil
}

// GetGenealogy builds and returns the complete genealogy for a job
func (n *Navigator) GetGenealogy(ctx context.Context, jobID string) (*JobGenealogy, error) {
	// Check cache first
	if n.config.EnableCaching {
		if cached := n.cache.GetTree(jobID); cached != nil {
			n.logger.Debug("Genealogy cache hit", zap.String("job_id", jobID))
			return cached, nil
		}
	}

	// Build genealogy from graph store
	genealogy, err := n.graphStore.BuildGenealogy(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to build genealogy for job %s: %w", jobID, err)
	}

	// Populate job details
	if err := n.populateJobDetails(ctx, genealogy); err != nil {
		n.logger.Warn("Failed to populate some job details",
			zap.String("job_id", jobID),
			zap.Error(err))
	}

	// Cache the result
	if n.config.EnableCaching {
		n.cache.SetTree(jobID, genealogy)
	}

	n.logger.Debug("Built genealogy",
		zap.String("job_id", jobID),
		zap.Int("total_jobs", genealogy.TotalJobs),
		zap.Int("max_depth", genealogy.MaxDepth))

	return genealogy, nil
}

// GetImpactAnalysis analyzes the impact of a job's failure
func (n *Navigator) GetImpactAnalysis(ctx context.Context, jobID string) (*ImpactAnalysis, error) {
	genealogy, err := n.GetGenealogy(ctx, jobID)
	if err != nil {
		return nil, err
	}

	descendants, err := n.graphStore.GetDescendants(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}

	analysis := &ImpactAnalysis{
		JobID:            jobID,
		TotalDescendants: len(descendants),
		AffectedQueues:   make([]string, 0),
		CriticalPath:     make([]string, 0),
	}

	// Count direct children
	if node, exists := genealogy.Nodes[jobID]; exists {
		analysis.DirectChildren = len(node.ChildIDs)
	}

	// Analyze descendants
	failedCount := 0
	queueSet := make(map[string]bool)
	var totalCost time.Duration
	var minTime, maxTime time.Time

	for _, descendantID := range descendants {
		if node, exists := genealogy.Nodes[descendantID]; exists {
			if node.Status == JobStatusFailed {
				failedCount++
			}
			queueSet[node.QueueName] = true
			totalCost += node.Duration

			if minTime.IsZero() || node.CreatedAt.Before(minTime) {
				minTime = node.CreatedAt
			}
			if node.CompletedAt != nil && node.CompletedAt.After(maxTime) {
				maxTime = *node.CompletedAt
			}
		}
	}

	analysis.FailedDescendants = failedCount
	analysis.ProcessingCost = totalCost
	if !maxTime.IsZero() {
		analysis.TimeSpan = maxTime.Sub(minTime)
	}

	// Extract affected queues
	for queue := range queueSet {
		analysis.AffectedQueues = append(analysis.AffectedQueues, queue)
	}

	// Find critical path (longest dependency chain)
	analysis.CriticalPath = n.findCriticalPath(genealogy, jobID)

	return analysis, nil
}

// GetBlameAnalysis traces failure to root cause
func (n *Navigator) GetBlameAnalysis(ctx context.Context, failedJobID string) (*BlameAnalysis, error) {
	genealogy, err := n.GetGenealogy(ctx, failedJobID)
	if err != nil {
		return nil, err
	}

	failedNode, exists := genealogy.Nodes[failedJobID]
	if !exists {
		return nil, fmt.Errorf("failed job %s not found in genealogy", failedJobID)
	}

	if failedNode.Status != JobStatusFailed {
		return nil, fmt.Errorf("job %s is not in failed status", failedJobID)
	}

	// Trace blame path to root cause
	blamePath := n.traceBlamePath(genealogy, failedJobID)
	rootCauseID := blamePath[0] // First in path is root cause

	// Count retry attempts in chain
	retryCount := 0
	for i := 1; i < len(blamePath); i++ {
		// Check if this is a retry relationship
		for _, rel := range genealogy.Relationships {
			if rel.ChildID == blamePath[i] && rel.ParentID == blamePath[i-1] && rel.Type == RelationshipRetry {
				retryCount++
				break
			}
		}
	}

	// Calculate time to failure
	var timeToFailure time.Duration
	if rootNode, exists := genealogy.Nodes[rootCauseID]; exists {
		timeToFailure = failedNode.CreatedAt.Sub(rootNode.CreatedAt)
	}

	analysis := &BlameAnalysis{
		FailedJobID:   failedJobID,
		RootCauseID:   rootCauseID,
		BlamePath:     blamePath,
		FailureReason: failedNode.Error,
		TimeToFailure: timeToFailure,
		RetryAttempts: retryCount,
	}

	return analysis, nil
}

// FocusOnJob sets focus to a specific job
func (n *Navigator) FocusOnJob(ctx context.Context, jobID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get or build genealogy for this job
	genealogy, err := n.GetGenealogy(ctx, jobID)
	if err != nil {
		return err
	}

	// Update navigation state
	n.navState.CurrentJobID = jobID
	n.navState.FocusPath = []string{jobID}
	n.navState.SelectedNode = jobID
	n.currentTree = genealogy

	// Recompute layout
	if n.renderer != nil {
		layout, err := n.renderer.ComputeLayout(genealogy, n.navState.LayoutMode, n.renderer.GetViewport())
		if err != nil {
			n.logger.Warn("Failed to compute layout", zap.Error(err))
		} else {
			n.currentLayout = layout
		}
	}

	// Emit navigation event
	n.emitEvent(NavigationEvent{
		Type:      "focus_changed",
		JobID:     jobID,
		Timestamp: time.Now(),
	})

	n.logger.Debug("Focused on job", zap.String("job_id", jobID))
	return nil
}

// ExpandNode expands a node to show its children
func (n *Navigator) ExpandNode(ctx context.Context, jobID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.navState.ExpandedNodes[jobID] = true

	// If lazy loading is enabled, load children
	if n.config.LazyLoading {
		children, err := n.graphStore.GetChildren(ctx, jobID)
		if err != nil {
			return fmt.Errorf("failed to get children for job %s: %w", jobID, err)
		}

		// Load job details for children if not already in tree
		if n.currentTree != nil {
			missingJobs := make([]string, 0)
			for _, childID := range children {
				if _, exists := n.currentTree.Nodes[childID]; !exists {
					missingJobs = append(missingJobs, childID)
				}
			}

			if len(missingJobs) > 0 {
				jobs, err := n.jobProvider.GetJobs(ctx, missingJobs)
				if err != nil {
					n.logger.Warn("Failed to load some child jobs", zap.Error(err))
				} else {
					for id, job := range jobs {
						n.currentTree.Nodes[id] = job
					}
					n.currentTree.TotalJobs += len(jobs)
				}
			}
		}
	}

	// Recompute layout if renderer is available
	if n.renderer != nil && n.currentTree != nil {
		layout, err := n.renderer.ComputeLayout(n.currentTree, n.navState.LayoutMode, n.renderer.GetViewport())
		if err != nil {
			n.logger.Warn("Failed to recompute layout", zap.Error(err))
		} else {
			n.currentLayout = layout
		}
	}

	// Emit navigation event
	n.emitEvent(NavigationEvent{
		Type:      "node_expanded",
		JobID:     jobID,
		Timestamp: time.Now(),
	})

	return nil
}

// CollapseNode collapses a node to hide its children
func (n *Navigator) CollapseNode(ctx context.Context, jobID string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.navState.ExpandedNodes[jobID] = false

	// Recompute layout if renderer is available
	if n.renderer != nil && n.currentTree != nil {
		layout, err := n.renderer.ComputeLayout(n.currentTree, n.navState.LayoutMode, n.renderer.GetViewport())
		if err != nil {
			n.logger.Warn("Failed to recompute layout", zap.Error(err))
		} else {
			n.currentLayout = layout
		}
	}

	// Emit navigation event
	n.emitEvent(NavigationEvent{
		Type:      "node_collapsed",
		JobID:     jobID,
		Timestamp: time.Now(),
	})

	return nil
}

// SetViewMode changes the current view mode
func (n *Navigator) SetViewMode(mode ViewMode) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.navState.ViewMode = mode

	// Recompute layout if needed
	if n.renderer != nil && n.currentTree != nil {
		layout, err := n.renderer.ComputeLayout(n.currentTree, n.navState.LayoutMode, n.renderer.GetViewport())
		if err != nil {
			n.logger.Warn("Failed to recompute layout for view mode change", zap.Error(err))
		} else {
			n.currentLayout = layout
		}
	}

	// Emit event
	n.emitEvent(NavigationEvent{
		Type:      "view_mode_changed",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"mode": string(mode)},
	})

	return nil
}

// SetLayoutMode changes the current layout mode
func (n *Navigator) SetLayoutMode(mode LayoutMode) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.navState.LayoutMode = mode

	// Recompute layout
	if n.renderer != nil && n.currentTree != nil {
		layout, err := n.renderer.ComputeLayout(n.currentTree, mode, n.renderer.GetViewport())
		if err != nil {
			n.logger.Warn("Failed to recompute layout for layout mode change", zap.Error(err))
		} else {
			n.currentLayout = layout
		}
	}

	// Emit event
	n.emitEvent(NavigationEvent{
		Type:      "layout_mode_changed",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"mode": string(mode)},
	})

	return nil
}

// Navigate handles directional navigation
func (n *Navigator) Navigate(direction string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.currentTree == nil || n.navState.SelectedNode == "" {
		return fmt.Errorf("no tree or node selected")
	}

	currentNode, exists := n.currentTree.Nodes[n.navState.SelectedNode]
	if !exists {
		return fmt.Errorf("selected node not found in tree")
	}

	var targetID string

	switch direction {
	case "up", "parent":
		if currentNode.ParentID != "" {
			targetID = currentNode.ParentID
		}
	case "down", "child":
		if len(currentNode.ChildIDs) > 0 {
			targetID = currentNode.ChildIDs[0] // Select first child
		}
	case "left", "previous_sibling":
		targetID = n.findPreviousSibling(currentNode)
	case "right", "next_sibling":
		targetID = n.findNextSibling(currentNode)
	default:
		return fmt.Errorf("unknown navigation direction: %s", direction)
	}

	if targetID == "" {
		return fmt.Errorf("no target found for navigation direction: %s", direction)
	}

	// Update selection
	n.navState.SelectedNode = targetID
	n.navState.FocusPath = append(n.navState.FocusPath, targetID)

	// Keep focus path reasonable length
	if len(n.navState.FocusPath) > 20 {
		n.navState.FocusPath = n.navState.FocusPath[1:]
	}

	// Emit navigation event
	n.emitEvent(NavigationEvent{
		Type:      "navigation",
		JobID:     targetID,
		Direction: direction,
		Timestamp: time.Now(),
	})

	return nil
}

// Search finds jobs matching the search term
func (n *Navigator) Search(term string) ([]*JobNode, error) {
	n.mu.RLock()
	currentTree := n.currentTree
	n.mu.RUnlock()

	if currentTree == nil {
		return nil, fmt.Errorf("no tree loaded")
	}

	term = strings.ToLower(term)
	matches := make([]*JobNode, 0)

	for _, node := range currentTree.Nodes {
		if strings.Contains(strings.ToLower(node.Name), term) ||
		   strings.Contains(strings.ToLower(node.ID), term) ||
		   strings.Contains(strings.ToLower(node.QueueName), term) {
			matches = append(matches, node)
		}
	}

	// Sort by relevance (exact matches first, then by creation time)
	sort.Slice(matches, func(i, j int) bool {
		iExact := strings.ToLower(matches[i].Name) == term || strings.ToLower(matches[i].ID) == term
		jExact := strings.ToLower(matches[j].Name) == term || strings.ToLower(matches[j].ID) == term

		if iExact && !jExact {
			return true
		}
		if !iExact && jExact {
			return false
		}

		return matches[i].CreatedAt.After(matches[j].CreatedAt)
	})

	// Update search state
	n.mu.Lock()
	n.navState.SearchTerm = term
	n.mu.Unlock()

	return matches, nil
}

// RenderCurrent renders the current tree state
func (n *Navigator) RenderCurrent() ([]string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.renderer == nil {
		return nil, fmt.Errorf("no renderer configured")
	}

	if n.currentTree == nil {
		return []string{"No genealogy loaded"}, nil
	}

	if n.currentLayout == nil {
		// Compute layout if not available
		layout, err := n.renderer.ComputeLayout(n.currentTree, n.navState.LayoutMode, n.renderer.GetViewport())
		if err != nil {
			return nil, fmt.Errorf("failed to compute layout: %w", err)
		}
		n.currentLayout = layout
	}

	return n.renderer.RenderTree(n.currentTree, n.currentLayout, n.navState)
}

// RefreshTree reloads the current tree from storage
func (n *Navigator) RefreshTree() error {
	n.mu.Lock()
	currentJobID := n.navState.CurrentJobID
	n.mu.Unlock()

	if currentJobID == "" {
		return fmt.Errorf("no current job to refresh")
	}

	// Clear cache for this tree
	if n.config.EnableCaching {
		n.cache.DeleteTree(currentJobID)
	}

	// Reload genealogy
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return n.FocusOnJob(ctx, currentJobID)
}

// GetNavigationState returns the current navigation state
func (n *Navigator) GetNavigationState() *NavigationState {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Return a copy to prevent external modification
	state := *n.navState
	state.ExpandedNodes = make(map[string]bool)
	for k, v := range n.navState.ExpandedNodes {
		state.ExpandedNodes[k] = v
	}
	state.FocusPath = make([]string, len(n.navState.FocusPath))
	copy(state.FocusPath, n.navState.FocusPath)

	return &state
}

// SetNavigationState updates the navigation state
func (n *Navigator) SetNavigationState(state *NavigationState) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.navState = state
	return nil
}

// Subscribe registers a callback for navigation events
func (n *Navigator) Subscribe(callback func(NavigationEvent)) error {
	n.eventMu.Lock()
	defer n.eventMu.Unlock()

	n.eventCallbacks = append(n.eventCallbacks, callback)
	return nil
}

// Unsubscribe removes all event callbacks
func (n *Navigator) Unsubscribe() error {
	n.eventMu.Lock()
	defer n.eventMu.Unlock()

	n.eventCallbacks = n.eventCallbacks[:0]
	return nil
}

// Helper methods

func (n *Navigator) populateJobDetails(ctx context.Context, genealogy *JobGenealogy) error {
	jobIDs := make([]string, 0, len(genealogy.Nodes))
	for id := range genealogy.Nodes {
		jobIDs = append(jobIDs, id)
	}

	jobs, err := n.jobProvider.GetJobs(ctx, jobIDs)
	if err != nil {
		return err
	}

	for id, job := range jobs {
		genealogy.Nodes[id] = job
	}

	return nil
}

func (n *Navigator) findCriticalPath(genealogy *JobGenealogy, startID string) []string {
	// Use DFS to find the longest path from start node
	visited := make(map[string]bool)
	var longestPath []string

	var dfs func(nodeID string, currentPath []string)
	dfs = func(nodeID string, currentPath []string) {
		if visited[nodeID] {
			return
		}
		visited[nodeID] = true
		currentPath = append(currentPath, nodeID)

		node, exists := genealogy.Nodes[nodeID]
		if !exists || len(node.ChildIDs) == 0 {
			// Leaf node - check if this path is longer
			if len(currentPath) > len(longestPath) {
				longestPath = make([]string, len(currentPath))
				copy(longestPath, currentPath)
			}
			return
		}

		// Continue DFS on children
		for _, childID := range node.ChildIDs {
			dfs(childID, currentPath)
		}
	}

	dfs(startID, []string{})
	return longestPath
}

func (n *Navigator) traceBlamePath(genealogy *JobGenealogy, failedJobID string) []string {
	// Trace backwards to find root cause
	path := []string{failedJobID}
	currentID := failedJobID

	for {
		node, exists := genealogy.Nodes[currentID]
		if !exists || node.ParentID == "" {
			break
		}

		path = append([]string{node.ParentID}, path...)
		currentID = node.ParentID
	}

	return path
}

func (n *Navigator) findPreviousSibling(node *JobNode) string {
	if node.ParentID == "" || n.currentTree == nil {
		return ""
	}

	parent, exists := n.currentTree.Nodes[node.ParentID]
	if !exists {
		return ""
	}

	// Find current node position in siblings
	for i, childID := range parent.ChildIDs {
		if childID == node.ID && i > 0 {
			return parent.ChildIDs[i-1]
		}
	}

	return ""
}

func (n *Navigator) findNextSibling(node *JobNode) string {
	if node.ParentID == "" || n.currentTree == nil {
		return ""
	}

	parent, exists := n.currentTree.Nodes[node.ParentID]
	if !exists {
		return ""
	}

	// Find current node position in siblings
	for i, childID := range parent.ChildIDs {
		if childID == node.ID && i < len(parent.ChildIDs)-1 {
			return parent.ChildIDs[i+1]
		}
	}

	return ""
}

func (n *Navigator) emitEvent(event NavigationEvent) {
	n.eventMu.RLock()
	callbacks := make([]func(NavigationEvent), len(n.eventCallbacks))
	copy(callbacks, n.eventCallbacks)
	n.eventMu.RUnlock()

	for _, callback := range callbacks {
		go callback(event)
	}
}

func (n *Navigator) backgroundRefreshLoop() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			if err := n.RefreshTree(); err != nil {
				n.logger.Debug("Background refresh failed", zap.Error(err))
			}
		}
	}
}

func (n *Navigator) cleanupLoop() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-n.config.RelationshipTTL)
			if err := n.graphStore.Cleanup(n.ctx, cutoff); err != nil {
				n.logger.Warn("Cleanup failed", zap.Error(err))
			}

			// Clean cache
			n.cache.Cleanup()
		}
	}
}