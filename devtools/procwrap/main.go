package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	cancel       = func() {}
	cmd          *exec.Cmd
	mx           sync.Mutex
	testAddr     string
	startTimeout time.Duration
)

func main() {
	log.SetPrefix("procwrap: ")
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", "127.0.0.1:3033", "address.")
	flag.StringVar(&testAddr, "test", "", "TCP address to connnect to as a healthcheck.")
	flag.DurationVar(&startTimeout, "timeout", 30*time.Second, "TCP test timeout when starting.")
	flag.Parse()

	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/signal", handleSignal)

	start()
	defer stop(true)

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("listen:", err)
	}

	log.Println("listening:", l.Addr().String())

	err = http.Serve(l, nil)
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
	log.Println("starting", flag.Arg(0))

	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	cmd = exec.CommandContext(ctx, flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	if testAddr == "" {
		return
	}

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	to := time.NewTimer(startTimeout)
	defer to.Stop()

	for {
		select {
		case <-to.C:
			log.Fatal("failed to start after 30 seconds.")
		case <-t.C:
			c, err := net.Dial("tcp", testAddr)
			if err != nil {
				continue
			}
			c.Close()
			log.Println("started", flag.Arg(0))
			return
		}
	}

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
	log.Println("stopping", flag.Arg(0))

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
