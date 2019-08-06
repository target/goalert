package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	listen, connect, drop string

	dropFirst bool
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&listen, "listen", "localhost:5434", "Listen address.")
	flag.StringVar(&connect, "connect", "localhost:5432", "Destination address.")
	flag.StringVar(&drop, "drop", "", "Drops all future traffic from one side of a connection if the string is found.")
	flag.BoolVar(&dropFirst, "drop-first-only", false, "Only processes the drop rule once.")
	flag.Parse()

	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening:", l.Addr().String())

	var n int
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		n++
		mx.Lock()
		fmt.Printf("#%04d: CONN :: %s -> %s\n", n, c.RemoteAddr().String(), c.LocalAddr().String())
		mx.Unlock()
		go handleConn(n, c)
	}
}
func handleConn(n int, c net.Conn) {
	pg, err := net.Dial("tcp", connect)
	if err != nil {
		log.Println("ERROR:", err)
		c.Close()
		return
	}

	go pipe(n, pg, c, "RECV")
	go pipe(n, c, pg, "SEND")
}

func pipe(n int, dst, src net.Conn, dir string) {
	io.Copy(&prefixWriter{prefix: fmt.Sprintf("#%04d: %s", n, dir), Writer: dst}, src)
	dst.Close()
	src.Close()
}
