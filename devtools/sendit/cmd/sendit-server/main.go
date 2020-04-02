package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/target/goalert/devtools/sendit"
)

func serverWrapper(redirHTTPFwd bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if redirHTTPFwd && req.Header.Get("X-Forwarded-Proto") == "http" {
			u := *req.URL
			u.Host = req.Host
			u.Scheme = "https"
			http.Redirect(w, req, u.String(), http.StatusTemporaryRedirect)
			return
		}

		if req.URL.Path == "/" {
			// blank 200 on root path
			return
		}

		next.ServeHTTP(w, req)
	})
}

func main() {
	addr := os.Getenv("SENDIT_ADDR")
	if addr == "" {
		addr = "127.0.0.1:5050"
	}
	flag.StringVar(&addr, "addr", addr, "Local address to listen on [SENDIT_ADDR].")
	secret := flag.String("secret", os.Getenv("SENDIT_SECRET"), "Secret signing string [SENDIT_SECRET].")
	prefix := flag.String("http-prefix", os.Getenv("SENDIT_HTTP_PREFIX"), "HTTP prefix [SENDIT_HTTP_PREFIX].")
	httpsRedir := flag.Bool("https-redir", os.Getenv("SENDIT_HTTPS_REDIR") == "1", "Enable HTTP -> HTTPS redirect if X-Forwarded-Proto == http [SENDIT_HTTPS_REDIR].")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening:", l.Addr().String())
	srv := sendit.NewServer([]byte(*secret), *prefix)
	log.Fatal(http.Serve(l, serverWrapper(*httpsRedir, srv)))
}
