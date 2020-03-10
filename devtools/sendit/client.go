package sendit

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
)

func forward(a, b net.Conn) {
	defer a.Close()
	defer b.Close()
	io.Copy(a, b)
}

// ConnectAndServe will connect to the server at `urlStr` and start serving/routing
// traffic to the local `addr`.
func ConnectAndServe(urlStr, addr, token string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if u.Host == addr {
		return errors.New("connect address must not match URL host")
	}
	serverPrefix, routePrefix := path.Split(u.Path)
	u.Path = path.Join(serverPrefix, readPath)
	v := make(url.Values)
	v.Set("prefix", routePrefix)
	v.Set("token", strings.TrimSpace(token))
	resp, err := http.PostForm(u.String(), v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Server Error: " + resp.Status)
	}

	r := bufio.NewReader(resp.Body)
	tok, err := r.ReadString('\n')

	pipeRead, pipeWrite := io.Pipe()

	go func() {
		defer pipeWrite.Close()
		var rwc struct {
			io.Reader
			io.WriteCloser
		}
		rwc.WriteCloser = pipeWrite
		rwc.Reader = r

		cfg := yamux.DefaultConfig()
		cfg.KeepAliveInterval = 3 * time.Second
		sess, err := yamux.Client(rwc, cfg)
		if err != nil {
			log.Println("ERROR: establish session:", err)
			return
		}
		defer sess.Close()

		log.Printf("Routing '%s' -> %s", urlStr, addr)

		for {
			remoteConn, err := sess.Accept()
			if err != nil {
				log.Println("ERROR: accept connection:", err)
				return
			}

			localConn, err := net.Dial("tcp", addr)
			if err != nil {
				log.Println("ERROR: connect:", err)
				remoteConn.Close()
				continue
			}
			go forward(remoteConn, localConn)
			go forward(localConn, remoteConn)
		}
	}()

	u.Path = path.Join(serverPrefix, writePath)

	q := u.Query()
	q.Set("token", strings.TrimSpace(tok))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), pipeRead)
	if err != nil {
		return err
	}
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Content-Type", "application/octet-stream")

	_, err = http.DefaultClient.Do(req)
	return err
}
