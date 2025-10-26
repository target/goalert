package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

var cacheDir string

func main() {
	tool := flag.String("t", "", "Tool to fetch.")
	version := flag.String("v", "", "Version of the tool to fetch.")
	output := flag.String("o", "", "Output file/dir.")
	flag.StringVar(&cacheDir, "c", os.Getenv("GETTOOL_CACHE"), "Cache dir.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if cacheDir != "" {
		// do nothing, already set
	} else if base := os.Getenv("XDG_CACHE_HOME"); base != "" {
		// use the XDG_CACHE_HOME
		cacheDir = filepath.Join(base, "goalert-gettool")
	} else if home := os.Getenv("HOME"); home != "" {
		// use the HOME dir
		cacheDir = filepath.Join(home, ".cache", "goalert-gettool")
	}

	if *tool == "" {
		log.Fatal("-t flag is required")
	}
	if *version == "" {
		log.Fatal("-v flag is required")
	}
	if *output == "" {
		log.Fatal("-o flag is required")
	}

	var err error
	switch *tool {
	case "prometheus":
		err = getPrometheus(*version, *output)
	case "protoc":
		err = getProtoC(*version, *output)
	case "sqlc":
		err = getSqlc(*version, *output)
	case "bun":
		err = getBun(*version, *output)
	case "mailpit":
		err = getMailpit(*version, *output)
	case "k6":
		err = getK6(*version, *output)
	default:
		log.Fatalf("unknown tool '%s'", *tool)
	}

	if err != nil {
		log.Fatal(err)
	}
}
