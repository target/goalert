package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

func SetDB(input, db string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}

	u.Path = "/" + strings.TrimPrefix(db, "/")

	return u.String()
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) < 3 {
		log.Fatal("missing argument")
		os.Exit(1)
	}

	newURL := SetDB(os.Args[1], os.Args[2])
	if newURL == "" {
		log.Fatal("invalid URL")
		os.Exit(1)
	}

	fmt.Println(newURL)
}
