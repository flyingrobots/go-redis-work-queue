package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// doEnqueueCmd pushes count dummy payloads to the given queue key.
func (m model) doEnqueueCmd(queueKey string, count int) tea.Cmd {
	return func() tea.Msg {
		if queueKey == "" || queueKey == m.cfg.Worker.CompletedList || queueKey == m.cfg.Worker.DeadLetterList {
			return enqueueMsg{n: 0, key: queueKey, err: fmt.Errorf("invalid target queue")}
		}
		n := 0
		for i := 0; i < count; i++ {
			payload := fmt.Sprintf(`{"id":"tui-%d","filepath":"/tui/%d","filesize":1,"priority":"%s","retries":0,"creation_time":"%s","trace_id":"","span_id":""}`,
				time.Now().UnixNano(), i, "manual", time.Now().UTC().Format(time.RFC3339Nano))
			if err := m.rdb.LPush(m.ctx, queueKey, payload).Err(); err != nil {
				return enqueueMsg{n: n, key: queueKey, err: err}
			}
			n++
		}
		return enqueueMsg{n: n, key: queueKey, err: nil}
	}
}
