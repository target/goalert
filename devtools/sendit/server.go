package sendit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
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

	pending  map[string]*pendingSession
	routes   map[string]string
	sessions map[string]*yamux.Session

	routeMx sync.RWMutex
	sessMx  sync.Mutex

	authSecret    []byte
	connectSecret []byte

	prefix string
}

type pendingSession struct {
	context.Context
	ch chan sessionReader
}

type sessionReader struct {
	io.Reader
	io.Closer
	context.Context
}

type listenerSession struct {
	ctx    context.Context
	cancel func()

	w     io.Writer
	flush func()
}

func genID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

const (
	readPath  = "/.well-known/sendit/v1/read"
	writePath = "/.well-known/sendit/v1/write"
)

// NewServer will create a new Server with a global pathPrefix and using
// the provided `authSecret` to require valid tokens for any connecting clients.
func NewServer(authSecret []byte, prefix string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		Handler:       mux,
		pending:       make(map[string]*pendingSession),
		routes:        make(map[string]string),
		sessions:      make(map[string]*yamux.Session),
		authSecret:    authSecret,
		connectSecret: make([]byte, 32),
	}
	_, err := rand.Read(s.connectSecret)
	if err != nil {
		panic(err)
	}
	mux.HandleFunc(path.Join(prefix, readPath), s.serveRead)
	mux.HandleFunc(path.Join(prefix, writePath), s.serveWrite)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = nil
	transport.DialTLS = nil
	transport.Dial = s.Dial

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

	mux.HandleFunc(path.Join(prefix, "/"), s.servePrefix)
	return s
}

func (s *Server) servePrefix(w http.ResponseWriter, req *http.Request) {
	parts := strings.SplitN(req.URL.Path, "/", 3)
	if len(parts) < 2 {
		http.NotFound(w, req)
		return
	}

	s.routeMx.RLock()
	hostID := s.routes[parts[1]]
	s.routeMx.RUnlock()

	if hostID == "" {
		http.NotFound(w, req)
	}

	s.proxy.ServeHTTP(w, req.WithContext(
		context.WithValue(req.Context(), serverContextValueHost, hostID),
	))
}

// Dial will return a new `net.Conn` for the given `addr`. `network` is ignored.
//
// It will only connect to "hosts" that match an active session ID.
func (s *Server) Dial(network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	s.routeMx.RLock()
	sess := s.sessions[host]
	s.routeMx.RUnlock()

	if sess == nil {
		return nil, &net.AddrError{Addr: addr, Err: "unknown host"}
	}

	return sess.Open()
}

// ValidPath will return true if `p` is a valid prefix path.
func ValidPath(p string) bool {
	if len(p) < 3 {
		return false
	}
	if len(p) > 64 {
		return false
	}
	for _, r := range p {
		if r == '-' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		return false
	}

	return true
}

func (s *Server) reserveRoutePrefix(prefix string) bool {
	s.routeMx.Lock()
	defer s.routeMx.Unlock()
	_, hasRoute := s.routes[prefix]
	if hasRoute {
		return false
	}
	// reserve the route, but don't assign host ID yet
	// this way we can't have multiple registrations in-flight
	// but still return 404 until the session is established.
	s.routes[prefix] = ""

	return true
}

func (s *Server) cleanupPrefix(prefix, id string) {
	s.routeMx.Lock()
	defer s.routeMx.Unlock()

	delete(s.routes, prefix)
	if id != "" {
		delete(s.sessions, id)
	}
}

func (s *Server) getSession(ctx context.Context, w io.Writer, prefix, id string) (context.Context, *yamux.Session, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tok, err := GenerateToken(s.connectSecret, TokenAudienceConnect, id)
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan sessionReader, 1)
	s.sessMx.Lock()
	s.pending[id] = &pendingSession{Context: ctx, ch: ch}
	s.sessMx.Unlock()
	defer func() {
		s.sessMx.Lock()
		delete(s.pending, id)
		s.sessMx.Unlock()
	}()

	w = FlushWriter(w)
	_, err = io.WriteString(w, tok+"\n")
	if err != nil {
		return nil, nil, err
	}

	var rwc struct {
		sessionReader
		io.Writer
	}
	rwc.Writer = w
	select {
	case rwc.sessionReader = <-ch:
	case <-waitCtx.Done():
		return nil, nil, waitCtx.Err()
	}

	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 3 * time.Second
	sess, err := yamux.Server(rwc, cfg)
	if err != nil {
		return nil, nil, err
	}

	s.routeMx.Lock()
	s.routes[prefix] = id
	s.sessions[id] = sess
	s.routeMx.Unlock()

	return rwc.Context, sess, nil
}

// serveRead will respond with a fixed-length "secret" that
// can be passed to the `/write` endpoint to establish a tunnel.
func (s *Server) serveRead(w http.ResponseWriter, req *http.Request) {
	if len(s.authSecret) > 0 {
		token := req.FormValue("token")
		_, err := TokenSubject(s.authSecret, TokenAudienceAuth, token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	prefix := req.FormValue("prefix")
	if !ValidPath(prefix) {
		http.Error(w, "invalid or missing prefix value", http.StatusBadRequest)
		return
	}
	if !s.reserveRoutePrefix(prefix) {
		http.Error(w, "prefix already active", http.StatusConflict)
		return
	}
	defer s.cleanupPrefix(prefix, "")

	id, err := genID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("NEW session. prefix=%s client=%s id=%s", prefix, req.RemoteAddr, id)
	defer log.Printf("END session. prefix=%s client=%s id=%s", prefix, req.RemoteAddr, id)

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "private, no-cache, no-store")

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	sessCtx, sess, err := s.getSession(ctx, w, prefix, id)
	if err != nil {
		log.Println("ERROR: getSession:", err)
		return
	}
	defer sess.Close()
	defer s.cleanupPrefix(prefix, id)

	log.Printf("ACTIVE session. prefix=%s client=%s id=%s", prefix, req.RemoteAddr, id)

	// wait for either end to hangup
	select {
	case <-sessCtx.Done():
	case <-ctx.Done():
	}
}

func (s *Server) getPendingSession(id string) *pendingSession {
	s.sessMx.Lock()
	defer s.sessMx.Unlock()
	defer delete(s.pending, id)
	return s.pending[id]
}

// serveWrite will read a fixed-length "secret" from the request body
// and establish the other side of a tunnel.
//
// If successful, the tunnel will then be passed/provided to the next
// Accept call.
func (s *Server) serveWrite(w http.ResponseWriter, req *http.Request) {
	tok := req.URL.Query().Get("token")
	id, err := TokenSubject(s.connectSecret, TokenAudienceConnect, tok)
	if err != nil {
		log.Println("ERROR: verify token:", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	p := s.getPendingSession(id)
	if p == nil {
		http.Error(w, "invalid session ID", http.StatusForbidden)
		return
	}

	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	sess := sessionReader{
		Reader:  req.Body,
		Closer:  req.Body,
		Context: ctx,
	}
	p.ch <- sess

	// wait for either end to hangup
	select {
	case <-sess.Done():
	case <-p.Done():
	}
}
