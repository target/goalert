package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
)

var logDir string

func main() {
	flag.StringVar(&logDir, "logs", "", "Directory to store copies of all logs. Overwritten on each start.")
	file := flag.String("file", "-", "File to load config from.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	var tasks []Task

	var in io.Reader
	if *file == "-" {
		in = io.Reader(os.Stdin)
	} else {
		fd, err := os.Open(*file)
		if err != nil {
			log.Fatal("open file:", err)
		}
		defer fd.Close()
		in = fd
	}

	dec := json.NewDecoder(in)
	var raw json.RawMessage
	err := dec.Decode(&raw)
	if err != nil {
		log.Fatal("read input:", err)
	}
	raw = json.RawMessage(os.ExpandEnv(string(raw)))

	dec = json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	err = dec.Decode(&tasks)
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
