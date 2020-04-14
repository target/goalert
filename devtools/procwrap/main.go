package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var cancel = func() {}
var cmd *exec.Cmd
var mx sync.Mutex
var testAddr string

func main() {
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", "127.0.0.1:3033", "address.")
	flag.StringVar(&testAddr, "test", "", "TCP address to connnect to as a healthcheck.")
	flag.Parse()

	start()
	defer stop(true)

	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/signal", handleSignal)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}

}

func handleStop(w http.ResponseWriter, req *http.Request) {
	stop(true)
}

func handleStart(w http.ResponseWriter, req *http.Request) {
	start()
}

func handleSignal(w http.ResponseWriter, req *http.Request) {
	mx.Lock()
	defer mx.Unlock()

	if cmd == nil || cmd.Process == nil {
		http.Error(w, "not running", http.StatusServiceUnavailable)
		return
	}

	if req.FormValue("sig") != "SIGUSR2" {
		http.Error(w, "unsupported signal", http.StatusBadRequest)
		return
	}

	err := cmd.Process.Signal(syscall.SIGUSR2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func start() {
	mx.Lock()
	defer mx.Unlock()
	// since it is a stub function, no not nil check needed
	stop(false)

	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	rawBin := flag.Arg(0)
	bin, err := exec.LookPath(rawBin)
	if err != nil {
		log.Fatalf("lookup error %v", err)
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		log.Fatalf("lookup error %v", err)
	}

	cmd = exec.CommandContext(ctx, bin, flag.Args()[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = ""

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	if testAddr == "" {
		return
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	for i := 0; i < 300; i++ {
		c, err := net.Dial("tcp", testAddr)
		if err != nil {
			continue
		}
		c.Close()
		return
	}

	log.Fatal("failed to start after 30 seconds.")
}

func stop(lock bool) {
	if lock {
		mx.Lock()
		defer mx.Unlock()
	}
	// since it is a stub function, no not nil check needed
	cancel()
	if cmd == nil {
		return
	}

	// waits for cancel to finish executing
	// waits for process to actually stop running
	cmd.Wait()
}

/*
func (t *Task) run(ctx context.Context) error {
	// Need logDir and pidDir?

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

	cmd.Dir = t.Dir
	// cmd.Stdout = stdout
	// cmd.Stderr = stderr

	err = cmd.Start()
	return nil
}
*/
