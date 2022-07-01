package swo

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/swo/swogrp"
)

// PauseApps puts all nodes into a "paused" state:
// - Engine no longer cycles
// - Idle DB connections are disabled
// - Event listeners (postgres pub/sub) are disabled
func (m *Manager) PauseApps(ctx context.Context) error {
	swogrp.Progressf(ctx, "pausing apps")
	err := m.grp.Pause(ctx)
	if err != nil {
		return fmt.Errorf("pause: %w", err)
	}

	t := time.NewTicker(10 * time.Millisecond)
	defer t.Stop()
	for range t.C {
		s := m.grp.Status()
		var pausing, waiting int
		for _, node := range s.Nodes {
			for _, task := range node.Tasks {
				if task.Name == "pause" {
					pausing++
				}
				if task.Name == "resume-after" {
					waiting++
				}
			}
		}

		if pausing == 0 && waiting == len(s.Nodes) {
			break
		}
		if waiting == 0 {
			return fmt.Errorf("pause: timed out waiting for nodes to pause")
		}
	}

	return nil
}
