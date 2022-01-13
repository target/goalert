package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	tasks, err := Parse(os.Stdin)
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
