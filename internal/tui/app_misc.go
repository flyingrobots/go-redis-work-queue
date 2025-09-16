package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
)

// addSample appends a value to a named series using StatsResult map.
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
	q := strings.TrimSpace(m.filter.Value())
	if q == "" {
		m.tbl.SetRows(m.allRows)
		m.peekTargets = append([]string(nil), m.allTargets...)
		return
	}
	labels := make([]string, len(m.allRows))
	for i, r := range m.allRows {
		labels[i] = r[0]
	}
	ranks := fuzzy.RankFindNormalizedFold(q, labels)
	sort.Sort(ranks)
	rows := make([]table.Row, 0, len(ranks))
	targets := make([]string, 0, len(ranks))
	for _, rk := range ranks {
		rows = append(rows, m.allRows[rk.OriginalIndex])
		targets = append(targets, m.allTargets[rk.OriginalIndex])
	}
	m.tbl.SetRows(rows)
	m.peekTargets = targets
}

// clamp limits v to the inclusive range [low, high].
