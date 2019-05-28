package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
)

var logDir string

func main() {
	flag.StringVar(&logDir, "logs", "", "Directory to store copies of all logs. Overwritten on each start.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	var tasks []Task
	dec := json.NewDecoder(os.Stdin)
	dec.DisallowUnknownFields()
	err := dec.Decode(&tasks)
	if err != nil {
		log.Fatal("decode input:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got signal, terminating.")
		cancel()
	}()

	err = Run(ctx, tasks)
	if err != nil {
		log.Fatal("run:", err)
	}
}
