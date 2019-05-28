package harness

import "syscall"

func (h *Harness) trigger() {
	if h.cmd.Process != nil {
		h.cmd.Process.Signal(syscall.SIGUSR2)
	}
}
