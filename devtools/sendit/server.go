package sendit

import (
	"context"
	"crypto/rand"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"sync"
)

const (
	pathOpen        = "/.well-known/sendit/v1/open"
	pathClientRead  = "/.well-known/sendit/v1/read"
	pathClientWrite = "/.well-known/sendit/v1/write"
)

type serverContextValue int

const (
	serverContextValueHost = serverContextValue(1)
)

// Server is an http.Handler that can multiplex reverse-proxied requests
// over 2 POST requests from a client.
type Server struct {
	http.Handler
	proxy *httputil.ReverseProxy

	sessionsByID     map[string]*session
	sessionsByPrefix map[string]*session

	mx sync.RWMutex

	authSecret    []byte
	connectSecret []byte

	prefix string
}

// NewServer will create a new Server with a global pathPrefix and using
// the provided `authSecret` to require valid tokens for any connecting clients.
func NewServer(authSecret []byte, prefix string) *Server {
	prefix = path.Join("/", prefix, "/")

	if prefix != "/" {
		prefix += "/"
	}
	mux := http.NewServeMux()
	s := &Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("REQUEST: %s %s", req.Method, req.URL.Path)
			mux.ServeHTTP(w, req)
		}),
		sessionsByPrefix: make(map[string]*session),
		sessionsByID:     make(map[string]*session),
		authSecret:       authSecret,
		connectSecret:    make([]byte, 32),
		prefix:           prefix,
	}

	_, err := rand.Read(s.connectSecret)
	if err != nil {
		panic(err)
	}
	mux.HandleFunc(path.Join(prefix, pathOpen), s.serveOpen)
	mux.HandleFunc(path.Join(prefix, pathClientRead), s.serveClientRead)
	mux.HandleFunc(path.Join(prefix, pathClientWrite), s.serveClientWrite)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = s.DialContext
	transport.DialTLS = nil
	transport.Dial = nil

	s.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = req.Context().Value(serverContextValueHost).(string)

			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		Transport: transport,
	}

	mux.HandleFunc(prefix, s.servePrefix)
	if prefix != "/" {
		mux.HandleFunc(strings.TrimSuffix(prefix, "/"), s.servePrefix)
	}
	return s
}

func (s *Server) sessionByPrefix(prefix string) *session {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.sessionsByPrefix[prefix]
}
func (s *Server) sessionByID(id string) *session {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.sessionsByID[id]
}

func (s *Server) servePrefix(w http.ResponseWriter, req *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(req.URL.Path, s.prefix), "/", 2)
	if len(parts) < 2 {
		http.NotFound(w, req)
		return
	}

	sess := s.sessionByPrefix(parts[0])
	if sess == nil {
		http.NotFound(w, req)
	}

	s.proxy.ServeHTTP(w, req.WithContext(
		context.WithValue(req.Context(), serverContextValueHost, sess.ID),
	))
}

// DialContext will return a new `net.Conn` for the given `addr`. `network` is ignored.
//
// It will only connect to "hosts" that match an active session ID.
func (s *Server) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	sess := s.sessionByID(host)
	if sess == nil {
		return nil, &net.AddrError{Addr: addr, Err: "unknown host"}
	}

	return sess.OpenContext(ctx)
}

// serveOpen will perform initial authentication and esablish a tunnel session.
func (s *Server) serveOpen(w http.ResponseWriter, req *http.Request) {
	sess, err := s.newSession(req.FormValue("prefix"))
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := GenerateToken(s.connectSecret, TokenAudienceConnect, sess.ID)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "private, no-cache, no-store")
	_, err = io.WriteString(w, token+"\n")
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	w.(http.Flusher).Flush()

	select {
	case <-req.Context().Done():
		log.Println("ERROR: request terminated before session established")
		sess.End()
		return
	case <-sess.stream.Ready():
	}

	// Session initiated.
}

// serveOpen will perform initial authentication and esablish a tunnel session.
func (s *Server) serveClientRead(w http.ResponseWriter, req *http.Request) {
	id, err := TokenSubject(s.connectSecret, TokenAudienceConnect, req.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sess := s.sessionByID(id)
	if sess == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "private, no-cache, no-store")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	sess.UseWriter(req.Context(), FlushWriter(w))
}

// serveClientWrite handles connecting.
func (s *Server) serveClientWrite(w http.ResponseWriter, req *http.Request) {
	id, err := TokenSubject(s.connectSecret, TokenAudienceConnect, req.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sess := s.sessionByID(id)
	if sess == nil {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "private, no-cache, no-store")

	sess.UseReader(req.Context(), req.Body)
}
