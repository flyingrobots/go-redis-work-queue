package tui

import (
	"fmt"
	"strings"
)

func cycleBenchFocus(m *model) {
	if m.benchCount.Focused() {
		m.benchCount.Blur()
		m.benchRate.Focus()
		return
	}
	if m.benchRate.Focused() {
		m.benchRate.Blur()
		m.benchPriority.Focus()
		return
	}
	if m.benchPriority.Focused() {
		m.benchPriority.Blur()
		m.benchTimeout.Focus()
		return
	}
	m.benchTimeout.Blur()
	m.benchCount.Focus()
}

func atoiDefault(s string, def int) int {
	var v int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
	if err != nil {
		return def
	}
	return v
}

// clamp limits v to the inclusive range [low, high].
func clamp(v, low, high int) int {
	if high < low {
		return low
	}
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
