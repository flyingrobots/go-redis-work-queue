package context_menus

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		handlers: make(map[string]ActionHandler),
		actions:  make(map[ContextType][]MenuAction),
	}
}

// RegisterHandler registers an action handler
func (ar *ActionRegistry) RegisterHandler(actionID string, handler ActionHandler) {
	ar.handlers[actionID] = handler
}

// RegisterAction registers an action for a specific context type
func (ar *ActionRegistry) RegisterAction(contextType ContextType, action MenuAction) {
	if ar.actions[contextType] == nil {
		ar.actions[contextType] = make([]MenuAction, 0)
	}
	ar.actions[contextType] = append(ar.actions[contextType], action)
}

// GetActions returns available actions for a context
func (ar *ActionRegistry) GetActions(ctx MenuContext) []MenuAction {
	actions := ar.actions[ctx.Type]
	if actions == nil {
		return []MenuAction{}
	}

	// Filter actions based on context-specific capabilities
	filtered := make([]MenuAction, 0, len(actions))
	for _, action := range actions {
		if ar.isActionAvailable(action, ctx) {
			filtered = append(filtered, action)
		}
	}

	return filtered
}

// ExecuteAction executes an action with the given context
func (ar *ActionRegistry) ExecuteAction(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	handler, exists := ar.handlers[action.ID]
	if !exists {
		return func() tea.Msg {
			return MenuActionMsg{
				Action:  action,
				Context: menuCtx,
			}
		}
	}

	return handler(ctx, action, menuCtx)
}

// isActionAvailable checks if an action is available for the given context
func (ar *ActionRegistry) isActionAvailable(action MenuAction, ctx MenuContext) bool {
	if action.Disabled {
		return false
	}

	// Context-specific availability checks
	switch ctx.Type {
	case ContextQueueRow:
		return ar.isQueueActionAvailable(action, ctx)
	case ContextDLQItem:
		return ar.isDLQActionAvailable(action, ctx)
	case ContextTab:
		return ar.isTabActionAvailable(action, ctx)
	case ContextChart:
		return ar.isChartActionAvailable(action, ctx)
	case ContextInfoRegion:
		return ar.isInfoActionAvailable(action, ctx)
	default:
		return true
	}
}

// isQueueActionAvailable checks queue-specific action availability
func (ar *ActionRegistry) isQueueActionAvailable(action MenuAction, ctx MenuContext) bool {
	switch action.ID {
	case "peek":
		return ctx.QueueName != ""
	case "enqueue":
		return ctx.QueueName != ""
	case "purge":
		return ctx.QueueName != "" && action.Destructive
	case "copy_queue_name":
		return ctx.QueueName != ""
	case "export_sample":
		return ctx.QueueName != ""
	default:
		return true
	}
}

// isDLQActionAvailable checks DLQ-specific action availability
func (ar *ActionRegistry) isDLQActionAvailable(action MenuAction, ctx MenuContext) bool {
	switch action.ID {
	case "requeue":
		return ctx.JobID != ""
	case "purge_dlq":
		return action.Destructive
	case "copy_job_id":
		return ctx.JobID != ""
	case "copy_payload":
		return ctx.JobID != ""
	case "open_trace":
		return ctx.JobID != ""
	default:
		return true
	}
}

// isTabActionAvailable checks tab-specific action availability
func (ar *ActionRegistry) isTabActionAvailable(action MenuAction, ctx MenuContext) bool {
	switch action.ID {
	case "close_tab":
		return ctx.ItemID != "" && ctx.ItemID != "jobs" // Can't close main tab
	case "duplicate_tab":
		return ctx.ItemID != ""
	default:
		return true
	}
}

// isChartActionAvailable checks chart-specific action availability
func (ar *ActionRegistry) isChartActionAvailable(action MenuAction, ctx MenuContext) bool {
	switch action.ID {
	case "export_chart":
		return ctx.ItemID != ""
	case "configure_chart":
		return ctx.ItemID != ""
	case "reset_zoom":
		return ctx.ItemID != ""
	default:
		return true
	}
}

// isInfoActionAvailable checks info region-specific action availability
func (ar *ActionRegistry) isInfoActionAvailable(action MenuAction, ctx MenuContext) bool {
	switch action.ID {
	case "copy_info":
		return ctx.ItemID != ""
	case "export_info":
		return ctx.ItemID != ""
	default:
		return true
	}
}

// InitializeDefaultActions sets up the default action registry
func (ar *ActionRegistry) InitializeDefaultActions() {
	// Queue row actions
	ar.RegisterAction(ContextQueueRow, MenuAction{
		ID:          "peek",
		Label:       "Peek Jobs",
		Accelerator: "p",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextQueueRow, MenuAction{
		ID:          "enqueue",
		Label:       "Enqueue Job",
		Accelerator: "e",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextQueueRow, MenuAction{
		ID:          "purge",
		Label:       "Purge Queue",
		Accelerator: "P",
		Destructive: true,
		Confirm:     true,
		ConfirmText: "Are you sure you want to purge this queue? This action cannot be undone.",
	})

	ar.RegisterAction(ContextQueueRow, MenuAction{
		ID:          "copy_queue_name",
		Label:       "Copy Queue Name",
		Accelerator: "c",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextQueueRow, MenuAction{
		ID:          "export_sample",
		Label:       "Export Sample",
		Accelerator: "x",
		Destructive: false,
		Confirm:     false,
	})

	// DLQ item actions
	ar.RegisterAction(ContextDLQItem, MenuAction{
		ID:          "requeue",
		Label:       "Requeue Job",
		Accelerator: "r",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextDLQItem, MenuAction{
		ID:          "purge_dlq",
		Label:       "Purge from DLQ",
		Accelerator: "P",
		Destructive: true,
		Confirm:     true,
		ConfirmText: "Are you sure you want to permanently delete this job from the DLQ?",
	})

	ar.RegisterAction(ContextDLQItem, MenuAction{
		ID:          "copy_job_id",
		Label:       "Copy Job ID",
		Accelerator: "i",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextDLQItem, MenuAction{
		ID:          "copy_payload",
		Label:       "Copy Payload",
		Accelerator: "c",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextDLQItem, MenuAction{
		ID:          "open_trace",
		Label:       "Open Trace",
		Accelerator: "t",
		Destructive: false,
		Confirm:     false,
	})

	// Tab actions
	ar.RegisterAction(ContextTab, MenuAction{
		ID:          "close_tab",
		Label:       "Close Tab",
		Accelerator: "w",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextTab, MenuAction{
		ID:          "duplicate_tab",
		Label:       "Duplicate Tab",
		Accelerator: "d",
		Destructive: false,
		Confirm:     false,
	})

	// Chart actions
	ar.RegisterAction(ContextChart, MenuAction{
		ID:          "export_chart",
		Label:       "Export Chart",
		Accelerator: "x",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextChart, MenuAction{
		ID:          "configure_chart",
		Label:       "Configure Chart",
		Accelerator: "c",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextChart, MenuAction{
		ID:          "reset_zoom",
		Label:       "Reset Zoom",
		Accelerator: "z",
		Destructive: false,
		Confirm:     false,
	})

	// Info region actions
	ar.RegisterAction(ContextInfoRegion, MenuAction{
		ID:          "copy_info",
		Label:       "Copy Information",
		Accelerator: "c",
		Destructive: false,
		Confirm:     false,
	})

	ar.RegisterAction(ContextInfoRegion, MenuAction{
		ID:          "export_info",
		Label:       "Export Information",
		Accelerator: "x",
		Destructive: false,
		Confirm:     false,
	})
}

// GetActionByID returns an action by its ID for a given context type
func (ar *ActionRegistry) GetActionByID(contextType ContextType, actionID string) (MenuAction, error) {
	actions := ar.actions[contextType]
	for _, action := range actions {
		if action.ID == actionID {
			return action, nil
		}
	}
	return MenuAction{}, &ContextMenuError{
		Message: fmt.Sprintf("action %s not found for context type %d", actionID, int(contextType)),
		Code:    "ACTION_NOT_FOUND",
	}
}