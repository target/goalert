package main

import (
	"golang.org/x/sys/unix"
)

func init() {
	shutdownSignals = append(shutdownSignals, unix.SIGTERM)
}
