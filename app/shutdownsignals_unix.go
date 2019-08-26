// +build !windows

package app

import (
	"syscall"
)

func init() {
	shutdownSignals = append(shutdownSignals, syscall.SIGTERM)
	triggerSignals = append(triggerSignals, syscall.SIGUSR2)
}
