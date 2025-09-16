package context_menus

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewContextMenu creates a new context menu system
func NewContextMenu() *ContextMenu {
	return &ContextMenu{
		state: MenuState{
			Visible:       false,
			SelectedIndex: 0,
		},
		registry: NewActionRegistry(),
		zones:    NewZoneManager(),
	}
}

// Init initializes the context menu with default actions
func (cm *ContextMenu) Init() tea.Cmd {
	return nil
}

// Update handles Bubble Tea messages
func (cm *ContextMenu) Update(msg tea.Msg) (*ContextMenu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cm.width = msg.Width
		cm.height = msg.Height

	case tea.KeyMsg:
		return cm.handleKeyMsg(msg)

	case tea.MouseMsg:
		return cm.handleMouseMsg(msg)

	case ShowMenuMsg:
		return cm.showMenu(msg.Context)

	case HideMenuMsg:
		cm.hideMenu()

	case MenuActionMsg:
		return cm.executeAction(msg.Action, msg.Context)
	}

	return cm, nil
}

// handleKeyMsg processes keyboard input
func (cm *ContextMenu) handleKeyMsg(msg tea.KeyMsg) (*ContextMenu, tea.Cmd) {
	if !cm.state.Visible {
		// 'm' key opens context menu for focused item
		if msg.String() == "m" {
			// TODO: Get focused item context from parent model
			return cm, nil
		}
		return cm, nil
	}

	switch msg.String() {
	case "esc":
		cm.hideMenu()
		return cm, nil

	case "up", "k":
		cm.moveCursor(-1)
		return cm, nil

	case "down", "j":
		cm.moveCursor(1)
		return cm, nil

	case "enter", " ":
		if len(cm.state.Actions) > 0 && cm.state.SelectedIndex < len(cm.state.Actions) {
			action := cm.state.Actions[cm.state.SelectedIndex]
			return cm.executeAction(action, cm.state.Context)
		}
		return cm, nil

	default:
		// Check for accelerator keys
		for i, action := range cm.state.Actions {
			if strings.EqualFold(msg.String(), action.Accelerator) {
				cm.state.SelectedIndex = i
				return cm.executeAction(action, cm.state.Context)
			}
		}
	}

	return cm, nil
}

// handleMouseMsg processes mouse input
func (cm *ContextMenu) handleMouseMsg(msg tea.MouseMsg) (*ContextMenu, tea.Cmd) {
	if msg.Type == tea.MouseLeft || msg.Type == tea.MouseRight {
		// Check if click is on an existing menu
		if cm.state.Visible && cm.isClickOnMenu(msg.X, msg.Y) {
			if msg.Type == tea.MouseLeft {
				// Select menu item and execute
				menuY := cm.state.Position.Y + 1 // Account for border
				itemIndex := msg.Y - menuY
				if itemIndex >= 0 && itemIndex < len(cm.state.Actions) {
					action := cm.state.Actions[itemIndex]
					return cm.executeAction(action, cm.state.Context)
				}
			}
			return cm, nil
		}

		// Hide menu if clicking outside
		if cm.state.Visible {
			cm.hideMenu()
			return cm, nil
		}

		// Right-click to show context menu
		if msg.Type == tea.MouseRight {
			zone, found := cm.zones.GetZoneAt(msg.X, msg.Y)
			if found {
				position := Position{X: msg.X, Y: msg.Y}
				zone.Context.Position = position
				return cm.showMenu(zone.Context)
			}
		}
	}

	return cm, nil
}

// showMenu displays the context menu
func (cm *ContextMenu) showMenu(ctx MenuContext) (*ContextMenu, tea.Cmd) {
	actions := cm.registry.GetActions(ctx)
	if len(actions) == 0 {
		return cm, nil
	}

	// Adjust position to keep menu on screen
	position := cm.adjustMenuPosition(ctx.Position, len(actions))

	cm.state = MenuState{
		Visible:       true,
		Context:       ctx,
		Actions:       actions,
		SelectedIndex: 0,
		Position:      position,
	}

	return cm, nil
}

// hideMenu hides the context menu
func (cm *ContextMenu) hideMenu() {
	cm.state.Visible = false
}

// executeAction executes a menu action
func (cm *ContextMenu) executeAction(action MenuAction, ctx MenuContext) (*ContextMenu, tea.Cmd) {
	cm.hideMenu()

	if action.Confirm && action.Destructive {
		// Show confirmation dialog
		return cm, func() tea.Msg {
			return ShowConfirmationMsg{
				Action:  action,
				Context: ctx,
				Message: action.ConfirmText,
			}
		}
	}

	// Execute action directly
	return cm, cm.registry.ExecuteAction(context.Background(), action, ctx)
}

// moveCursor moves the menu cursor
func (cm *ContextMenu) moveCursor(delta int) {
	if len(cm.state.Actions) == 0 {
		return
	}

	cm.state.SelectedIndex += delta
	if cm.state.SelectedIndex < 0 {
		cm.state.SelectedIndex = len(cm.state.Actions) - 1
	} else if cm.state.SelectedIndex >= len(cm.state.Actions) {
		cm.state.SelectedIndex = 0
	}
}

// isClickOnMenu checks if a click is within the menu bounds
func (cm *ContextMenu) isClickOnMenu(x, y int) bool {
	if !cm.state.Visible {
		return false
	}

	menuWidth := cm.getMenuWidth()
	menuHeight := len(cm.state.Actions) + 2 // +2 for borders

	return x >= cm.state.Position.X &&
		x < cm.state.Position.X+menuWidth &&
		y >= cm.state.Position.Y &&
		y < cm.state.Position.Y+menuHeight
}

// adjustMenuPosition adjusts menu position to keep it on screen
func (cm *ContextMenu) adjustMenuPosition(pos Position, itemCount int) Position {
	menuWidth := cm.getMenuWidth()
	menuHeight := itemCount + 2 // +2 for borders

	adjustedX := pos.X
	adjustedY := pos.Y

	// Keep menu within horizontal bounds
	if adjustedX+menuWidth > cm.width {
		adjustedX = cm.width - menuWidth
	}
	if adjustedX < 0 {
		adjustedX = 0
	}

	// Keep menu within vertical bounds
	if adjustedY+menuHeight > cm.height {
		adjustedY = cm.height - menuHeight
	}
	if adjustedY < 0 {
		adjustedY = 0
	}

	return Position{X: adjustedX, Y: adjustedY}
}

// getMenuWidth calculates the width needed for the menu
func (cm *ContextMenu) getMenuWidth() int {
	maxWidth := 20 // Minimum width
	for _, action := range cm.state.Actions {
		labelWidth := len(action.Label)
		if action.Accelerator != "" {
			labelWidth += len(action.Accelerator) + 3 // " (x)"
		}
		if labelWidth > maxWidth {
			maxWidth = labelWidth
		}
	}
	return maxWidth + 4 // +4 for borders and padding
}

// View renders the context menu
func (cm *ContextMenu) View() string {
	if !cm.state.Visible {
		return ""
	}

	return cm.renderMenu()
}

// renderMenu renders the menu with proper styling
func (cm *ContextMenu) renderMenu() string {
	if len(cm.state.Actions) == 0 {
		return ""
	}

	menuWidth := cm.getMenuWidth()

	// Menu styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true)

	destructiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	acceleratorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	// Build menu items
	var items []string
	for i, action := range cm.state.Actions {
		label := action.Label

		// Add accelerator if present
		if action.Accelerator != "" {
			label += " " + acceleratorStyle.Render(fmt.Sprintf("(%s)", action.Accelerator))
		}

		// Apply destructive styling
		if action.Destructive {
			label = destructiveStyle.Render(label)
		}

		// Apply selection styling
		if i == cm.state.SelectedIndex {
			label = selectedStyle.Render(fmt.Sprintf("%-*s", menuWidth-4, label))
		} else {
			label = fmt.Sprintf("%-*s", menuWidth-4, label)
		}

		items = append(items, label)
	}

	menu := strings.Join(items, "\n")
	return borderStyle.Render(menu)
}

// GetZoneManager returns the zone manager
func (cm *ContextMenu) GetZoneManager() *ZoneManager {
	return cm.zones
}

// GetRegistry returns the action registry
func (cm *ContextMenu) GetRegistry() *ActionRegistry {
	return cm.registry
}

// IsVisible returns whether the menu is currently visible
func (cm *ContextMenu) IsVisible() bool {
	return cm.state.Visible
}

// GetState returns the current menu state
func (cm *ContextMenu) GetState() MenuState {
	return cm.state
}

// Additional message types
type ShowConfirmationMsg struct {
	Action  MenuAction
	Context MenuContext
	Message string
}