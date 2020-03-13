package sendit

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
)

func postWriter(urlStr string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()
	req, err := http.NewRequest("POST", urlStr, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Transfer-Encoding", "chunked")

	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			pr.CloseWithError(err)
			return
		}
		if resp.StatusCode != 200 {
			pr.CloseWithError(errors.New(resp.Status))
			return
		}
		pr.Close()
	}()

	return pw, nil
}
func postReader(urlStr string) (io.ReadCloser, error) {
	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Transfer-Encoding", "chunked")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New(resp.Status)
	}
	return resp.Body, nil
}

// ConnectAndServe will connect to the server at `urlStr` and start serving/routing
// traffic to the local `addr`. If ttl > 0 then connections in each direction
// will be refreshed at the specified interval.
func ConnectAndServe(urlStr, dstURLStr, token string, ttl time.Duration) error {
	dstU, err := url.Parse(dstURLStr)
	if err != nil {
		return err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	serverPrefix, routePrefix := path.Split(u.Path)
	u.Path = path.Join(serverPrefix, pathOpen)
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
	if err != nil {
		return err
	}

	v = make(url.Values)
	v.Set("token", strings.TrimSpace(tok))
	u.RawQuery = v.Encode()

	stream := NewStream()

	setPipe := func() error {
		u.Path = path.Join(serverPrefix, pathClientRead)
		reader, err := postReader(u.String())
		if err != nil {
			return err
		}

		u.Path = path.Join(serverPrefix, pathClientWrite)
		writer, err := postWriter(u.String())
		if err != nil {
			reader.Close()
			return err
		}

		err = stream.SetPipe(reader, writer)
		if err != nil {
			reader.Close()
			writer.Close()
			return err
		}

		return nil
	}
	err = setPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stream.Close()
		t := time.NewTicker(ttl)
		defer t.Stop()
		for range t.C {
			err := setPipe()
			if err != nil {
				log.Println("ERROR:", err)
				return
			}
		}
	}()

	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 3 * time.Second
	sess, err := yamux.Client(stream, cfg)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Printf("Ready; Forwarding %s -> %s", urlStr, dstURLStr)
	proxy := httputil.NewSingleHostReverseProxy(dstU)
	return http.Serve(sess, proxy)
}
