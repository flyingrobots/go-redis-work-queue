package context_menus

import (
	"context"
	"github.com/charmbracelet/bubbletea"
)

// ContextType represents the type of context for menu actions
type ContextType int

const (
	ContextQueueRow ContextType = iota
	ContextDLQItem
	ContextTab
	ContextChart
	ContextInfoRegion
)

// MenuAction represents an available action in the context menu
type MenuAction struct {
	ID          string
	Label       string
	Accelerator string
	Destructive bool
	Disabled    bool
	Confirm     bool
	ConfirmText string
}

// MenuContext contains the context information for menu generation
type MenuContext struct {
	Type         ContextType
	ItemID       string
	QueueName    string
	JobID        string
	RowIndex     int
	Position     Position
	Metadata     map[string]interface{}
}

// Position represents screen coordinates
type Position struct {
	X int
	Y int
}

// MenuState represents the current menu state
type MenuState struct {
	Visible       bool
	Context       MenuContext
	Actions       []MenuAction
	SelectedIndex int
	Position      Position
}

// ActionHandler is a function that handles menu action execution
type ActionHandler func(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd

// ActionRegistry manages available actions for different contexts
type ActionRegistry struct {
	handlers map[string]ActionHandler
	actions  map[ContextType][]MenuAction
}

// BubbleZone represents a clickable area with associated context
type BubbleZone struct {
	ID       string
	X        int
	Y        int
	Width    int
	Height   int
	Context  MenuContext
	Enabled  bool
}

// ZoneManager manages all bubble zones for mouse interaction
type ZoneManager struct {
	zones   map[string]BubbleZone
	enabled bool
}

// ContextMenu manages the context menu system
type ContextMenu struct {
	state    MenuState
	registry *ActionRegistry
	zones    *ZoneManager
	width    int
	height   int
}

// Messages for Bubble Tea
type (
	ShowMenuMsg struct {
		Context MenuContext
	}

	HideMenuMsg struct{}

	MenuActionMsg struct {
		Action  MenuAction
		Context MenuContext
	}

	ZoneClickMsg struct {
		Zone     BubbleZone
		Position Position
		Button   int // 0=left, 1=right, 2=middle
	}
)

// Note: Error types are defined in errors.go