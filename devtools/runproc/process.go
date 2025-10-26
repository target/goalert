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

	"github.com/creack/pty/v2"
	"github.com/fatih/color"
)

type Process struct {
	Task

	p   io.Writer
	cmd *exec.Cmd
	pty *os.File

	state chan ProcessState

	result bool
	exited chan struct{}
}

type ProcessState int

const (
	ProcessStateIdle ProcessState = iota
	ProcessStateStarting
	ProcessStateRunning
	ProcessStateStopping
	ProcessStateKilling
	ProcessStateRestarting
	ProcessStateExited
)

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

var pColors = make(map[string]color.Attribute)

func coloredName(pName string) string {
	logMx.Lock()
	defer logMx.Unlock()

	c, ok := pColors[pName]
	if !ok {
		c = colors[colorIndex%len(colors)]
		colorIndex++
		pColors[pName] = c
	}

	return color.New(color.Reset, c).Sprint(pName)
}

func NewProcess(t Task, padding int) *Process {
	pName := t.Name + strings.Repeat(" ", padding-len(t.Name))
	pName = coloredName(pName)
	stateCh := make(chan ProcessState, 1)
	stateCh <- ProcessStateIdle
	return &Process{
		Task:   t,
		p:      NewPrefixer(os.Stdout, pName),
		state:  stateCh,
		exited: make(chan struct{}),
	}
}

var _ ProcessRunner = (*Process)(nil)

var logMx sync.Mutex

func (p *Process) logError(err error) {
	logMx.Lock()
	defer logMx.Unlock()
	_, _ = color.New(color.Reset, color.FgRed).Fprintln(p.p, err.Error())
}

func (p *Process) logAction(s string) {
	logMx.Lock()
	defer logMx.Unlock()
	_, _ = color.New(color.Reset, color.Bold).Fprintln(p.p, s)
}

func (p *Process) Stop() {
	s := <-p.state
	if s != ProcessStateRunning {
		// nothing to do
		p.state <- s
		return
	}

	p.logAction("Stopping...")
	go p.gracefulTerm()
	p.state <- ProcessStateStopping

	t := time.NewTimer(time.Second)
	defer t.Stop()
	select {
	case <-p.exited:
	case <-t.C:
		p.Kill()
	}
}

func (p *Process) gracefulTerm() {
	if p.pty != nil {
		// Since this could be called after the process exits, we can ignore the error.
		_, _ = io.WriteString(p.pty, "\x03")
		time.Sleep(100 * time.Millisecond)

		_, _ = io.WriteString(p.pty, "\x03")
		return
	}

	// The process may have already terminated and we can ignore the error.
	_ = p.cmd.Process.Signal(os.Interrupt)
}

func (p *Process) Kill() {
	s := <-p.state
	switch s {
	case ProcessStateStopping, ProcessStateRunning:
	default:
		return
	}

	p.logAction("Killing...")

	// The process may have already terminated and we can ignore the error.
	_ = p.cmd.Process.Kill()
	p.state <- ProcessStateKilling

	<-p.exited
}

func (p *Process) run() error {
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
			go p.copyIO(p.p, p.pty)
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
		return err
	}

	go func() {
		err := p.cmd.Wait()
		<-p.state
		p.result = err == nil
		p.logAction("Exited.")
		if err != nil {
			p.logError(err)
		}
		close(p.exited)
		p.state <- ProcessStateExited
	}()

	return nil
}

func (p *Process) Start() {
	s := <-p.state
	if s != ProcessStateIdle {
		p.state <- s
		return
	}

	p.logAction("Starting...")
	err := p.run()
	if err != nil {
		p.logError(err)
		close(p.exited)
		p.state <- ProcessStateExited
		return
	}

	p.state <- ProcessStateRunning
}

func (p *Process) Wait() bool {
	<-p.exited

	return p.result && p.OneShot
}

func (p *Process) Done() bool {
	select {
	case <-p.exited:
		return true
	default:
		return false
	}
}

func (p *Process) copyIO(dst io.Writer, src io.Reader) {
	_, err := io.Copy(dst, src)
	if err != nil {
		p.logError(err)
	}
}
