package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"
)

var (
	cancel       = func() {}
	cmd          *exec.Cmd
	mx           sync.Mutex
	testAddr     string
	startTimeout time.Duration

	lastExitCode int
	wait         chan struct{}
	stopping     chan struct{}
)

func main() {
	log.SetPrefix("procwrap: ")
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", "127.0.0.1:3033", "address.")
	flag.StringVar(&testAddr, "test", "", "TCP address to connect to as a healthcheck.")
	flag.DurationVar(&startTimeout, "timeout", 30*time.Second, "TCP test timeout when starting.")
	flag.Parse()

	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/signal", handleSignal)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, shutdownSignals...)
	go func() {
		sig := <-ch
		signal.Reset(shutdownSignals...)
		log.Printf("got signal %v, shutting down", sig)
		os.Exit(stopWith(true, sig))
	}()

	start(nil)
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
	_ = req.ParseForm()
	extraArgs := req.Form["extra-arg"]
	if len(extraArgs) == 0 {
		start(nil)
		return
	}

	if len(extraArgs) != 2 {
		http.Error(w, "invalid extra-arg", http.StatusBadRequest)
		return
	}
	if extraArgs[0] != "--experimental" {
		http.Error(w, "invalid extra-arg", http.StatusBadRequest)
		return
	}
	flags := strings.Split(extraArgs[1], ",")
	if extraArgs[1] == "" || len(flags) == 0 || len(flags) > 10 {
		http.Error(w, "invalid extra-arg", http.StatusBadRequest)
		return
	}

	start([]string{"--experimental", strings.Join(flags, ",")})
}

func handleSignal(w http.ResponseWriter, req *http.Request) {
	mx.Lock()
	defer mx.Unlock()

	if cmd == nil || cmd.Process == nil {
		http.Error(w, "not running", http.StatusServiceUnavailable)
		return
	}

	sig, ok := supportedSignals[req.FormValue("sig")]
	if !ok {
		http.Error(w, "unsupported signal", http.StatusBadRequest)
		return
	}

	if err := cmd.Process.Signal(sig); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func start(extraArgs []string) {
	mx.Lock()
	defer mx.Unlock()

	stop(false)
	args := flag.Args()
	args = append(args, extraArgs...)
	log.Println("starting", args[0])

	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	wait = make(chan struct{})
	stopping = make(chan struct{})

	// signal when the process has ended
	go func() {
		_ = cmd.Wait()
		close(wait)
	}()

	// ensure stop is called before process exits, or terminate
	go func() {
		select {
		case <-stopping:
		case <-wait:
			code := stop(true)
			log.Printf("process terminated unexpectedly: %v", code)
			os.Exit(code)
		}
	}()

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

func stopWith(lock bool, sig os.Signal) int {
	if lock {
		mx.Lock()
		defer mx.Unlock()
	}
	if cmd == nil {
		return lastExitCode
	}
	log.Println("stopping", flag.Arg(0))
	close(stopping)

	if err := cmd.Process.Signal(sig); err != nil {
		panic(err)
	}
	t := time.NewTimer(startTimeout)
	defer t.Stop()
	select {
	case <-t.C:
		log.Println("timed out waiting for process to exit sending KILL")
	case <-wait:
	}
	cancel()

	<-wait
	lastExitCode = cmd.ProcessState.ExitCode()
	cmd = nil
	return lastExitCode
}

func stop(lock bool) int { return stopWith(lock, os.Interrupt) }
