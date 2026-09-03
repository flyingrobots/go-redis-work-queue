package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
)

func TestQueueMouseWheelRefreshesSelectionDecoration(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	m := initialModel(cfg, nil, zap.NewNop(), time.Second, Options{})
	t.Cleanup(m.cancel)
	m.tbl.SetHeight(4)
	m.allRowData = []queueRowData{
		{label: "first", count: 1},
		{label: "second", count: 2},
	}
	m.allTargets = []string{"first", "second"}
	m.applyFilterAndSetRows()

	next, _ := m.Update(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	got := next.(model)
	if got.tbl.Cursor() != 1 {
		t.Fatalf("cursor = %d, want 1", got.tbl.Cursor())
	}
	if strings.Contains(got.allRows[0][0], selectionGlyph) {
		t.Fatal("old row retained the selection glyph after mouse-wheel movement")
	}
	if !strings.Contains(got.allRows[1][0], selectionGlyph) {
		t.Fatal("new row is missing the selection glyph after mouse-wheel movement")
	}

	next, _ = got.Update(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	got = next.(model)
	if got.tbl.Cursor() != 0 {
		t.Fatalf("cursor after wheel up = %d, want 0", got.tbl.Cursor())
	}
	if !strings.Contains(got.allRows[0][0], selectionGlyph) || strings.Contains(got.allRows[1][0], selectionGlyph) {
		t.Fatal("wheel up did not return the selection decoration to the first row")
	}
}
