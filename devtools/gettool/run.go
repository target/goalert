package main

import (
	"flag"
	"log"
)

func main() {
	tool := flag.String("t", "", "Tool to fetch.")
	version := flag.String("v", "", "Version of the tool to fetch.")
	output := flag.String("o", "", "Output file/dir.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

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
	case "golangci-lint":
		err = getGolangCiLint(*version, *output)
	case "sqlc":
		err = getSqlc(*version, *output)
	case "mailpit":
		err = getMailpit(*version, *output)
	default:
		log.Fatalf("unknown tool '%s'", *tool)
	}

	if err != nil {
		log.Fatal(err)
	}
}
