package main

import "os"

var shutdownSignals = []os.Signal{os.Interrupt}

var supportedSignals = map[string]os.Signal{
	"SIGINT":  os.Interrupt,
	"SIGKILL": os.Kill,
}
