package mocktwilio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/twilio"
	"github.com/ttacon/libphonenumber"
)

// Config is used to configure the mock server.
type Config struct {
	// AccountSID is the Twilio account SID.
	AccountSID string

	// AuthToken is the Twilio auth token.
	AuthToken string

	Numbers     []Number
	MsgServices []MsgService

	// If EnableAuth is true, incoming requests will need to have a valid Authorization header.
	EnableAuth bool

	OnError func(context.Context, error)
}

// Number represents a mock phone number.
type Number struct {
	Number string

	VoiceWebhookURL string
	SMSWebhookURL   string
}

// MsgService allows configuring a mock messaging service that can rotate between available numbers.
type MsgService struct {
	// SID is the messaging service ID, it must start with 'MG'.
	SID string

	Numbers []string

	// SMSWebhookURL is the URL to which SMS messages will be sent.
	//
	// It takes precedence over the SMSWebhookURL field in the Config.Numbers field
	// for all numbers in the service.
	SMSWebhookURL string
}

// Server implements the Twilio API for SMS and Voice calls
// via the http.Handler interface.
type Server struct {
	cfg Config

	smsDB         chan map[string]*sms
	messagesCh    chan Message
	outboundSMSCh chan *sms

	numbers map[string]*Number
	msgSvc  map[string][]*Number

	mux *http.ServeMux

	once         sync.Once
	shutdown     chan struct{}
	shutdownDone chan struct{}

	id uint64

	workers sync.WaitGroup

	carrierInfo   map[string]twilio.CarrierInfo
	carrierInfoMx sync.Mutex
}

func validateURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		return errors.Errorf("invalid URL (missing scheme): %s", s)
	}

	return nil
}

// NewServer creates a new Server.
func NewServer(cfg Config) (*Server, error) {
	if cfg.AccountSID == "" {
		return nil, errors.New("AccountSID is required")
	}

	srv := &Server{
		cfg:     cfg,
		msgSvc:  make(map[string][]*Number),
		numbers: make(map[string]*Number),
		mux:     http.NewServeMux(),

		smsDB:         make(chan map[string]*sms, 1),
		messagesCh:    make(chan Message),
		outboundSMSCh: make(chan *sms),

		shutdown:     make(chan struct{}),
		shutdownDone: make(chan struct{}),
	}

	for _, n := range cfg.Numbers {
		_, err := libphonenumber.Parse(n.Number, "")
		if err != nil {
			return nil, fmt.Errorf("invalid phone number %s: %v", n.Number, err)
		}
		if n.SMSWebhookURL != "" {
			err = validateURL(n.SMSWebhookURL)
			if err != nil {
				return nil, err
			}
		}
		if n.VoiceWebhookURL != "" {
			err = validateURL(n.VoiceWebhookURL)
			if err != nil {
				return nil, err
			}
		}

		// point to copy
		_n := n
		srv.numbers[n.Number] = &_n
	}
	for _, m := range cfg.MsgServices {
		if !strings.HasPrefix(m.SID, "MG") {
			return nil, fmt.Errorf("invalid MsgService SID %s", m.SID)
		}

		if m.SMSWebhookURL != "" {
			err := validateURL(m.SMSWebhookURL)
			if err != nil {
				return nil, err
			}
		}

		for _, nStr := range m.Numbers {
			_, err := libphonenumber.Parse(nStr, "")
			if err != nil {
				return nil, fmt.Errorf("invalid phone number %s: %v", nStr, err)
			}

			n := srv.numbers[nStr]
			if n == nil {
				n = &Number{Number: nStr}
				srv.numbers[nStr] = n
			}

			if m.SMSWebhookURL == "" {
				continue
			}

			n.SMSWebhookURL = m.SMSWebhookURL
		}
	}

	srv.mux.HandleFunc(srv.basePath()+"/Messages.json", srv.HandleNewMessage)
	srv.mux.HandleFunc(srv.basePath()+"/Messages/", srv.HandleMessageStatus)
	// s.mux.HandleFunc(base+"/Calls.json", s.serveNewCall)
	// s.mux.HandleFunc(base+"/Calls/", s.serveCallStatus)
	// s.mux.HandleFunc("/v1/PhoneNumbers/", s.serveLookup)

	go srv.loop()

	return srv, nil
}

func (srv *Server) basePath() string {
	return "/2010-04-01/Accounts/" + srv.cfg.AccountSID
}

func (srv *Server) nextID(prefix string) string {
	return fmt.Sprintf("%s%032d", prefix, atomic.AddUint64(&srv.id, 1))
}

func (srv *Server) logErr(ctx context.Context, err error) {
	if srv.cfg.OnError == nil {
		return
	}

	srv.cfg.OnError(ctx, err)
}

func (srv *Server) Close() error {
	srv.once.Do(func() {
		close(srv.shutdown)
	})

	<-srv.shutdownDone
	return nil
}

func (srv *Server) loop() {
	var wg sync.WaitGroup

	defer close(srv.shutdownDone)
	defer close(srv.messagesCh)
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-srv.shutdown:

			return
		case sms := <-srv.outboundSMSCh:
			wg.Add(1)
			go func() {
				sms.lifecycle(ctx)
				wg.Done()
			}()
		}
	}
}

func (s *Server) post(ctx context.Context, url string, v url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", Signature(s.cfg.AuthToken, url, v))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.Errorf("non-2xx response: %s", resp.Status)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 && resp.StatusCode != 204 {
		return nil, errors.Errorf("non-204 response on empty body: %s", resp.Status)
	}

	return data, nil
}

// ServeHTTP implements the http.Handler interface for serving [mock] API requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if s.cfg.EnableAuth {
		user, pass, ok := req.BasicAuth()
		if !ok || user != s.cfg.AccountSID || pass != s.cfg.AuthToken {
			respondErr(w, twError{
				Status:  401,
				Code:    20003,
				Message: "Authenticate",
			})
			return
		}
	}

	s.mux.ServeHTTP(w, req)
}
