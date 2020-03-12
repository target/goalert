package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	addr := flag.String("addr", ":3040", "Address to listen for HTTP traffic.")
	trim := flag.Bool("trim", false, "Trim matching URL path before forwarding request.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)
	mux := http.NewServeMux()
	for _, route := range flag.Args() {
		parts := strings.SplitN(route, "=", 2)
		if len(parts) == 1 {
			parts = []string{"/", parts[0]}
		}

		u, err := url.Parse(parts[1])
		if err != nil {
			log.Fatalf("ERORR: parse %s: %v", parts[1], err)
		}

		p := httputil.NewSingleHostReverseProxy(u)
		h := http.Handler(p)
		if *trim {
			h = http.StripPrefix(parts[0], h)
		}
		mux.Handle(parts[0], h)
		log.Printf("Registered: %s -> %s", parts[0], parts[1])
	}

	log.Println("Listening:", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		log.Println("ERROR: serve:", err)
	}
}
