// Copyright 2025 James Ross
package genealogy

import (
	"context"
	"sync"
	"time"
)

// RelationshipType represents the type of relationship between jobs
type RelationshipType string

const (
	RelationshipRetry        RelationshipType = "retry"        // Job is retry of parent
	RelationshipSpawn        RelationshipType = "spawn"        // Child job created by parent
	RelationshipFork         RelationshipType = "fork"         // Parallel processing split
	RelationshipCallback     RelationshipType = "callback"     // Callback job triggered by parent completion
	RelationshipCompensation RelationshipType = "compensation" // Cleanup/rollback job triggered by parent failure
	RelationshipContinuation RelationshipType = "continuation" // Sequential processing step
	RelationshipBatchMember  RelationshipType = "batch_member" // Part of same batch operation
)

func (rt RelationshipType) String() string {
	return string(rt)
}

// ViewMode represents different visualization modes
type ViewMode string

const (
	ViewModeFull        ViewMode = "full"        // Show complete genealogy
	ViewModeAncestors   ViewMode = "ancestors"   // Trace backwards to root causes
	ViewModeDescendants ViewMode = "descendants" // Show forward impact
	ViewModeBlamePath   ViewMode = "blame"       // Highlight path from failure to root cause
	ViewModeImpactZone  ViewMode = "impact"      // Highlight all affected descendants
)

// LayoutMode represents different tree layout algorithms
type LayoutMode string

const (
	LayoutModeTopDown LayoutMode = "topdown" // Root jobs at top, children below
	LayoutModeTimeline LayoutMode = "timeline" // Jobs arranged chronologically
	LayoutModeRadial   LayoutMode = "radial"   // Root job at center, radiating outward
	LayoutModeCompact  LayoutMode = "compact"  // Minimized vertical spacing
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusSuccess    JobStatus = "success"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetry      JobStatus = "retry"
	JobStatusCancelled  JobStatus = "cancelled"
)

// JobRelationship represents a parent-child relationship between jobs
type JobRelationship struct {
	ParentID     string                 `json:"parent_id"`
	ChildID      string                 `json:"child_id"`
	Type         RelationshipType       `json:"relationship_type"`
	SpawnReason  string                 `json:"spawn_reason"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// JobNode represents a job in the genealogy tree
// JobDetails represents detailed information about a job (for compatibility with tests)
type JobDetails struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      JobStatus              `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Queue       string                 `json:"queue,omitempty"`
	Priority    int                    `json:"priority"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Error       string                 `json:"error,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type JobNode struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Status       JobStatus              `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Duration     time.Duration          `json:"duration"`
	AttemptCount int                    `json:"attempt_count"`
	Priority     int                    `json:"priority"`
	QueueName    string                 `json:"queue_name"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Error        string                 `json:"error,omitempty"`
	JobDetails   *JobDetails            `json:"job_details,omitempty"`

	// Tree structure
	ParentID     string   `json:"parent_id,omitempty"`
	ChildIDs     []string `json:"child_ids,omitempty"`
	Generation   int      `json:"generation"`
	TreePosition int      `json:"tree_position"`
}

// JobGenealogy represents a complete family tree of jobs
type JobGenealogy struct {
	RootID         string                       `json:"root_id"`
	Nodes          map[string]*JobNode          `json:"nodes"`
	Relationships  []JobRelationship            `json:"relationships"`
	GenerationMap  map[int][]string             `json:"generation_map"`
	MaxDepth       int                          `json:"max_depth"`
	TotalJobs      int                          `json:"total_jobs"`
	CreatedAt      time.Time                    `json:"created_at"`
	UpdatedAt      time.Time                    `json:"updated_at"`
}

// ImpactAnalysis represents the analysis of job impact
type ImpactAnalysis struct {
	JobID             string        `json:"job_id"`
	DirectChildren    int           `json:"direct_children"`
	TotalDescendants  int           `json:"total_descendants"`
	FailedDescendants int           `json:"failed_descendants"`
	ProcessingCost    time.Duration `json:"processing_cost"`
	TimeSpan          time.Duration `json:"time_span"`
	AffectedQueues    []string      `json:"affected_queues"`
	CriticalPath      []string      `json:"critical_path"`
}

// BlameAnalysis represents root cause analysis
type BlameAnalysis struct {
	FailedJobID    string   `json:"failed_job_id"`
	RootCauseID    string   `json:"root_cause_id"`
	BlamePath      []string `json:"blame_path"`
	FailureReason  string   `json:"failure_reason"`
	TimeToFailure  time.Duration `json:"time_to_failure"`
	RetryAttempts  int      `json:"retry_attempts"`
	PatternMatch   string   `json:"pattern_match,omitempty"`
}

// NavigationState holds the current state of tree navigation
type NavigationState struct {
	CurrentJobID   string            `json:"current_job_id"`
	ViewMode       ViewMode          `json:"view_mode"`
	LayoutMode     LayoutMode        `json:"layout_mode"`
	ExpandedNodes  map[string]bool   `json:"expanded_nodes"`
	FocusPath      []string          `json:"focus_path"`
	FilterStatus   []JobStatus       `json:"filter_status,omitempty"`
	FilterQueues   []string          `json:"filter_queues,omitempty"`
	SearchTerm     string            `json:"search_term,omitempty"`
	ScrollOffset   int               `json:"scroll_offset"`
	SelectedNode   string            `json:"selected_node,omitempty"`
}

// TreeLayoutNode represents a positioned node in the rendered tree
type TreeLayoutNode struct {
	JobID      string `json:"job_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Level      int    `json:"level"`
	IsExpanded bool   `json:"is_expanded"`
	HasChildren bool  `json:"has_children"`
	IsVisible  bool   `json:"is_visible"`
}

// TreeLayout represents the computed layout of a genealogy tree
type TreeLayout struct {
	Nodes         map[string]*TreeLayoutNode `json:"nodes"`
	TotalWidth    int                        `json:"total_width"`
	TotalHeight   int                        `json:"total_height"`
	ViewportWidth int                        `json:"viewport_width"`
	ViewportHeight int                       `json:"viewport_height"`
	LayoutMode    LayoutMode                 `json:"layout_mode"`
	ComputedAt    time.Time                  `json:"computed_at"`
}

// GenealogyConfig configures the genealogy navigator
type GenealogyConfig struct {
	// Storage configuration
	RedisKeyPrefix    string        `json:"redis_key_prefix"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	MaxTreeSize       int           `json:"max_tree_size"`
	MaxGenerations    int           `json:"max_generations"`

	// Rendering configuration
	DefaultViewMode   ViewMode      `json:"default_view_mode"`
	DefaultLayoutMode LayoutMode    `json:"default_layout_mode"`
	NodeWidth         int           `json:"node_width"`
	NodeHeight        int           `json:"node_height"`
	HorizontalSpacing int           `json:"horizontal_spacing"`
	VerticalSpacing   int           `json:"vertical_spacing"`

	// Performance configuration
	EnableCaching     bool          `json:"enable_caching"`
	LazyLoading       bool          `json:"lazy_loading"`
	BackgroundRefresh bool          `json:"background_refresh"`
	RefreshInterval   time.Duration `json:"refresh_interval"`

	// Data retention
	RelationshipTTL   time.Duration `json:"relationship_ttl"`
	TreeCacheTTL      time.Duration `json:"tree_cache_ttl"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// DefaultGenealogyConfig returns sensible defaults
func DefaultGenealogyConfig() GenealogyConfig {
	return GenealogyConfig{
		RedisKeyPrefix:    "queue:genealogy",
		CacheTTL:          5 * time.Minute,
		MaxTreeSize:       10000,
		MaxGenerations:    50,
		DefaultViewMode:   ViewModeFull,
		DefaultLayoutMode: LayoutModeTopDown,
		NodeWidth:         30,
		NodeHeight:        3,
		HorizontalSpacing: 2,
		VerticalSpacing:   1,
		EnableCaching:     true,
		LazyLoading:       true,
		BackgroundRefresh: true,
		RefreshInterval:   30 * time.Second,
		RelationshipTTL:   24 * time.Hour,
		TreeCacheTTL:      5 * time.Minute,
		CleanupInterval:   1 * time.Hour,
	}
}

// GraphStore defines the interface for storing and retrieving genealogy data
type GraphStore interface {
	// Relationship management
	AddRelationship(ctx context.Context, rel JobRelationship) error
	GetRelationships(ctx context.Context, jobID string) ([]JobRelationship, error)
	GetParents(ctx context.Context, jobID string) ([]string, error)
	GetChildren(ctx context.Context, jobID string) ([]string, error)

	// Tree operations
	GetAncestors(ctx context.Context, jobID string) ([]string, error)
	GetDescendants(ctx context.Context, jobID string) ([]string, error)
	BuildGenealogy(ctx context.Context, rootID string) (*JobGenealogy, error)

	// Cleanup
	RemoveRelationships(ctx context.Context, jobID string) error
	Cleanup(ctx context.Context, olderThan time.Time) error
}

// JobProvider defines the interface for retrieving job information
type JobProvider interface {
	GetJob(ctx context.Context, jobID string) (*JobNode, error)
	GetJobs(ctx context.Context, jobIDs []string) (map[string]*JobNode, error)
	SearchJobs(ctx context.Context, filter JobFilter) ([]*JobNode, error)
}

// JobFilter defines criteria for filtering jobs
type JobFilter struct {
	Status    []JobStatus `json:"status,omitempty"`
	Queues    []string    `json:"queues,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	NamePattern   string     `json:"name_pattern,omitempty"`
	Limit         int        `json:"limit,omitempty"`
}

// TreeRenderer defines the interface for rendering genealogy trees
type TreeRenderer interface {
	RenderTree(genealogy *JobGenealogy, layout *TreeLayout, state *NavigationState) ([]string, error)
	ComputeLayout(genealogy *JobGenealogy, mode LayoutMode, viewport Viewport) (*TreeLayout, error)
	GetViewport() Viewport
	SetViewport(viewport Viewport)
}

// Viewport defines the rendering viewport
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	OffsetX int `json:"offset_x"`
	OffsetY int `json:"offset_y"`
}

// NavigationEvent represents user navigation events
type NavigationEvent struct {
	Type      string    `json:"type"`
	JobID     string    `json:"job_id,omitempty"`
	Direction string    `json:"direction,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// GenealogyNavigator is the main interface for the job genealogy system
type GenealogyNavigator interface {
	// Core operations
	GetGenealogy(ctx context.Context, jobID string) (*JobGenealogy, error)
	GetImpactAnalysis(ctx context.Context, jobID string) (*ImpactAnalysis, error)
	GetBlameAnalysis(ctx context.Context, failedJobID string) (*BlameAnalysis, error)

	// Tree operations
	ExpandNode(ctx context.Context, jobID string) error
	CollapseNode(ctx context.Context, jobID string) error
	FocusOnJob(ctx context.Context, jobID string) error

	// Navigation
	SetViewMode(mode ViewMode) error
	SetLayoutMode(mode LayoutMode) error
	Navigate(direction string) error
	Search(term string) ([]*JobNode, error)

	// Rendering
	RenderCurrent() ([]string, error)
	RefreshTree() error

	// State management
	GetNavigationState() *NavigationState
	SetNavigationState(state *NavigationState) error

	// Events
	Subscribe(callback func(NavigationEvent)) error
	Unsubscribe() error

	// Lifecycle
	Start(ctx context.Context) error
	Stop() error
}

// GenealogyCache provides caching for genealogy data
type GenealogyCache struct {
	trees         map[string]*JobGenealogy
	layouts       map[string]*TreeLayout
	relationships map[string][]JobRelationship
	mu            sync.RWMutex
	maxSize       int
	ttl           time.Duration
}

// NewGenealogyCache creates a new genealogy cache
func NewGenealogyCache(maxSize int, ttl time.Duration) *GenealogyCache {
	return &GenealogyCache{
		trees:         make(map[string]*JobGenealogy),
		layouts:       make(map[string]*TreeLayout),
		relationships: make(map[string][]JobRelationship),
		maxSize:       maxSize,
		ttl:           ttl,
	}
}

// PatternDetector analyzes genealogy patterns
type PatternDetector interface {
	AnalyzeFailurePatterns(ctx context.Context, genealogies []*JobGenealogy) ([]FailurePattern, error)
	DetectAnomalies(ctx context.Context, genealogy *JobGenealogy) ([]GenealogyAnomaly, error)
	FindSimilarGenealogies(ctx context.Context, targetID string, limit int) ([]*JobGenealogy, error)
}

// FailurePattern represents a recurring failure pattern
type FailurePattern struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Frequency      int       `json:"frequency"`
	LastSeen       time.Time `json:"last_seen"`
	JobTypes       []string  `json:"job_types"`
	FailureReasons []string  `json:"failure_reasons"`
	Signature      string    `json:"signature"`
}

// GenealogyAnomaly represents an unusual pattern in job genealogy
type GenealogyAnomaly struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	JobID       string    `json:"job_id"`
	DetectedAt  time.Time `json:"detected_at"`
	Metrics     map[string]float64 `json:"metrics"`
}

// TreeMetrics provides performance and usage metrics
type TreeMetrics struct {
	TreeSize           int           `json:"tree_size"`
	MaxDepth           int           `json:"max_depth"`
	RenderTime         time.Duration `json:"render_time"`
	LayoutTime         time.Duration `json:"layout_time"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	NavigationEvents   int           `json:"navigation_events"`
	ExpandOperations   int           `json:"expand_operations"`
	CollapseOperations int           `json:"collapse_operations"`
}

// ColorScheme defines colors for different job statuses
type ColorScheme struct {
	Success    string `json:"success"`
	Failed     string `json:"failed"`
	Processing string `json:"processing"`
	Pending    string `json:"pending"`
	Retry      string `json:"retry"`
	Cancelled  string `json:"cancelled"`
	Background string `json:"background"`
	Border     string `json:"border"`
	Text       string `json:"text"`
	Highlight  string `json:"highlight"`
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Success:    "green",
		Failed:     "red",
		Processing: "yellow",
		Pending:    "gray",
		Retry:      "orange",
		Cancelled:  "purple",
		Background: "black",
		Border:     "white",
		Text:       "white",
		Highlight:  "cyan",
	}
}