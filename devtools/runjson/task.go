package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// A Task is an independent unit of work
type Task struct {
	// Name is used to identify the task in the case of errors, as well as the prefix for logs.
	Name string

	// Dir is the working directory. If empty, the current working directory is used.
	Dir string

	// Quiet will omit starting messages.
	Quiet bool

	// Before is a task that must complete before the current one starts.
	Before *Task

	// After is a task that will run after the current one exits for any reason.
	After *Task

	// Command contains the binary to run, followed by any/all arguments.
	Command []string

	// Env parameters will be set in addition to any current ones.
	Env []string

	// Restart will cause the process to restart automatically if it terminates.
	Restart bool

	// IgnoreErrors will allow all processes to continue, even if a non-zero exit status is returned.
	IgnoreErrors bool

	// Watch will cause the process to restart if/when the binary changes.
	Watch bool

	// ExitAfter, if true, will cause all tasks to be terminated when this one finishes.
	ExitAfter bool
}

func hashFile(path string) string {
	fd, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer fd.Close()

	h := sha1.New()
	io.Copy(h, fd)
	return hex.EncodeToString(h.Sum(nil))
}

func (t *Task) run(ctx context.Context, pad int, attr color.Attribute, w io.Writer) error {
	c := color.New(attr)
	cb := color.New(attr, color.Bold)

	pref := fmt.Sprintf("\t%-"+strconv.Itoa(pad)+"s", t.Name)

	stdout := newWritePrefixer(attr, pref+" (out): ", w)
	stderr := newWritePrefixer(attr, pref+" (err): ", w)
	if logDir != "" {
		os.MkdirAll(logDir, 0755)
		outFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("%s.out.log", t.Name)))
		if err != nil {
			return errors.Wrap(err, "create stdout log")
		}
		defer outFile.Close()
		errFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("%s.err.log", t.Name)))
		if err != nil {
			return errors.Wrap(err, "create stderr log")
		}
		defer errFile.Close()
		stdout = io.MultiWriter(stdout, outFile)
		stderr = io.MultiWriter(stderr, errFile)
	}

	defer log.Println(color.New(color.BgRed).Sprint(" QUIT "), cb.Sprint(t.Name))
	rawBin := t.Command[0]
	if t.Dir != "" && strings.HasPrefix(rawBin, ".") {
		rawBin = filepath.Join(t.Dir, rawBin)
	}
	bin, err := exec.LookPath(rawBin)
	if err != nil {
		return errors.Wrapf(err, "lookup %s", rawBin)
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		return errors.Wrapf(err, "lookup %s", rawBin)
	}

	if t.Before != nil {
		err = t.Before.run(ctx, pad, attr|color.Faint, w)
		if err != nil {
			return errors.Wrap(err, "before")
		}
	}
	if t.After != nil {
		defer t.After.run(context.Background(), pad, attr|color.Faint, w)
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		procCtx, cancel := context.WithCancel(ctx)
		hash := hashFile(bin)
		if t.Watch {
			go func() {
				defer cancel()
				t := time.NewTicker(time.Second)
				for {
					select {
					case <-procCtx.Done():
						return
					case <-t.C:
					}
					newHash := hashFile(bin)
					if newHash == hash {
						continue
					}
					return
				}
			}()
		}

		cmd := exec.CommandContext(procCtx, bin, t.Command[1:]...)
		cmd.Dir = t.Dir
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Env = append(os.Environ(), t.Env...)

		if !t.Quiet {
			log.Println(c.Sprint("Starting"), cb.Sprintf("%s[%s]", t.Name, hash), c.Sprint(bin+" "+strings.Join(t.Command[1:], " ")))
		}

		err := cmd.Start()
		if err != nil && !t.IgnoreErrors {
			cancel()
			return errors.Wrapf(err, "run %s", t.Name)
		}
		if pidDir != "" {
			err = ioutil.WriteFile(filepath.Join(pidDir, t.Name+".pid"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
			if err != nil && !t.IgnoreErrors {
				cancel()
				return errors.Wrapf(err, " record pid %s", t.Name)
			}
		}

		err = cmd.Wait()
		cancel()
		if err != nil && !t.IgnoreErrors {
			return errors.Wrapf(err, "run %s", t.Name)
		}
		if !t.Restart {
			break
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
