package calendarview

import (
	"time"
)

// DefaultConfig returns the default calendar configuration
func DefaultConfig() *CalendarConfig {
	return &CalendarConfig{
		DefaultView:     ViewTypeMonth,
		DefaultTimezone: "UTC",
		MaxEventsPerCell: 100,
		RefreshInterval: 30 * time.Second,
		KeyBindings:     DefaultKeyBindings(),
		Colors:          DefaultColorScheme(),
	}
}

// DefaultKeyBindings returns the default keyboard shortcuts
func DefaultKeyBindings() []KeyBinding {
	return []KeyBinding{
		{Key: "h", Action: NavLeft, Description: "Move left"},
		{Key: "j", Action: NavDown, Description: "Move down"},
		{Key: "k", Action: NavUp, Description: "Move up"},
		{Key: "l", Action: NavRight, Description: "Move right"},
		{Key: "Left", Action: NavLeft, Description: "Move left"},
		{Key: "Down", Action: NavDown, Description: "Move down"},
		{Key: "Up", Action: NavUp, Description: "Move up"},
		{Key: "Right", Action: NavRight, Description: "Move right"},
		{Key: "[", Action: NavPrevPeriod, Description: "Previous period"},
		{Key: "]", Action: NavNextPeriod, Description: "Next period"},
		{Key: "g", Action: NavToday, Description: "Go to today"},
		{Key: "G", Action: NavEnd, Description: "Go to end"},
		{Key: "Enter", Action: NavInspect, Description: "Inspect selected"},
		{Key: "r", Action: NavReschedule, Description: "Reschedule event"},
		{Key: "p", Action: NavPause, Description: "Pause/Resume rule"},
		{Key: "/", Action: NavFilter, Description: "Filter events"},
		{Key: " ", Action: NavSelect, Description: "Select/Deselect"},
	}
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Background:      "#1e1e1e",
		Border:          "#444444",
		Text:            "#ffffff",
		Today:          "#ffff00",
		Selected:       "#0078d4",
		LowDensity:     "#2d5016",
		MediumDensity:  "#39a845",
		HighDensity:    "#57d364",
		StatusRunning:  "#ffa500",
		StatusFailed:   "#ff0000",
		StatusCompleted: "#00ff00",
	}
}

// ViewConfig represents configuration for a specific view type
type ViewConfig struct {
	ViewType        ViewType      `json:"view_type"`
	WeeksToShow     int           `json:"weeks_to_show"`     // For month view
	DaysToShow      int           `json:"days_to_show"`      // For week view
	HoursToShow     int           `json:"hours_to_show"`     // For day view
	StartHour       int           `json:"start_hour"`        // Day view start hour
	EndHour         int           `json:"end_hour"`          // Day view end hour
	ShowWeekends    bool          `json:"show_weekends"`     // Include weekends
	ShowWeekNumbers bool          `json:"show_week_numbers"` // Show week numbers
	DensityLevels   []DensityLevel `json:"density_levels"`   // Density thresholds
}

// DensityLevel represents a density threshold for visual styling
type DensityLevel struct {
	Threshold int    `json:"threshold"`
	Color     string `json:"color"`
	Label     string `json:"label"`
}

// GetViewConfig returns configuration for a specific view type
func GetViewConfig(viewType ViewType) *ViewConfig {
	switch viewType {
	case ViewTypeMonth:
		return &ViewConfig{
			ViewType:        ViewTypeMonth,
			WeeksToShow:     6,
			ShowWeekends:    true,
			ShowWeekNumbers: true,
			DensityLevels: []DensityLevel{
				{Threshold: 0, Color: "#2d5016", Label: "Low"},
				{Threshold: 5, Color: "#39a845", Label: "Medium"},
				{Threshold: 15, Color: "#57d364", Label: "High"},
				{Threshold: 30, Color: "#7dd87d", Label: "Very High"},
			},
		}
	case ViewTypeWeek:
		return &ViewConfig{
			ViewType:        ViewTypeWeek,
			DaysToShow:      7,
			ShowWeekends:    true,
			ShowWeekNumbers: false,
			DensityLevels: []DensityLevel{
				{Threshold: 0, Color: "#2d5016", Label: "Low"},
				{Threshold: 3, Color: "#39a845", Label: "Medium"},
				{Threshold: 8, Color: "#57d364", Label: "High"},
				{Threshold: 15, Color: "#7dd87d", Label: "Very High"},
			},
		}
	case ViewTypeDay:
		return &ViewConfig{
			ViewType:        ViewTypeDay,
			HoursToShow:     24,
			StartHour:       0,
			EndHour:         23,
			ShowWeekends:    true,
			ShowWeekNumbers: false,
			DensityLevels: []DensityLevel{
				{Threshold: 0, Color: "#2d5016", Label: "Low"},
				{Threshold: 1, Color: "#39a845", Label: "Medium"},
				{Threshold: 3, Color: "#57d364", Label: "High"},
				{Threshold: 6, Color: "#7dd87d", Label: "Very High"},
			},
		}
	default:
		return GetViewConfig(ViewTypeMonth)
	}
}

// RenderConfig represents configuration for calendar rendering
type RenderConfig struct {
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	CellWidth       int    `json:"cell_width"`
	CellHeight      int    `json:"cell_height"`
	ShowHeader      bool   `json:"show_header"`
	ShowFooter      bool   `json:"show_footer"`
	ShowLegend      bool   `json:"show_legend"`
	ShowTimezone    bool   `json:"show_timezone"`
	CompactMode     bool   `json:"compact_mode"`
	UseUnicode      bool   `json:"use_unicode"`
	BorderStyle     string `json:"border_style"`
	DateFormat      string `json:"date_format"`
	TimeFormat      string `json:"time_format"`
	MaxTooltipLines int    `json:"max_tooltip_lines"`
}

// DefaultRenderConfig returns the default render configuration
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		Width:           80,
		Height:          24,
		CellWidth:       10,
		CellHeight:      3,
		ShowHeader:      true,
		ShowFooter:      true,
		ShowLegend:      true,
		ShowTimezone:    true,
		CompactMode:     false,
		UseUnicode:      true,
		BorderStyle:     "rounded",
		DateFormat:      "2006-01-02",
		TimeFormat:      "15:04",
		MaxTooltipLines: 10,
	}
}

// TimezoneConfig represents timezone-related configuration
type TimezoneConfig struct {
	DefaultTimezone string   `json:"default_timezone"`
	AllowedTimezones []string `json:"allowed_timezones"`
	ShowUTCOffset   bool     `json:"show_utc_offset"`
	AutoDetect      bool     `json:"auto_detect"`
}

// DefaultTimezoneConfig returns the default timezone configuration
func DefaultTimezoneConfig() *TimezoneConfig {
	return &TimezoneConfig{
		DefaultTimezone: "UTC",
		AllowedTimezones: []string{
			"UTC",
			"America/New_York",
			"America/Chicago",
			"America/Denver",
			"America/Los_Angeles",
			"Europe/London",
			"Europe/Paris",
			"Europe/Berlin",
			"Asia/Tokyo",
			"Asia/Shanghai",
			"Asia/Kolkata",
			"Australia/Sydney",
		},
		ShowUTCOffset: true,
		AutoDetect:    true,
	}
}

// FilterConfig represents configuration for event filtering
type FilterConfig struct {
	MaxQueueNames     int      `json:"max_queue_names"`
	MaxJobTypes       int      `json:"max_job_types"`
	MaxTags           int      `json:"max_tags"`
	MaxSearchLength   int      `json:"max_search_length"`
	CaseSensitive     bool     `json:"case_sensitive"`
	AllowWildcards    bool     `json:"allow_wildcards"`
	DefaultStatuses   []string `json:"default_statuses"`
	SavedFilters      []string `json:"saved_filters"`
	RecentFilters     []string `json:"recent_filters"`
	MaxRecentFilters  int      `json:"max_recent_filters"`
}

// DefaultFilterConfig returns the default filter configuration
func DefaultFilterConfig() *FilterConfig {
	return &FilterConfig{
		MaxQueueNames:    10,
		MaxJobTypes:      20,
		MaxTags:          15,
		MaxSearchLength:  100,
		CaseSensitive:    false,
		AllowWildcards:   true,
		DefaultStatuses:  []string{"scheduled", "running"},
		SavedFilters:     []string{},
		RecentFilters:    []string{},
		MaxRecentFilters: 10,
	}
}

// PerformanceConfig represents performance-related configuration
type PerformanceConfig struct {
	MaxEventsPerQuery   int           `json:"max_events_per_query"`
	CacheTimeout        time.Duration `json:"cache_timeout"`
	RefreshInterval     time.Duration `json:"refresh_interval"`
	BackgroundRefresh   bool          `json:"background_refresh"`
	LazyLoading         bool          `json:"lazy_loading"`
	PreloadDays         int           `json:"preload_days"`
	MaxConcurrentReqs   int           `json:"max_concurrent_requests"`
	QueryTimeout        time.Duration `json:"query_timeout"`
	DebounceInterval    time.Duration `json:"debounce_interval"`
}

// DefaultPerformanceConfig returns the default performance configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		MaxEventsPerQuery: 1000,
		CacheTimeout:      5 * time.Minute,
		RefreshInterval:   30 * time.Second,
		BackgroundRefresh: true,
		LazyLoading:       true,
		PreloadDays:       7,
		MaxConcurrentReqs: 5,
		QueryTimeout:      10 * time.Second,
		DebounceInterval:  250 * time.Millisecond,
	}
}

// SecurityConfig represents security-related configuration
type SecurityConfig struct {
	RequireAuth        bool     `json:"require_auth"`
	AllowedUsers       []string `json:"allowed_users"`
	AllowedRoles       []string `json:"allowed_roles"`
	AuditActions       bool     `json:"audit_actions"`
	RateLimitRequests  int      `json:"rate_limit_requests"`
	RateLimitWindow    string   `json:"rate_limit_window"`
	RequireConfirm     bool     `json:"require_confirmation"`
	MaxRescheduleRange string   `json:"max_reschedule_range"`
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequireAuth:        true,
		AllowedUsers:       []string{},
		AllowedRoles:       []string{"admin", "operator"},
		AuditActions:       true,
		RateLimitRequests:  100,
		RateLimitWindow:    "1h",
		RequireConfirm:     true,
		MaxRescheduleRange: "90d",
	}
}