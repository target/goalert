package main

type Task struct {
	// Name used in logs
	Name string

	// Command is the primary command to run.
	Command string

	// WatchFiles will cause the process to restart if any if the files change (can be patterns).
	WatchFiles []string
}
