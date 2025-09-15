// Copyright 2025 James Ross
package genealogy

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ASCIIRenderer implements TreeRenderer for ASCII art output
type ASCIIRenderer struct {
	viewport    Viewport
	colorScheme ColorScheme
	config      GenealogyConfig
}

// NewASCIIRenderer creates a new ASCII tree renderer
func NewASCIIRenderer(config GenealogyConfig, viewport Viewport) *ASCIIRenderer {
	return &ASCIIRenderer{
		viewport:    viewport,
		colorScheme: DefaultColorScheme(),
		config:      config,
	}
}

// BoxDrawingChars defines Unicode box drawing characters
type BoxDrawingChars struct {
	Vertical   string
	Horizontal string
	Corner     string
	Branch     string
	LastBranch string
	Expansion  string
	Collapse   string
}

// DefaultBoxChars returns standard Unicode box drawing characters
func DefaultBoxChars() BoxDrawingChars {
	return BoxDrawingChars{
		Vertical:   "│",
		Horizontal: "─",
		Corner:     "┌",
		Branch:     "├",
		LastBranch: "└",
		Expansion:  "+",
		Collapse:   "-",
	}
}

// RenderTree renders the genealogy tree as ASCII art
func (r *ASCIIRenderer) RenderTree(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error) {
	if genealogy == nil || len(genealogy.Nodes) == 0 {
		return []string{"No jobs in genealogy"}, nil
	}

	switch state.LayoutMode {
	case LayoutModeTopDown:
		return r.renderTopDown(genealogy, layout, state)
	case LayoutModeTimeline:
		return r.renderTimeline(genealogy, layout, state)
	case LayoutModeRadial:
		return r.renderRadial(genealogy, layout, state)
	case LayoutModeCompact:
		return r.renderCompact(genealogy, layout, state)
	default:
		return r.renderTopDown(genealogy, layout, state)
	}
}

// ComputeLayout calculates the positioning of nodes for rendering
func (r *ASCIIRenderer) ComputeLayout(genealogy *JobGenealogy, mode LayoutMode, viewport Viewport) (*TreeLayout, error) {
	layout := &TreeLayout{
		Nodes:          make(map[string]*TreeLayoutNode),
		LayoutMode:     mode,
		ViewportWidth:  viewport.Width,
		ViewportHeight: viewport.Height,
		ComputedAt:     time.Now(),
	}

	switch mode {
	case LayoutModeTopDown:
		return r.computeTopDownLayout(genealogy, layout, viewport)
	case LayoutModeTimeline:
		return r.computeTimelineLayout(genealogy, layout, viewport)
	case LayoutModeRadial:
		return r.computeRadialLayout(genealogy, layout, viewport)
	case LayoutModeCompact:
		return r.computeCompactLayout(genealogy, layout, viewport)
	default:
		return r.computeTopDownLayout(genealogy, layout, viewport)
	}
}

// GetViewport returns the current viewport
func (r *ASCIIRenderer) GetViewport() Viewport {
	return r.viewport
}

// SetViewport updates the viewport
func (r *ASCIIRenderer) SetViewport(viewport Viewport) {
	r.viewport = viewport
}

// renderTopDown renders tree in traditional top-down format
func (r *ASCIIRenderer) renderTopDown(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error) {
	if genealogy.RootID == "" {
		return []string{"No root job found"}, nil
	}

	lines := make([]string, 0)
	boxes := DefaultBoxChars()

	// Start rendering from root
	visited := make(map[string]bool)
	r.renderNodeRecursive(genealogy, genealogy.RootID, "", true, state, boxes, &lines, visited, 0)

	// Add viewport scrolling
	return r.applyViewport(lines, state), nil
}

// renderNodeRecursive recursively renders a node and its children
func (r *ASCIIRenderer) renderNodeRecursive(genealogy *JobGenealogy, nodeID string, prefix string, isLast bool, state *NavigationState, boxes BoxDrawingChars, lines *[]string, visited map[string]bool, depth int) {
	if visited[nodeID] || depth > 20 { // Prevent infinite recursion
		return
	}
	visited[nodeID] = true

	node, exists := genealogy.Nodes[nodeID]
	if !exists {
		return
	}

	// Should this node be visible based on view mode?
	if !r.shouldShowNode(node, state) {
		return
	}

	// Build line prefix
	var linePrefix, childPrefix string
	if depth == 0 {
		linePrefix = boxes.Corner + boxes.Horizontal + " "
		childPrefix = ""
	} else {
		if isLast {
			linePrefix = prefix + boxes.LastBranch + boxes.Horizontal + " "
			childPrefix = prefix + "  "
		} else {
			linePrefix = prefix + boxes.Branch + boxes.Horizontal + " "
			childPrefix = prefix + boxes.Vertical + " "
		}
	}

	// Render node
	nodeStr := r.formatNode(node, state)
	isExpanded := state.ExpandedNodes[nodeID]
	hasChildren := len(node.ChildIDs) > 0

	// Add expansion indicator
	if hasChildren {
		if isExpanded {
			nodeStr = linePrefix + boxes.Collapse + " " + nodeStr
		} else {
			nodeStr = linePrefix + boxes.Expansion + " " + nodeStr
		}
	} else {
		nodeStr = linePrefix + "  " + nodeStr
	}

	*lines = append(*lines, nodeStr)

	// Render children if expanded
	if isExpanded && hasChildren {
		// Sort children by creation time or ID for consistent ordering
		children := make([]*JobNode, 0, len(node.ChildIDs))
		for _, childID := range node.ChildIDs {
			if child, exists := genealogy.Nodes[childID]; exists {
				children = append(children, child)
			}
		}

		sort.Slice(children, func(i, j int) bool {
			return children[i].CreatedAt.Before(children[j].CreatedAt)
		})

		for i, child := range children {
			isLastChild := i == len(children)-1
			r.renderNodeRecursive(genealogy, child.ID, childPrefix, isLastChild, state, boxes, lines, visited, depth+1)
		}
	}
}

// renderTimeline renders jobs in chronological order
func (r *ASCIIRenderer) renderTimeline(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error) {
	// Collect all nodes and sort by creation time
	nodes := make([]*JobNode, 0, len(genealogy.Nodes))
	for _, node := range genealogy.Nodes {
		if r.shouldShowNode(node, state) {
			nodes = append(nodes, node)
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].CreatedAt.Before(nodes[j].CreatedAt)
	})

	lines := make([]string, 0)
	boxes := DefaultBoxChars()

	// Group nodes by time periods
	var currentTime time.Time
	var timeGroupLines []string

	for _, node := range nodes {
		// Check if we need a new time group
		if currentTime.IsZero() || node.CreatedAt.Sub(currentTime) > time.Minute {
			if len(timeGroupLines) > 0 {
				lines = append(lines, timeGroupLines...)
				timeGroupLines = make([]string, 0)
			}
			currentTime = node.CreatedAt
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("=== %s ===", currentTime.Format("15:04:05")))
		}

		// Add relationship indicator
		relationshipStr := ""
		for _, rel := range genealogy.Relationships {
			if rel.ChildID == node.ID {
				relationshipStr = fmt.Sprintf(" (%s)", rel.Type)
				break
			}
		}

		nodeStr := r.formatNode(node, state) + relationshipStr
		timeGroupLines = append(timeGroupLines, boxes.Branch+boxes.Horizontal+" "+nodeStr)
	}

	if len(timeGroupLines) > 0 {
		lines = append(lines, timeGroupLines...)
	}

	return r.applyViewport(lines, state), nil
}

// renderRadial renders tree in radial layout (simplified for ASCII)
func (r *ASCIIRenderer) renderRadial(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error) {
	// For ASCII radial, we'll create a simplified version with the root in center
	lines := make([]string, 0)

	if genealogy.RootID == "" {
		return []string{"No root job found"}, nil
	}

	rootNode := genealogy.Nodes[genealogy.RootID]
	if rootNode == nil {
		return []string{"Root job not found in genealogy"}, nil
	}

	// Center the root
	centerY := r.viewport.Height / 2
	centerX := r.viewport.Width / 2

	// Initialize grid
	grid := make([][]rune, r.viewport.Height)
	for i := range grid {
		grid[i] = make([]rune, r.viewport.Width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Place root at center
	rootStr := r.formatNode(rootNode, state)
	if centerY >= 0 && centerY < len(grid) {
		r.placeStringInGrid(grid, centerX-len(rootStr)/2, centerY, rootStr)
	}

	// Place children around root
	if len(rootNode.ChildIDs) > 0 {
		angleStep := 360.0 / float64(len(rootNode.ChildIDs))
		radius := 5

		for i, childID := range rootNode.ChildIDs {
			if child, exists := genealogy.Nodes[childID]; exists && r.shouldShowNode(child, state) {
				angle := float64(i) * angleStep
				x := centerX + int(float64(radius)*1.5) // ASCII is wider than tall
				y := centerY + int(float64(radius)*0.8)

				if i%4 == 1 {
					x = centerX + radius*2
				} else if i%4 == 2 {
					y = centerY + radius
				} else if i%4 == 3 {
					x = centerX - radius*2
				} else {
					y = centerY - radius
				}

				childStr := r.formatNode(child, state)
				if y >= 0 && y < len(grid) {
					r.placeStringInGrid(grid, x-len(childStr)/2, y, childStr)
				}
			}
		}
	}

	// Convert grid to lines
	for _, row := range grid {
		lines = append(lines, strings.TrimRight(string(row), " "))
	}

	return lines, nil
}

// renderCompact renders tree in compact format
func (r *ASCIIRenderer) renderCompact(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error) {
	lines := make([]string, 0)

	// Group by generation
	for generation := 0; generation <= genealogy.MaxDepth; generation++ {
		generationNodes, exists := genealogy.GenerationMap[generation]
		if !exists || len(generationNodes) == 0 {
			continue
		}

		lines = append(lines, fmt.Sprintf("Generation %d:", generation))

		for _, nodeID := range generationNodes {
			if node, exists := genealogy.Nodes[nodeID]; exists && r.shouldShowNode(node, state) {
				nodeStr := r.formatNodeCompact(node, state)
				lines = append(lines, "  "+nodeStr)
			}
		}

		lines = append(lines, "")
	}

	return r.applyViewport(lines, state), nil
}

// formatNode formats a job node for display
func (r *ASCIIRenderer) formatNode(node *JobNode, state *NavigationState) string {
	statusChar := r.getStatusChar(node.Status)

	// Highlight selected node
	prefix := ""
	suffix := ""
	if node.ID == state.SelectedNode {
		prefix = "→ "
		suffix = " ←"
	}

	// Format basic info
	duration := ""
	if node.Duration > 0 {
		if node.Duration < time.Second {
			duration = fmt.Sprintf("%dms", node.Duration.Milliseconds())
		} else {
			duration = fmt.Sprintf("%.1fs", node.Duration.Seconds())
		}
	}

	// Create main display string
	display := fmt.Sprintf("%s%s %s [%s]%s",
		prefix,
		statusChar,
		r.truncateString(node.Name, 25),
		node.Status,
		suffix)

	if duration != "" {
		display += fmt.Sprintf(" (%s)", duration)
	}

	// Add error info for failed jobs
	if node.Status == JobStatusFailed && node.Error != "" {
		display += fmt.Sprintf(" - %s", r.truncateString(node.Error, 30))
	}

	return display
}

// formatNodeCompact formats a node in compact display
func (r *ASCIIRenderer) formatNodeCompact(node *JobNode, state *NavigationState) string {
	statusChar := r.getStatusChar(node.Status)

	selected := ""
	if node.ID == state.SelectedNode {
		selected = "→ "
	}

	return fmt.Sprintf("%s%s %s (%s)",
		selected,
		statusChar,
		r.truncateString(node.Name, 20),
		node.QueueName)
}

// getStatusChar returns a character representing job status
func (r *ASCIIRenderer) getStatusChar(status JobStatus) string {
	switch status {
	case JobStatusSuccess:
		return "✓"
	case JobStatusFailed:
		return "✗"
	case JobStatusProcessing:
		return "⚡"
	case JobStatusPending:
		return "○"
	case JobStatusRetry:
		return "↻"
	case JobStatusCancelled:
		return "⊘"
	default:
		return "?"
	}
}

// shouldShowNode determines if a node should be visible based on view mode and filters
func (r *ASCIIRenderer) shouldShowNode(node *JobNode, state *NavigationState) bool {
	// Apply status filter
	if len(state.FilterStatus) > 0 {
		found := false
		for _, status := range state.FilterStatus {
			if node.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Apply queue filter
	if len(state.FilterQueues) > 0 {
		found := false
		for _, queue := range state.FilterQueues {
			if node.QueueName == queue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Apply search filter
	if state.SearchTerm != "" {
		term := strings.ToLower(state.SearchTerm)
		if !strings.Contains(strings.ToLower(node.Name), term) &&
		   !strings.Contains(strings.ToLower(node.ID), term) &&
		   !strings.Contains(strings.ToLower(node.QueueName), term) {
			return false
		}
	}

	return true
}

// truncateString truncates a string to the specified length
func (r *ASCIIRenderer) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// placeStringInGrid places a string at the specified position in the grid
func (r *ASCIIRenderer) placeStringInGrid(grid [][]rune, x, y int, s string) {
	if y < 0 || y >= len(grid) {
		return
	}

	runes := []rune(s)
	for i, r := range runes {
		if x+i >= 0 && x+i < len(grid[y]) {
			grid[y][x+i] = r
		}
	}
}

// applyViewport applies scrolling and viewport clipping
func (r *ASCIIRenderer) applyViewport(lines []string, state *NavigationState) []string {
	if len(lines) == 0 {
		return lines
	}

	// Apply vertical scrolling
	startY := state.ScrollOffset
	endY := startY + r.viewport.Height

	if startY < 0 {
		startY = 0
	}
	if endY > len(lines) {
		endY = len(lines)
	}
	if startY >= len(lines) {
		return []string{}
	}

	visibleLines := lines[startY:endY]

	// Apply horizontal clipping
	for i, line := range visibleLines {
		if len(line) > r.viewport.Width {
			visibleLines[i] = line[:r.viewport.Width]
		}
	}

	return visibleLines
}

// Layout computation methods

func (r *ASCIIRenderer) computeTopDownLayout(genealogy *JobGenealogy, layout *TreeLayout, viewport Viewport) (*TreeLayout, error) {
	// For ASCII rendering, layout is mostly virtual since we render on-demand
	// We'll compute basic positioning for reference

	for nodeID, node := range genealogy.Nodes {
		layoutNode := &TreeLayoutNode{
			JobID:       nodeID,
			X:           node.TreePosition * (r.config.NodeWidth + r.config.HorizontalSpacing),
			Y:           node.Generation * (r.config.NodeHeight + r.config.VerticalSpacing),
			Width:       r.config.NodeWidth,
			Height:      r.config.NodeHeight,
			Level:       node.Generation,
			HasChildren: len(node.ChildIDs) > 0,
			IsVisible:   true,
		}

		layout.Nodes[nodeID] = layoutNode
	}

	// Calculate total dimensions
	maxX, maxY := 0, 0
	for _, node := range layout.Nodes {
		if node.X+node.Width > maxX {
			maxX = node.X + node.Width
		}
		if node.Y+node.Height > maxY {
			maxY = node.Y + node.Height
		}
	}

	layout.TotalWidth = maxX
	layout.TotalHeight = maxY

	return layout, nil
}

func (r *ASCIIRenderer) computeTimelineLayout(genealogy *JobGenealogy, layout *TreeLayout, viewport Viewport) (*TreeLayout, error) {
	// Timeline layout arranges nodes chronologically
	nodes := make([]*JobNode, 0, len(genealogy.Nodes))
	for _, node := range genealogy.Nodes {
		nodes = append(nodes, node)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].CreatedAt.Before(nodes[j].CreatedAt)
	})

	y := 0
	for i, node := range nodes {
		layoutNode := &TreeLayoutNode{
			JobID:       node.ID,
			X:           10,
			Y:           y,
			Width:       r.config.NodeWidth,
			Height:      r.config.NodeHeight,
			Level:       i,
			HasChildren: len(node.ChildIDs) > 0,
			IsVisible:   true,
		}

		layout.Nodes[node.ID] = layoutNode
		y += r.config.NodeHeight + r.config.VerticalSpacing
	}

	layout.TotalWidth = r.config.NodeWidth + 20
	layout.TotalHeight = y

	return layout, nil
}

func (r *ASCIIRenderer) computeRadialLayout(genealogy *JobGenealogy, layout *TreeLayout, viewport Viewport) (*TreeLayout, error) {
	centerX := viewport.Width / 2
	centerY := viewport.Height / 2

	// Place root at center
	if rootNode, exists := genealogy.Nodes[genealogy.RootID]; exists {
		layout.Nodes[genealogy.RootID] = &TreeLayoutNode{
			JobID:       genealogy.RootID,
			X:           centerX,
			Y:           centerY,
			Width:       r.config.NodeWidth,
			Height:      r.config.NodeHeight,
			Level:       0,
			HasChildren: len(rootNode.ChildIDs) > 0,
			IsVisible:   true,
		}

		// Place children in a circle around root
		if len(rootNode.ChildIDs) > 0 {
			radius := 10
			angleStep := 360.0 / float64(len(rootNode.ChildIDs))

			for i, childID := range rootNode.ChildIDs {
				angle := float64(i) * angleStep
				x := centerX + int(float64(radius)*1.5) // ASCII aspect ratio adjustment
				y := centerY + radius

				// Simplified positioning for ASCII
				switch i % 4 {
				case 0:
					x, y = centerX, centerY-radius
				case 1:
					x, y = centerX+radius*2, centerY
				case 2:
					x, y = centerX, centerY+radius
				case 3:
					x, y = centerX-radius*2, centerY
				}

				if child, exists := genealogy.Nodes[childID]; exists {
					layout.Nodes[childID] = &TreeLayoutNode{
						JobID:       childID,
						X:           x,
						Y:           y,
						Width:       r.config.NodeWidth,
						Height:      r.config.NodeHeight,
						Level:       1,
						HasChildren: len(child.ChildIDs) > 0,
						IsVisible:   true,
					}
				}
			}
		}
	}

	layout.TotalWidth = viewport.Width
	layout.TotalHeight = viewport.Height

	return layout, nil
}

func (r *ASCIIRenderer) computeCompactLayout(genealogy *JobGenealogy, layout *TreeLayout, viewport Viewport) (*TreeLayout, error) {
	// Compact layout is similar to top-down but with minimal spacing
	y := 0
	for generation := 0; generation <= genealogy.MaxDepth; generation++ {
		if nodes, exists := genealogy.GenerationMap[generation]; exists {
			x := 0
			for _, nodeID := range nodes {
				layout.Nodes[nodeID] = &TreeLayoutNode{
					JobID:       nodeID,
					X:           x,
					Y:           y,
					Width:       r.config.NodeWidth,
					Height:      1, // Compact height
					Level:       generation,
					HasChildren: len(genealogy.Nodes[nodeID].ChildIDs) > 0,
					IsVisible:   true,
				}
				x += r.config.NodeWidth + 1
			}
			y += 2 // Minimal vertical spacing
		}
	}

	layout.TotalHeight = y

	return layout, nil
}