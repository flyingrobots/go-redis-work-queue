# Right-Click Context Menus API Documentation

## Overview

The Right-Click Context Menus module provides an interactive context menu system for the TUI (Terminal User Interface). It enables users to access contextual actions through right-click interactions and keyboard shortcuts, improving the overall user experience by making actions discoverable and easily accessible.

## Key Features

- **Precise Hitboxes**: Integration with bubblezone for accurate click detection
- **Context-Aware Menus**: Different menu options based on where the user clicks
- **Keyboard Navigation**: Full keyboard support with arrow keys and accelerators
- **Safety Confirmations**: Destructive actions require explicit confirmation
- **Extensible Actions**: Plugin-based action system for easy customization

## Core Components

### ContextMenuSystem

The main interface for the context menu system.

```go
type ContextMenuSystem struct {
    menu *ContextMenu
}

// Create a new context menu system
func New() *ContextMenuSystem

// Initialize the system
func (cms *ContextMenuSystem) Init() tea.Cmd

// Update with Bubble Tea messages
func (cms *ContextMenuSystem) Update(msg tea.Msg) (*ContextMenuSystem, tea.Cmd)

// Render the menu if visible
func (cms *ContextMenuSystem) View() string
```

### Context Types

The system supports different context types for targeted actions:

```go
type ContextType int

const (
    ContextQueueRow     // Table rows in queue view
    ContextDLQItem      // Dead letter queue items
    ContextTab          // Tab headers
    ContextChart        // Chart areas
    ContextInfoRegion   // Information regions
)
```

### Menu Actions

Actions are structured objects that define available operations:

```go
type MenuAction struct {
    ID          string  // Unique action identifier
    Label       string  // Display text
    Accelerator string  // Keyboard shortcut
    Destructive bool    // Requires confirmation
    Disabled    bool    // Currently unavailable
    Confirm     bool    // Show confirmation dialog
    ConfirmText string  // Confirmation message
}
```

## Zone Management

### Registering Zones

Register clickable areas for context menu interaction:

```go
// Register a table row
err := cms.RegisterTableRow(rowIndex, x, y, width, queueName)

// Register a tab
err := cms.RegisterTab(tabID, label, x, y, width)

// Register a chart area
err := cms.RegisterChart(chartID, x, y, width, height)

// Register a DLQ item
err := cms.RegisterDLQItem(itemIndex, jobID, x, y, width)

// Register an info region
err := cms.RegisterInfoRegion(regionID, x, y, width, height)
```

### Zone Properties

```go
type BubbleZone struct {
    ID       string      // Unique zone identifier
    X        int         // Left coordinate
    Y        int         // Top coordinate
    Width    int         // Zone width
    Height   int         // Zone height
    Context  MenuContext // Associated context
    Enabled  bool        // Zone is active
}
```

## Action Registry

### Default Actions

The system comes with pre-configured actions for each context type:

#### Queue Row Actions
- **Peek Jobs** (p): View jobs in the queue
- **Enqueue Job** (e): Add a new job to the queue
- **Purge Queue** (P): Delete all jobs (destructive, requires confirmation)
- **Copy Queue Name** (c): Copy queue name to clipboard
- **Export Sample** (x): Export sample data

#### DLQ Item Actions
- **Requeue Job** (r): Move job back to active queue
- **Purge from DLQ** (P): Permanently delete job (destructive)
- **Copy Job ID** (i): Copy job identifier
- **Copy Payload** (c): Copy job payload data
- **Open Trace** (t): Open distributed trace

#### Tab Actions
- **Close Tab** (w): Close the current tab
- **Duplicate Tab** (d): Create a copy of the tab

#### Chart Actions
- **Export Chart** (x): Export chart data
- **Configure Chart** (c): Open chart settings
- **Reset Zoom** (z): Reset chart zoom level

### Custom Actions

Register custom actions for specific contexts:

```go
customAction := MenuAction{
    ID:          "custom_action",
    Label:       "My Custom Action",
    Accelerator: "m",
    Destructive: false,
    Confirm:     false,
}

cms.RegisterAction(ContextQueueRow, customAction)

// Register a handler for the action
customHandler := func(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
    // Handle the action
    return func() tea.Msg {
        return CustomActionMsg{Data: "action executed"}
    }
}

cms.RegisterHandler("custom_action", customHandler)
```

## Message Types

The system generates various message types for action handling:

### Action Messages

```go
type PeekActionMsg struct {
    QueueName string
    Context   MenuContext
}

type EnqueueActionMsg struct {
    QueueName string
    Context   MenuContext
}

type RequeueActionMsg struct {
    JobID   string
    Context MenuContext
}

type CopyActionMsg struct {
    Text    string
    Type    string
    Context MenuContext
}
```

### Control Messages

```go
type ShowMenuMsg struct {
    Context MenuContext
}

type HideMenuMsg struct{}

type ZoneClickMsg struct {
    Zone     BubbleZone
    Position Position
    Button   int // 0=left, 1=right, 2=middle
}
```

## Usage Examples

### Basic Setup

```go
// Create and initialize the context menu system
cms := context_menus.New()
cmd := cms.Init()

// In your main model's Update method
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // Update the context menu system
    m.contextMenu, cmd = m.contextMenu.Update(msg)

    // Handle action messages
    switch msg := msg.(type) {
    case context_menus.PeekActionMsg:
        // Handle peek action
        return m, m.handlePeek(msg.QueueName)
    case context_menus.EnqueueActionMsg:
        // Handle enqueue action
        return m, m.handleEnqueue(msg.QueueName)
    }

    return m, cmd
}

// In your main model's View method
func (m Model) View() string {
    base := m.renderMainContent()
    menu := m.contextMenu.View()

    // Overlay the menu on top of the base content
    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Left, lipgloss.Top,
        base,
        lipgloss.Place(
            m.width, m.height,
            lipgloss.Left, lipgloss.Top,
            menu,
        ),
    )
}
```

### Registering Table Zones

```go
// When rendering a table, register zones for each row
func (m Model) renderTable() string {
    // Clear existing zones
    m.contextMenu.ClearZones()

    // Register zones for each table row
    for i, row := range m.tableRows {
        queueName := row[0] // Assuming first column is queue name
        err := m.contextMenu.RegisterTableRow(
            i,           // row index
            0,           // x position
            i + 5,       // y position (accounting for headers)
            m.width,     // full width
            queueName,   // queue name for context
        )
        if err != nil {
            // Handle error
        }
    }

    return tableContent
}
```

### Handling Destructive Actions

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case context_menus.ShowConfirmationMsg:
        // Show confirmation dialog for destructive actions
        m.confirmDialog = ConfirmDialog{
            Visible: true,
            Message: msg.Message,
            Action:  msg.Action,
            Context: msg.Context,
        }
        return m, nil

    case ConfirmDialogResult:
        if msg.Confirmed {
            // Execute the destructive action
            return m, context_menus.ConfirmationHandler(
                true,
                msg.Action,
                msg.Context,
            )
        }
        return m, nil
    }

    return m, nil
}
```

## Configuration

The system supports extensive configuration through the Config struct:

```go
config := context_menus.Config{
    Enabled: true,
    Animation: context_menus.AnimationConfig{
        Enabled:  true,
        Duration: 150 * time.Millisecond,
        Easing:   "ease-out",
    },
    Appearance: context_menus.AppearanceConfig{
        BorderStyle:      "rounded",
        BackgroundColor:  "235",
        HighlightColor:   "62",
        DestructiveColor: "196",
        MinWidth:         20,
        MaxWidth:         60,
    },
    Behavior: context_menus.BehaviorConfig{
        ConfirmDestructive: true,
        CloseAfterAction:   true,
        Mouse: context_menus.MouseConfig{
            RightClickEnabled:  true,
            HideOnClickOutside: true,
        },
        Keyboard: context_menus.KeyboardConfig{
            MKeyEnabled:     true,
            ArrowNavigation: true,
            VimNavigation:   true,
            AcceleratorKeys: true,
        },
    },
}

// Apply configuration
err := config.Apply(cms)
```

## Error Handling

The system defines specific error types for different failure scenarios:

```go
// Check for specific error types
if context_menus.IsErrorCode(err, context_menus.ErrCodeZoneNotFound) {
    // Handle zone not found error
}

// Get error details
if details := context_menus.GetErrorDetails(err); details != nil {
    zoneID := details["zoneID"]
    // Handle with additional context
}
```

## Best Practices

### Performance

1. **Clear zones** when rebuilding UI to prevent memory leaks
2. **Batch zone registration** when possible
3. **Use specific context types** for better action filtering

### User Experience

1. **Provide accelerator keys** for power users
2. **Group related actions** logically in menus
3. **Use confirmation dialogs** for destructive actions
4. **Keep menu labels concise** but descriptive

### Development

1. **Register custom handlers** for application-specific actions
2. **Test zone boundaries** to ensure accurate hit detection
3. **Handle all action messages** in your Update method
4. **Use the configuration system** for customizable behavior

## Integration with Existing TUI

The context menu system is designed to integrate seamlessly with existing Bubble Tea applications:

1. **Add the context menu system** to your main model
2. **Forward messages** to the context menu's Update method
3. **Render the menu overlay** in your View method
4. **Register zones** as you render UI components
5. **Handle action messages** in your Update method

This approach ensures minimal disruption to existing code while adding powerful context menu functionality.