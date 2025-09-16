package context_menus

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ContextMenuSystem is the main interface for the right-click context menu system
type ContextMenuSystem struct {
	menu *ContextMenu
}

// New creates a new context menu system
func New() *ContextMenuSystem {
	menu := NewContextMenu()

	// Initialize default actions and handlers
	menu.registry.InitializeDefaultActions()
	InitializeDefaultHandlers(menu.GetRegistry())

	return &ContextMenuSystem{
		menu: menu,
	}
}

// Init initializes the context menu system
func (cms *ContextMenuSystem) Init() tea.Cmd {
	return cms.menu.Init()
}

// Update handles Bubble Tea messages
func (cms *ContextMenuSystem) Update(msg tea.Msg) (*ContextMenuSystem, tea.Cmd) {
	var cmd tea.Cmd
	cms.menu, cmd = cms.menu.Update(msg)
	return cms, cmd
}

// View renders the context menu if visible
func (cms *ContextMenuSystem) View() string {
	return cms.menu.View()
}

// ShowMenu displays a context menu for the given context
func (cms *ContextMenuSystem) ShowMenu(ctx MenuContext) tea.Cmd {
	return func() tea.Msg {
		return ShowMenuMsg{Context: ctx}
	}
}

// HideMenu hides the context menu
func (cms *ContextMenuSystem) HideMenu() tea.Cmd {
	return func() tea.Msg {
		return HideMenuMsg{}
	}
}

// RegisterZone registers a clickable zone for context menu interaction
func (cms *ContextMenuSystem) RegisterZone(zone BubbleZone) error {
	return cms.menu.GetZoneManager().RegisterZone(zone)
}

// UnregisterZone removes a zone from the manager
func (cms *ContextMenuSystem) UnregisterZone(id string) {
	cms.menu.GetZoneManager().UnregisterZone(id)
}

// RegisterTableRow is a convenience method for registering table row zones
func (cms *ContextMenuSystem) RegisterTableRow(rowIndex int, x, y, width int, queueName string) error {
	return cms.menu.GetZoneManager().RegisterTableRow(rowIndex, x, y, width, queueName)
}

// RegisterTab is a convenience method for registering tab zones
func (cms *ContextMenuSystem) RegisterTab(tabID, label string, x, y, width int) error {
	return cms.menu.GetZoneManager().RegisterTab(tabID, label, x, y, width)
}

// RegisterChart is a convenience method for registering chart zones
func (cms *ContextMenuSystem) RegisterChart(chartID string, x, y, width, height int) error {
	return cms.menu.GetZoneManager().RegisterChart(chartID, x, y, width, height)
}

// RegisterInfoRegion is a convenience method for registering info region zones
func (cms *ContextMenuSystem) RegisterInfoRegion(regionID string, x, y, width, height int) error {
	return cms.menu.GetZoneManager().RegisterInfoRegion(regionID, x, y, width, height)
}

// RegisterDLQItem is a convenience method for registering DLQ item zones
func (cms *ContextMenuSystem) RegisterDLQItem(itemIndex int, jobID string, x, y, width int) error {
	return cms.menu.GetZoneManager().RegisterDLQItem(itemIndex, jobID, x, y, width)
}

// ClearZones removes all registered zones
func (cms *ContextMenuSystem) ClearZones() {
	cms.menu.GetZoneManager().ClearZones()
}

// GetZoneAt returns the zone at the given coordinates
func (cms *ContextMenuSystem) GetZoneAt(x, y int) (BubbleZone, bool) {
	return cms.menu.GetZoneManager().GetZoneAt(x, y)
}

// RegisterAction registers a custom action for a context type
func (cms *ContextMenuSystem) RegisterAction(contextType ContextType, action MenuAction) {
	cms.menu.GetRegistry().RegisterAction(contextType, action)
}

// RegisterHandler registers a custom action handler
func (cms *ContextMenuSystem) RegisterHandler(actionID string, handler ActionHandler) {
	cms.menu.GetRegistry().RegisterHandler(actionID, handler)
}

// IsVisible returns whether the context menu is currently visible
func (cms *ContextMenuSystem) IsVisible() bool {
	return cms.menu.IsVisible()
}

// GetActions returns available actions for a given context
func (cms *ContextMenuSystem) GetActions(ctx MenuContext) []MenuAction {
	return cms.menu.GetRegistry().GetActions(ctx)
}

// SetEnabled enables or disables the entire context menu system
func (cms *ContextMenuSystem) SetEnabled(enabled bool) {
	cms.menu.GetZoneManager().SetEnabled(enabled)
}

// IsEnabled returns whether the context menu system is enabled
func (cms *ContextMenuSystem) IsEnabled() bool {
	return cms.menu.GetZoneManager().IsEnabled()
}