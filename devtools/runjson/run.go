package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var colors = []color.Attribute{
	color.FgRed,
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgHiRed,
	color.FgHiGreen,
	color.FgHiYellow,
	color.FgHiBlue,
	color.FgHiMagenta,
	color.FgHiCyan,
}

var mx sync.Mutex

func newWritePrefixer(attr color.Attribute, prefix string, out io.Writer) io.Writer {
	r, w := io.Pipe()

	s := bufio.NewScanner(r)
	pref := color.New(attr, color.Bold)
	txt := color.New(attr)

	go func() {
		for s.Scan() {
			mx.Lock()
			pref.Fprint(out, prefix)
			txt.Fprintln(out, s.Text())
			mx.Unlock()
		}
		r.CloseWithError(s.Err())
	}()

	return w
}

func Run(ctx context.Context, tasks []Task) error {
	l := 0
	for _, t := range tasks {
		if len(t.Name) > l {
			l = len(t.Name)
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var alwaysPass bool

	w := colorable.NewColorable(os.Stdout)

	ch := make(chan error, len(tasks))
	for i, t := range tasks {
		go func(i int, t Task) {
			attr := colors[i%len(colors)]
			err := t.run(ctx, l, attr, w)
			ch <- err
			if t.ExitAfter {
				alwaysPass = err == nil
				cancel()
			}
		}(i, t)
	}

	var hasError bool
	for range tasks {
		err := <-ch
		if err != nil {
			cancel()
			if !errors.Is(err, context.Canceled) {
				hasError = true
				log.Println("ERROR:", err)
			}
		}
	}

	if hasError && !alwaysPass {
		return errors.New("one or more tasks failed")
	}

	return nil
}
