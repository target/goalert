package main

type ProcessRunner interface {
	Start()
	Stop()
	Kill()
	Done() bool
	Wait() bool
}
