package budgeting

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BudgetTUI provides a terminal user interface for budget management
type BudgetTUI struct {
	app            *tview.Application
	pages          *tview.Pages
	budgetManager  *BudgetManager
	currentTenant  string
	refreshChannel chan struct{}
}

// NewBudgetTUI creates a new budget TUI
func NewBudgetTUI(budgetManager *BudgetManager, defaultTenant string) *BudgetTUI {
	return &BudgetTUI{
		app:            tview.NewApplication(),
		pages:          tview.NewPages(),
		budgetManager:  budgetManager,
		currentTenant:  defaultTenant,
		refreshChannel: make(chan struct{}, 1),
	}
}

// Start starts the budget TUI application
func (b *BudgetTUI) Start() error {
	b.setupLayout()
	b.app.SetRoot(b.pages, true)
	return b.app.Run()
}

// setupLayout creates the main TUI layout
func (b *BudgetTUI) setupLayout() {
	// Create main container with tabs
	main := tview.NewFlex().SetDirection(tview.FlexRow)

	// Header with tenant info and navigation
	header := b.createHeader()
	main.AddItem(header, 3, 0, false)

	// Tab container
	tabContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Create tab content
	overviewTab := b.createOverviewTab()
	trendsTab := b.createTrendsTab()
	controlsTab := b.createControlsTab()
	alertsTab := b.createAlertsTab()

	// Tab bar
	tabBar := b.createTabBar()
	main.AddItem(tabBar, 1, 0, false)

	// Tab content area
	tabPages := tview.NewPages()
	tabPages.AddPage("overview", overviewTab, true, true)
	tabPages.AddPage("trends", trendsTab, true, false)
	tabPages.AddPage("controls", controlsTab, true, false)
	tabPages.AddPage("alerts", alertsTab, true, false)

	main.AddItem(tabPages, 0, 1, true)

	// Footer with shortcuts
	footer := b.createFooter()
	main.AddItem(footer, 2, 0, false)

	// Set up key bindings
	main.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case '1':
			tabPages.SwitchToPage("overview")
			return nil
		case '2':
			tabPages.SwitchToPage("trends")
			return nil
		case '3':
			tabPages.SwitchToPage("controls")
			return nil
		case '4':
			tabPages.SwitchToPage("alerts")
			return nil
		case 'q':
			b.app.Stop()
			return nil
		case 'r':
			b.refresh()
			return nil
		}
		return event
	})

	b.pages.AddPage("main", main, true, true)
}

// createHeader creates the header with tenant information
func (b *BudgetTUI) createHeader() *tview.TextView {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	header.SetBorder(true).SetTitle(" Budget Management ")

	b.updateHeader(header)
	return header
}

// updateHeader updates the header content
func (b *BudgetTUI) updateHeader(header *tview.TextView) {
	now := time.Now().Format("2006-01-02 15:04:05")

	content := fmt.Sprintf(
		"[yellow]Tenant:[white] %s    [yellow]Time:[white] %s    [yellow]Status:[green] Active\n"+
			"[gray]Use Tab/1-4 to switch tabs, 'r' to refresh, 'q' to quit",
		b.currentTenant, now,
	)

	header.SetText(content)
}

// createTabBar creates the tab navigation bar
func (b *BudgetTUI) createTabBar() *tview.TextView {
	tabBar := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	tabBar.SetText("[yellow][1] Overview[white]  [2] Trends  [3] Controls  [4] Alerts")
	return tabBar
}

// createOverviewTab creates the overview tab content
func (b *BudgetTUI) createOverviewTab() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left panel: Budget summary
	budgetSummary := b.createBudgetSummary()
	flex.AddItem(budgetSummary, 0, 1, true)

	// Right panel: Top cost drivers
	costDrivers := b.createCostDrivers()
	flex.AddItem(costDrivers, 0, 1, false)

	return flex
}

// createBudgetSummary creates the budget summary panel
func (b *BudgetTUI) createBudgetSummary() *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	table.SetBorder(true).SetTitle(" Budget Summary ")

	// Headers
	headers := []string{"Queue", "Current", "Budget", "Utilization", "Status", "Trend"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	b.updateBudgetSummary(table)
	return table
}

// updateBudgetSummary updates the budget summary table
func (b *BudgetTUI) updateBudgetSummary(table *tview.Table) {
	budgets, err := b.budgetManager.ListBudgets(b.currentTenant, true)
	if err != nil {
		// Show error in table
		table.SetCell(1, 0, tview.NewTableCell("Error loading budgets").
			SetTextColor(tcell.ColorRed))
		return
	}

	row := 1
	for _, budget := range budgets {
		status, err := b.budgetManager.GetBudgetStatus(budget.ID)
		if err != nil {
			continue
		}

		queueName := budget.QueueName
		if queueName == "" {
			queueName = "all"
		}

		// Queue name
		table.SetCell(row, 0, tview.NewTableCell(queueName))

		// Current spend
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("$%.2f", status.CurrentSpend)))

		// Budget amount
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("$%.2f", status.BudgetAmount)))

		// Utilization with color coding
		utilizationCell := tview.NewTableCell(fmt.Sprintf("%.1f%%", status.Utilization*100))
		if status.Utilization >= 0.9 {
			utilizationCell.SetTextColor(tcell.ColorRed)
		} else if status.Utilization >= 0.75 {
			utilizationCell.SetTextColor(tcell.ColorYellow)
		} else {
			utilizationCell.SetTextColor(tcell.ColorGreen)
		}
		table.SetCell(row, 3, utilizationCell)

		// Status
		statusCell := tview.NewTableCell(strings.Title(status.CurrentThreshold))
		switch status.CurrentThreshold {
		case "block":
			statusCell.SetTextColor(tcell.ColorRed)
		case "throttle":
			statusCell.SetTextColor(tcell.ColorYellow)
		case "warning":
			statusCell.SetTextColor(tcell.ColorOrange)
		default:
			statusCell.SetTextColor(tcell.ColorGreen)
			statusCell.SetText("OK")
		}
		table.SetCell(row, 4, statusCell)

		// Trend (simplified)
		trendText := "→"
		if status.ProjectedSpend > status.BudgetAmount {
			trendText = "↗"
		}
		table.SetCell(row, 5, tview.NewTableCell(trendText))

		row++
	}
}

// createCostDrivers creates the cost drivers panel
func (b *BudgetTUI) createCostDrivers() *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	table.SetBorder(true).SetTitle(" Top Cost Drivers ")

	// Headers
	headers := []string{"Queue", "Cost", "Jobs", "Avg", "%"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	b.updateCostDrivers(table)
	return table
}

// updateCostDrivers updates the cost drivers table
func (b *BudgetTUI) updateCostDrivers(table *tview.Table) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Last 30 days

	drivers, err := b.budgetManager.aggregator.GetTopCostDrivers(b.currentTenant, startDate, endDate, 10)
	if err != nil {
		table.SetCell(1, 0, tview.NewTableCell("Error loading cost drivers").
			SetTextColor(tcell.ColorRed))
		return
	}

	row := 1
	for _, driver := range drivers {
		table.SetCell(row, 0, tview.NewTableCell(driver.QueueName))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("$%.2f", driver.TotalCost)))
		table.SetCell(row, 2, tview.NewTableCell(strconv.Itoa(driver.JobCount)))
		table.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("$%.4f", driver.AvgCostPerJob)))
		table.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%.1f%%", driver.Percentage)))
		row++
	}
}

// createTrendsTab creates the trends tab content
func (b *BudgetTUI) createTrendsTab() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Daily spending chart (text-based)
	dailyChart := b.createDailySpendingChart()
	flex.AddItem(dailyChart, 0, 2, true)

	// Component breakdown
	componentBreakdown := b.createComponentBreakdown()
	flex.AddItem(componentBreakdown, 0, 1, false)

	return flex
}

// createDailySpendingChart creates a text-based daily spending chart
func (b *BudgetTUI) createDailySpendingChart() *tview.TextView {
	chart := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)

	chart.SetBorder(true).SetTitle(" Daily Spending Trend (Last 30 Days) ")

	b.updateDailySpendingChart(chart)
	return chart
}

// updateDailySpendingChart updates the daily spending chart
func (b *BudgetTUI) updateDailySpendingChart(chart *tview.TextView) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Get aggregated daily data
	dailyData, err := b.budgetManager.aggregator.GetDailySpend(b.currentTenant, "", startDate, endDate)
	if err != nil {
		chart.SetText(fmt.Sprintf("[red]Error loading daily data: %v", err))
		return
	}

	if len(dailyData) == 0 {
		chart.SetText("[gray]No spending data available for the selected period")
		return
	}

	// Group by date and sum across queues
	dailyTotals := make(map[string]float64)
	for _, data := range dailyData {
		dateStr := data.Date.Format("01-02")
		dailyTotals[dateStr] += data.TotalCost
	}

	// Find max value for scaling
	maxCost := 0.0
	for _, cost := range dailyTotals {
		if cost > maxCost {
			maxCost = cost
		}
	}

	// Create text chart
	var chartText strings.Builder
	chartText.WriteString("Amount  Date      Bar Chart\n")
	chartText.WriteString("------  --------  --------------------\n")

	// Sort dates
	var dates []string
	for date := range dailyTotals {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	for _, date := range dates {
		cost := dailyTotals[date]
		barLength := int((cost / maxCost) * 20) // Scale to 20 chars max

		bar := strings.Repeat("█", barLength)
		if barLength < 20 {
			bar += strings.Repeat("░", 20-barLength)
		}

		chartText.WriteString(fmt.Sprintf("$%6.2f  %s  %s\n", cost, date, bar))
	}

	chart.SetText(chartText.String())
}

// createComponentBreakdown creates the component cost breakdown
func (b *BudgetTUI) createComponentBreakdown() *tview.Table {
	table := tview.NewTable().
		SetBorders(true)

	table.SetBorder(true).SetTitle(" Cost Breakdown by Component ")

	// Headers
	headers := []string{"Component", "Cost", "Percentage"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))
	}

	b.updateComponentBreakdown(table)
	return table
}

// updateComponentBreakdown updates the component breakdown table
func (b *BudgetTUI) updateComponentBreakdown(table *tview.Table) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	breakdown, err := b.budgetManager.aggregator.GetComponentBreakdown(b.currentTenant, "", startDate, endDate)
	if err != nil {
		table.SetCell(1, 0, tview.NewTableCell("Error loading breakdown").
			SetTextColor(tcell.ColorRed))
		return
	}

	// Calculate total for percentages
	total := 0.0
	for _, cost := range breakdown {
		total += cost
	}

	row := 1
	components := []string{"cpu", "memory", "payload", "redis", "network"}
	for _, component := range components {
		cost := breakdown[component]
		percentage := 0.0
		if total > 0 {
			percentage = cost / total * 100
		}

		table.SetCell(row, 0, tview.NewTableCell(strings.Title(component)))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("$%.2f", cost)))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f%%", percentage)))
		row++
	}
}

// createControlsTab creates the controls tab content
func (b *BudgetTUI) createControlsTab() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Budget creation form
	budgetForm := b.createBudgetForm()
	flex.AddItem(budgetForm, 0, 1, true)

	// Existing budgets list
	budgetsList := b.createBudgetsList()
	flex.AddItem(budgetsList, 0, 1, false)

	return flex
}

// createBudgetForm creates the budget creation/editing form
func (b *BudgetTUI) createBudgetForm() *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Budget Configuration ")

	form.AddInputField("Queue Name", "", 20, nil, nil)
	form.AddInputField("Budget Amount", "", 20, nil, nil)
	form.AddDropDown("Period Type", []string{"monthly", "weekly", "daily"}, 0, nil)
	form.AddInputField("Warning Threshold (%)", "75", 10, nil, nil)
	form.AddInputField("Throttle Threshold (%)", "90", 10, nil, nil)
	form.AddInputField("Block Threshold (%)", "100", 10, nil, nil)
	form.AddCheckbox("Enable Enforcement", true, nil)
	form.AddButton("Create Budget", func() {
		b.createBudgetFromForm(form)
	})
	form.AddButton("Clear", func() {
		form.Clear(true)
	})

	return form
}

// createBudgetFromForm creates a budget from form data
func (b *BudgetTUI) createBudgetFromForm(form *tview.Form) {
	// Extract form values (simplified implementation)
	// In a real implementation, this would validate and create the budget
	b.showMessage("Budget creation not implemented in demo")
}

// createBudgetsList creates the existing budgets list
func (b *BudgetTUI) createBudgetsList() *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	table.SetBorder(true).SetTitle(" Existing Budgets ")

	// Headers
	headers := []string{"Queue", "Amount", "Period", "Status", "Actions"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	b.updateBudgetsList(table)
	return table
}

// updateBudgetsList updates the budgets list table
func (b *BudgetTUI) updateBudgetsList(table *tview.Table) {
	budgets, err := b.budgetManager.ListBudgets(b.currentTenant, false)
	if err != nil {
		table.SetCell(1, 0, tview.NewTableCell("Error loading budgets").
			SetTextColor(tcell.ColorRed))
		return
	}

	row := 1
	for _, budget := range budgets {
		queueName := budget.QueueName
		if queueName == "" {
			queueName = "all"
		}

		table.SetCell(row, 0, tview.NewTableCell(queueName))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("$%.2f", budget.Amount)))
		table.SetCell(row, 2, tview.NewTableCell(budget.Period.Type))

		statusText := "Active"
		statusColor := tcell.ColorGreen
		if !budget.Active {
			statusText = "Inactive"
			statusColor = tcell.ColorGray
		}
		table.SetCell(row, 3, tview.NewTableCell(statusText).SetTextColor(statusColor))
		table.SetCell(row, 4, tview.NewTableCell("Edit | Delete"))

		row++
	}
}

// createAlertsTab creates the alerts tab content
func (b *BudgetTUI) createAlertsTab() *tview.Table {
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	table.SetBorder(true).SetTitle(" Budget Alerts ")

	// Headers
	headers := []string{"Time", "Type", "Queue", "Message", "Status"}
	for col, header := range headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false))
	}

	b.updateAlerts(table)
	return table
}

// updateAlerts updates the alerts table
func (b *BudgetTUI) updateAlerts(table *tview.Table) {
	// This would fetch real alerts from the database
	// For now, showing placeholder data
	table.SetCell(1, 0, tview.NewTableCell("No alerts").SetTextColor(tcell.ColorGray))
}

// createFooter creates the footer with keyboard shortcuts
func (b *BudgetTUI) createFooter() *tview.TextView {
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	footer.SetBorder(true)

	footerText := "[yellow]Shortcuts:[white] " +
		"[1-4] Switch tabs  " +
		"[r] Refresh  " +
		"[q] Quit  " +
		"[Tab] Navigate  " +
		"[Enter] Select"

	footer.SetText(footerText)
	return footer
}

// refresh refreshes all data in the current view
func (b *BudgetTUI) refresh() {
	select {
	case b.refreshChannel <- struct{}{}:
	default:
	}
}

// showMessage shows a modal message dialog
func (b *BudgetTUI) showMessage(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			b.pages.RemovePage("message")
		})

	b.pages.AddPage("message", modal, false, true)
}

// GetKeyboardShortcuts returns available keyboard shortcuts
func (b *BudgetTUI) GetKeyboardShortcuts() map[string]string {
	return map[string]string{
		"1":     "Switch to Overview tab",
		"2":     "Switch to Trends tab",
		"3":     "Switch to Controls tab",
		"4":     "Switch to Alerts tab",
		"r":     "Refresh current view",
		"q":     "Quit application",
		"Tab":   "Navigate between elements",
		"Enter": "Select/Edit item",
		"Esc":   "Cancel/Back",
	}
}