package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Runner struct {
	procs []*Process

	doneCh chan struct{}
	stopCh chan struct{}

	stopOnce sync.Once
}

func NewRunner(tasks []Task) *Runner {
	procs := make([]*Process, 0, len(tasks))
	logName := "RUNPROC"
	maxLen := len(logName)
	for _, t := range tasks {
		if len(t.Name) > maxLen {
			maxLen = len(t.Name)
		}
	}

	if maxLen > len(logName) {
		logName += strings.Repeat(" ", maxLen-len(logName))
	}
	log.SetOutput(&prefixer{
		out:    log.Default().Writer(),
		prefix: color.New(color.BgRed, color.FgWhite).Sprint(logName),
	})
	for _, t := range tasks {
		procs = append(procs, NewProcess(t, maxLen))
	}
	return &Runner{
		procs:  procs,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (r *Runner) Run() error {
	defer close(r.doneCh)
	result := make(chan bool, len(r.procs))
	for _, proc := range r.procs {
		proc.Start()
		go func(p *Process) {
			result <- p.Wait()
		}(proc)
	}

	var err error
	for range r.procs {
		if !<-result {
			go r.Stop()
			err = fmt.Errorf("one or more commands failed")
		}
	}

	return err
}

func (r *Runner) Stop() {
	r.stopOnce.Do(r._stop)
}

func (r *Runner) _stop() {
	close(r.stopCh)

	for _, proc := range r.procs {
		go proc.Stop()
	}

	t := time.NewTimer(time.Second)
	defer t.Stop()
	select {
	case <-r.doneCh:
	case <-t.C:
		for _, proc := range r.procs {
			select {
			case <-proc.cmdDone:
			default:
				proc.Kill()
			}
		}
	}

	<-r.doneCh
}
