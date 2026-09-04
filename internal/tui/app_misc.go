package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

// addSample appends a value to a named series using StatsResult map.

const (
	selectionGlyph      = "▸"
	selectionSpacer     = " "
	queueWarnThreshold  = 1000
	queueAlertThreshold = 5000
)

var (
	queueRowBaseStyle     = lipgloss.NewStyle().Padding(0, 1, 0, 1)
	queueRowStripeStyle   = queueRowBaseStyle.Copy().Background(colorBgSecondary)
	queueRowSelectedStyle = queueRowBaseStyle.Copy().Background(colorBgAccent)
)

func (m *model) addSample(alias, key string, s admin.StatsResult) {
	if alias == "" || key == "" {
		return
	}
	display := fmt.Sprintf("%s (%s)", alias, key)
	val := s.Queues[display]
	arr := m.series[alias]
	arr = append(arr, float64(val))
	if len(arr) > m.seriesMax {
		arr = arr[len(arr)-m.seriesMax:]
	}
	m.series[alias] = arr
}

func (m *model) applyFilterAndSetRows() {
	filter := strings.TrimSpace(m.filter.Value())
	filtered := make([]queueRowData, 0)
	targets := make([]string, 0)

	if filter == "" {
		filtered = append(filtered, m.allRowData...)
		targets = append(targets, m.allTargets...)
	} else {
		labels := make([]string, len(m.allRowData))
		for i, data := range m.allRowData {
			labels[i] = data.label
		}
		ranks := fuzzy.RankFindNormalizedFold(filter, labels)
		sort.Sort(ranks)
		filtered = make([]queueRowData, 0, len(ranks))
		targets = make([]string, 0, len(ranks))
		for _, rk := range ranks {
			filtered = append(filtered, m.allRowData[rk.OriginalIndex])
			targets = append(targets, m.allTargets[rk.OriginalIndex])
		}
	}

	m.filteredRowData = filtered
	m.peekTargets = targets

	if len(filtered) == 0 {
		m.tbl.SetRows(nil)
		m.allRows = nil
		return
	}

	cur := clamp(m.tbl.Cursor(), 0, len(filtered)-1)
	rows := make([]table.Row, len(filtered))
	for i, data := range filtered {
		rows[i] = m.renderQueueRow(data, i, i == cur)
	}

	m.tbl.SetRows(rows)
	m.tbl.SetCursor(cur)
	m.allRows = rows
}

func backgroundStyleForRow(index int, selected bool) lipgloss.Style {
	switch {
	case selected:
		return queueRowSelectedStyle
	case index%2 == 1:
		return queueRowStripeStyle
	default:
		return queueRowBaseStyle
	}
}

func countStyleFor(count int, base lipgloss.Style, selected bool) lipgloss.Style {
	style := base.Copy().Align(lipgloss.Right).Width(10)
	if selected {
		return style.Foreground(colorTextInverse).Bold(true)
	}
	switch {
	case count >= queueAlertThreshold:
		return style.Foreground(colorError).Bold(true)
	case count >= queueWarnThreshold:
		return style.Foreground(colorWarning).Bold(true)
	case count > 0:
		return style.Foreground(colorSuccess)
	default:
		return style.Foreground(colorTextSecondary)
	}
}

func (m *model) renderQueueRow(data queueRowData, index int, selected bool) table.Row {
	base := backgroundStyleForRow(index, selected)
	glyph := selectionSpacer
	if selected {
		glyph = selectionGlyph
	}

	labelStyle := base.Copy().Foreground(colorTextPrimary)
	if selected {
		labelStyle = labelStyle.Foreground(colorTextInverse).Bold(true)
	}

	countStyle := countStyleFor(data.count, base, selected)

	labelCell := labelStyle.Render(fmt.Sprintf("%s %s", glyph, data.label))
	countCell := countStyle.Render(fmt.Sprintf("%d", data.count))

	return table.Row{labelCell, countCell}
}

func (m *model) refreshDecoratedRows() {
	if len(m.filteredRowData) == 0 {
		return
	}

	cur := clamp(m.tbl.Cursor(), 0, len(m.filteredRowData)-1)
	rows := make([]table.Row, len(m.filteredRowData))
	for i, data := range m.filteredRowData {
		rows[i] = m.renderQueueRow(data, i, i == cur)
	}

	m.tbl.SetRows(rows)
	m.tbl.SetCursor(cur)
	m.allRows = rows
}
