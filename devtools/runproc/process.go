package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/fatih/color"
)

type Process struct {
	Task

	p   io.Writer
	cmd *exec.Cmd
	pty *os.File

	cmdErr  error
	cmdDone chan struct{}

	doneCh, failCh   chan struct{}
	stopped, started bool
	mx               sync.Mutex
}

var colors = []color.Attribute{
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgRed,
	color.FgHiGreen,
	color.FgHiYellow,
	color.FgHiBlue,
	color.FgHiMagenta,
	color.FgHiCyan,
	color.FgHiRed,
}
var colorIndex int

func NewProcess(t Task, padding int) *Process {
	pName := t.Name + strings.Repeat(" ", padding-len(t.Name))
	pName = color.New(color.Reset, colors[colorIndex%len(colors)]).Sprint(pName)
	colorIndex++
	return &Process{
		Task:   t,
		p:      NewPrefixer(os.Stdout, pName),
		doneCh: make(chan struct{}),
		failCh: make(chan struct{}),
	}
}

var logMx sync.Mutex

func (p *Process) logError(err error) {
	logMx.Lock()
	defer logMx.Unlock()
	color.New(color.Reset, color.FgRed).Fprintln(p.p, err.Error())
}
func (p *Process) logAction(s string) {
	logMx.Lock()
	defer logMx.Unlock()
	color.New(color.Reset, color.Bold).Fprintln(p.p, s)
}
func (p *Process) Stop() {
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.stopped {
		return
	}

	p.logAction("Stopping...")

	p.gracefulTerm()
	<-p.cmdDone
}
func (p *Process) gracefulTerm() {
	logMx.Lock()
	defer logMx.Unlock()
	if p.pty != nil {
		io.WriteString(p.pty, "\x03")
		time.Sleep(100 * time.Millisecond)
		io.WriteString(p.pty, "\x03")
		return
	}

	p.cmd.Process.Signal(os.Interrupt)
}
func (p *Process) Kill() {
	p.logAction("Killing...")
	logMx.Lock()
	p.cmd.Process.Kill()
	if p.pty != nil {
		p.pty.Close()
	}
	logMx.Unlock()

	<-p.cmdDone
}

func (p *Process) finish() {
	p.stopped = true
	close(p.cmdDone)

	defer p.logAction("Exited.")
	if p.cmdErr != nil {
		p.logError(p.cmdErr)
		close(p.failCh)
		return
	}

	close(p.doneCh)
}

func (p *Process) watchFiles() {
	hash := groupHash(p.WatchFiles)
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-p.doneCh:
			return
		case <-p.failCh:
			return
		case <-t.C:
		}

		newHash := groupHash(p.WatchFiles)
		if newHash == hash {
			continue
		}

		hash = newHash
		p.Restart()
	}
}

func (p *Process) run() {
	p.cmd = exec.Command("sh", "-ce", p.Command)

	ptty, tty, err := pty.Open()
	if err == nil {
		p.pty = ptty
		p.cmd.SysProcAttr = &syscall.SysProcAttr{
			Setctty: true,
			Setsid:  true,
			Ctty:    3,
		}
		p.cmd.ExtraFiles = []*os.File{tty}
		p.cmd.Stdout = tty
		p.cmd.Stderr = tty
		p.cmd.Stdin = tty
		err = p.cmd.Start()
		if err == nil {
			go io.Copy(p.p, p.pty)
		} else {
			p.pty.Close()
		}
	} else if errors.Is(err, pty.ErrUnsupported) {
		p.pty = nil
		p.cmd.Stdout = p.p
		p.cmd.Stderr = p.p
		err = p.cmd.Start()
	}
	if err != nil {
		p.cmdErr = err
		p.finish()
		return
	}

	p.cmdDone = make(chan struct{})
	go func() {
		p.cmdErr = p.cmd.Wait()
		p.finish()
	}()
}

func (p *Process) Start() {
	p.logAction("Starting...")
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.stopped || p.started {
		return
	}

	if len(p.WatchFiles) > 0 {
		go p.watchFiles()
	}
	p.run()

	p.started = true
}

func (p *Process) Restart() {
	p.logAction("Restarting...")
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.stopped {
		return
	}

	if !p.started {
		panic("cannot restart process that didn't start")
	}

	p.gracefulTerm()
	<-p.cmdDone
	if p.cmdErr != nil {
		p.logError(p.cmdErr)
	}

	p.run()
}

func (p *Process) Wait() bool {
	select {
	case <-p.doneCh:
		return true
	case <-p.failCh:
		return false
	}
}
