package context_menus

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Default action handlers for the context menu system

// PeekHandler handles the peek action
func PeekHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return PeekActionMsg{
			QueueName: menuCtx.QueueName,
			Context:   menuCtx,
		}
	}
}

// EnqueueHandler handles the enqueue action
func EnqueueHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return EnqueueActionMsg{
			QueueName: menuCtx.QueueName,
			Context:   menuCtx,
		}
	}
}

// PurgeHandler handles the purge action
func PurgeHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return PurgeActionMsg{
			QueueName: menuCtx.QueueName,
			Context:   menuCtx,
		}
	}
}

// CopyQueueNameHandler handles copying queue names
func CopyQueueNameHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return CopyActionMsg{
			Text:    menuCtx.QueueName,
			Type:    "queue_name",
			Context: menuCtx,
		}
	}
}

// ExportSampleHandler handles exporting sample data
func ExportSampleHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ExportActionMsg{
			QueueName: menuCtx.QueueName,
			Type:      "sample",
			Context:   menuCtx,
		}
	}
}

// RequeueHandler handles requeuing DLQ items
func RequeueHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return RequeueActionMsg{
			JobID:   menuCtx.JobID,
			Context: menuCtx,
		}
	}
}

// PurgeDLQHandler handles purging items from DLQ
func PurgeDLQHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return PurgeDLQActionMsg{
			JobID:   menuCtx.JobID,
			Context: menuCtx,
		}
	}
}

// CopyJobIDHandler handles copying job IDs
func CopyJobIDHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return CopyActionMsg{
			Text:    menuCtx.JobID,
			Type:    "job_id",
			Context: menuCtx,
		}
	}
}

// CopyPayloadHandler handles copying job payloads
func CopyPayloadHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return CopyPayloadActionMsg{
			JobID:   menuCtx.JobID,
			Context: menuCtx,
		}
	}
}

// OpenTraceHandler handles opening traces
func OpenTraceHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return OpenTraceActionMsg{
			JobID:   menuCtx.JobID,
			Context: menuCtx,
		}
	}
}

// CloseTabHandler handles closing tabs
func CloseTabHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return CloseTabActionMsg{
			TabID:   menuCtx.ItemID,
			Context: menuCtx,
		}
	}
}

// DuplicateTabHandler handles duplicating tabs
func DuplicateTabHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return DuplicateTabActionMsg{
			TabID:   menuCtx.ItemID,
			Context: menuCtx,
		}
	}
}

// ExportChartHandler handles exporting charts
func ExportChartHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ExportActionMsg{
			ChartID: menuCtx.ItemID,
			Type:    "chart",
			Context: menuCtx,
		}
	}
}

// ConfigureChartHandler handles chart configuration
func ConfigureChartHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ConfigureChartActionMsg{
			ChartID: menuCtx.ItemID,
			Context: menuCtx,
		}
	}
}

// ResetZoomHandler handles resetting chart zoom
func ResetZoomHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ResetZoomActionMsg{
			ChartID: menuCtx.ItemID,
			Context: menuCtx,
		}
	}
}

// CopyInfoHandler handles copying information
func CopyInfoHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return CopyInfoActionMsg{
			RegionID: menuCtx.ItemID,
			Context:  menuCtx,
		}
	}
}

// ExportInfoHandler handles exporting information
func ExportInfoHandler(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ExportActionMsg{
			RegionID: menuCtx.ItemID,
			Type:     "info",
			Context:  menuCtx,
		}
	}
}

// Action message types
type (
	PeekActionMsg struct {
		QueueName string
		Context   MenuContext
	}

	EnqueueActionMsg struct {
		QueueName string
		Context   MenuContext
	}

	PurgeActionMsg struct {
		QueueName string
		Context   MenuContext
	}

	CopyActionMsg struct {
		Text    string
		Type    string
		Context MenuContext
	}

	ExportActionMsg struct {
		QueueName string
		ChartID   string
		RegionID  string
		Type      string
		Context   MenuContext
	}

	RequeueActionMsg struct {
		JobID   string
		Context MenuContext
	}

	PurgeDLQActionMsg struct {
		JobID   string
		Context MenuContext
	}

	CopyPayloadActionMsg struct {
		JobID   string
		Context MenuContext
	}

	OpenTraceActionMsg struct {
		JobID   string
		Context MenuContext
	}

	CloseTabActionMsg struct {
		TabID   string
		Context MenuContext
	}

	DuplicateTabActionMsg struct {
		TabID   string
		Context MenuContext
	}

	ConfigureChartActionMsg struct {
		ChartID string
		Context MenuContext
	}

	ResetZoomActionMsg struct {
		ChartID string
		Context MenuContext
	}

	CopyInfoActionMsg struct {
		RegionID string
		Context  MenuContext
	}
)

// InitializeDefaultHandlers registers all default handlers with the registry
func InitializeDefaultHandlers(registry *ActionRegistry) {
	// Queue actions
	registry.RegisterHandler("peek", PeekHandler)
	registry.RegisterHandler("enqueue", EnqueueHandler)
	registry.RegisterHandler("purge", PurgeHandler)
	registry.RegisterHandler("copy_queue_name", CopyQueueNameHandler)
	registry.RegisterHandler("export_sample", ExportSampleHandler)

	// DLQ actions
	registry.RegisterHandler("requeue", RequeueHandler)
	registry.RegisterHandler("purge_dlq", PurgeDLQHandler)
	registry.RegisterHandler("copy_job_id", CopyJobIDHandler)
	registry.RegisterHandler("copy_payload", CopyPayloadHandler)
	registry.RegisterHandler("open_trace", OpenTraceHandler)

	// Tab actions
	registry.RegisterHandler("close_tab", CloseTabHandler)
	registry.RegisterHandler("duplicate_tab", DuplicateTabHandler)

	// Chart actions
	registry.RegisterHandler("export_chart", ExportChartHandler)
	registry.RegisterHandler("configure_chart", ConfigureChartHandler)
	registry.RegisterHandler("reset_zoom", ResetZoomHandler)

	// Info actions
	registry.RegisterHandler("copy_info", CopyInfoHandler)
	registry.RegisterHandler("export_info", ExportInfoHandler)
}

// ConfirmationHandler handles confirmation dialogs for destructive actions
func ConfirmationHandler(confirmed bool, action MenuAction, menuCtx MenuContext) tea.Cmd {
	if !confirmed {
		return nil
	}

	// Re-execute the action after confirmation
	switch action.ID {
	case "purge":
		return PurgeHandler(context.Background(), action, menuCtx)
	case "purge_dlq":
		return PurgeDLQHandler(context.Background(), action, menuCtx)
	default:
		return func() tea.Msg {
			return fmt.Sprintf("Unknown destructive action: %s", action.ID)
		}
	}
}