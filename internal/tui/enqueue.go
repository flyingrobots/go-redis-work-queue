package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
)

// doEnqueueCmd pushes count dummy benchmark jobs to the given queue key.
func (m model) doEnqueueCmd(queueKey string, count int) tea.Cmd {
	return func() tea.Msg {
		if queueKey == "" || queueKey == m.cfg.Worker.CompletedList || queueKey == m.cfg.Worker.DeadLetterList {
			return enqueueMsg{n: 0, key: queueKey, err: fmt.Errorf("invalid target queue")}
		}
		n := 0
		for i := 0; i < count; i++ {
			job := queue.NewJob(fmt.Sprintf("tui-%d", time.Now().UnixNano()), fmt.Sprintf("/tui/%d", i), 1, "manual", "", "")
			if err := queue.Enqueue(m.ctx, m.rdb, queueKey, job, m.cfg.Queue.MaxPayloadSize); err != nil {
				return enqueueMsg{n: n, key: queueKey, err: err}
			}
			n++
		}
		return enqueueMsg{n: n, key: queueKey, err: nil}
	}
}
