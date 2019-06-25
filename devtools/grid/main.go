package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
)

func main() {
	config := flag.String("config", "", "JSON file for advanced configuration.")
	file := flag.String("file", "-", "Filename to load parameters from (one per line).")
	flag.IntVar(&minPort, "min-port", minPort, "Minimum port number to reserve (when $_PORT:1 and friends are used).")
	flag.IntVar(&maxPort, "max-port", maxPort, "Maximum port number to reserve.")
	flag.StringVar(&namePrefix, "name-prefix", namePrefix, "Prefix used for $_NAME variables.")
	maxP := flag.Int("p", runtime.NumCPU(), "Maximum number of parallel processes (0 to disable limit).")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	// initialize port manager
	startPorts()

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

	var cfg Config
	if *config != "" {
		data, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Fatal("read config file:", err)
		}
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			log.Fatal("parse config file:", err)
		}
	}

	sc := bufio.NewScanner(in)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = maxP
	args := flag.Args()
	for sc.Scan() {
		cur := strings.TrimSpace(sc.Text())
		if cur == "" {
			break
		}
		if len(cfg.Variants) == 0 {
			err := Run(ctx, args, cur, append([]string{}, cfg.Env...))
			if err != nil {
				cancel()
				log.Fatal(err)
			}
			continue
		}
		for _, v := range cfg.Variants {
			Run(ctx, args, cur,
				append(
					append([]string{}, cfg.Env...),
					v.Env...,
				),
			)
		}
	}

	if sc.Err() != nil {
		log.Fatal(sc.Err())
	}
}
