package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
)

var logDir string

func isVar(b byte) bool {
	if b >= 'a' && b <= 'z' {
		return true
	}
	if b >= 'A' && b <= 'Z' {
		return true
	}
	if b == '_' {
		return true
	}

	return false
}

func envRep(r io.Reader) io.Reader {
	br := bufio.NewReader(r)

	rt, wt := io.Pipe()
	go func() {
		defer wt.Close()
		w := bufio.NewWriter(wt)
		defer w.Flush()

		var escape bool
		for {
			b, err := br.ReadByte()
			if err != nil {
				w.Flush()
				wt.CloseWithError(err)
				return
			}

			if escape {
				w.WriteByte(b)
				if b == '}' {
					w.Flush()
				}
				escape = false
				continue
			}
			if b == '\\' {
				escape = true
				continue
			}

			if b == '$' {
				var name string
				for {
					b, err = br.ReadByte()
					if err != nil {
						w.Flush()
						wt.CloseWithError(err)
						return
					}
					if !isVar(b) {
						br.UnreadByte()
						break
					}
					name += string(b)
				}
				io.WriteString(w, os.Getenv(name))
				continue
			}

			w.WriteByte(b)
			if b == '}' {
				w.Flush()
			}
		}
	}()

	return rt
}

func main() {
	flag.StringVar(&logDir, "logs", "", "Directory to store copies of all logs. Overwritten on each start.")
	rep := flag.Bool("replace", false, "Replace env vars specified with $VAR with the current value.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	var tasks []Task
	in := io.Reader(os.Stdin)
	if *rep {
		in = envRep(in)
	}
	dec := json.NewDecoder(in)
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
