package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type Task struct {

	// Command contains the binary to run, followed by any/all arguments.
	Command []string

	// IgnoreErrors will allow all processes to continue, even if a non-zero exit status is returned.
	IgnoreErrors bool

	// maybe we might need Dir here
}

func main() {
	http.HandleFunc("/stop", stop)
	http.HandleFunc("/start", start)
	// http.ListenAndServe(":9090", nil)

	var t Task

	err = t.run(ctx)
	if err != nil {
		log.Fatal("run:", err)
	}

	// log.Println("Listening:", *addr)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func stop(w http.ResponseWriter, req *http.Request) {
	fmt.Println("stop")
}

func start(w http.ResponseWriter, req *http.Request) {
	fmt.Println("start")
}

func (t *Task) run(ctx context.Context) error {
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", ":9090", "address.")
	flag.Parse()

	// Get binary to be run
	rawBin := t.Command[0]

	bin, err := exec.LookPath(rawBin)
	if err != nil {
		return errors.Wrapf(err, "lookup %s", rawBin)
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		return errors.Wrapf(err, "lookup %s", rawBin)
	}

	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, bin, t.Command[1:]...)

	err := cmd.Start()
	if err != nil && !t.IgnoreErrors {
		cancel()
		return errors.Wrapf(err, "run %s", t.Name)
	}

	// Do we need cmd.Wait() here??

}
