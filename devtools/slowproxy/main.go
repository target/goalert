package main

import (
	"flag"
	"io"
	"log"
	"net"
)

func main() {
	rateOut := flag.Int("o", 0, "Max data rate (in bytes/sec) to DB.")
	rateIn := flag.Int("i", 0, "Max data rate (in bytes/sec) from DB.")
	latency := flag.Duration("d", 0, "Min latency (one-way).")
	jitter := flag.Duration("j", 0, "Jitter in (random +/- to latency).")
	l := flag.String("l", "localhost:5435", "Listen address.")
	c := flag.String("c", "localhost:5432", "Connect address.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	limitOut := newRateLimiter(*rateOut, *latency, *jitter)
	limitIn := newRateLimiter(*rateIn, *latency, *jitter)

	srv, err := net.Listen("tcp", *l)
	if err != nil {
		log.Fatal(err)
	}

	proxy := func(dst, src net.Conn, limiter *rateLimiter) {
		defer dst.Close()
		defer src.Close()

		io.Copy(limiter.NewWriter(dst), src)
	}

	for {
		conn, err := srv.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			dbConn, err := net.Dial("tcp", *c)
			if err != nil {
				log.Println("connect error:", err)
				conn.Close()
				return
			}

			log.Println("CONNECT", conn.RemoteAddr().String())
			go proxy(conn, dbConn, limitOut)
			go proxy(dbConn, conn, limitIn)
		}()
	}
}
