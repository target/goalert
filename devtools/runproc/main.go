package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/signal"
)

func main() {
	file := flag.String("f", "", "Procfile to run.")
	localFile := flag.String("l", "", "Local Procfile to append.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	if *file == "" {
		log.Fatal("No Procfile specified.")
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	var envData []byte
	if *localFile != "" {
		envData, _ = os.ReadFile(*localFile)
	}

	buf := bytes.NewBuffer(data)
	buf.WriteString("\n")
	buf.Write(envData)

	tasks, err := Parse(buf)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetFlags(log.Lshortfile)

	run := NewRunner(tasks)

	ch := make(chan os.Signal, 3)
	signal.Notify(ch, shutdownSignals...)
	go func() {
		<-ch
		go run.Stop()
		<-ch
		os.Exit(1)
	}()

	err = run.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
