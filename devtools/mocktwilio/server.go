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
	// ID is the messaging service SID, it must start with 'MG'.
	ID string

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

	numbersDB chan map[string]*Number
	msgSvcDB  chan map[string][]*Number

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
func NewServer(cfg Config) *Server {
	if cfg.AccountSID == "" {
		panic("AccountSID is required")
	}

	srv := &Server{
		cfg:       cfg,
		msgSvcDB:  make(chan map[string][]*Number, 1),
		numbersDB: make(chan map[string]*Number, 1),
		mux:       http.NewServeMux(),

		smsDB:         make(chan map[string]*sms, 1),
		messagesCh:    make(chan Message),
		outboundSMSCh: make(chan *sms),

		shutdown:     make(chan struct{}),
		shutdownDone: make(chan struct{}),
	}
	srv.msgSvcDB <- make(map[string][]*Number)
	srv.numbersDB <- make(map[string]*Number)

	srv.mux.HandleFunc(srv.basePath()+"/Messages.json", srv.HandleNewMessage)
	srv.mux.HandleFunc(srv.basePath()+"/Messages/", srv.HandleMessageStatus)
	// s.mux.HandleFunc(base+"/Calls.json", s.serveNewCall)
	// s.mux.HandleFunc(base+"/Calls/", s.serveCallStatus)
	// s.mux.HandleFunc("/v1/PhoneNumbers/", s.serveLookup)

	go srv.loop()

	return srv
}

func (srv *Server) number(s string) *Number {
	db := <-srv.numbersDB
	n := db[s]
	srv.numbersDB <- db
	return n
}

func (srv *Server) numberSvc(id string) []*Number {
	db := <-srv.msgSvcDB
	nums := db[id]
	srv.msgSvcDB <- db

	return nums
}

// AddNumber adds a new number to the mock server.
func (srv *Server) AddNumber(n Number) error {
	_, err := libphonenumber.Parse(n.Number, "")
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %v", n.Number, err)
	}
	if n.SMSWebhookURL != "" {
		err = validateURL(n.SMSWebhookURL)
		if err != nil {
			return err
		}
	}
	if n.VoiceWebhookURL != "" {
		err = validateURL(n.VoiceWebhookURL)
		if err != nil {
			return err
		}
	}

	db := <-srv.numbersDB
	if _, ok := db[n.Number]; ok {
		srv.numbersDB <- db
		return fmt.Errorf("number %s already exists", n.Number)
	}
	db[n.Number] = &n
	srv.numbersDB <- db
	return nil
}

// AddMsgService adds a new messaging service to the mock server.
func (srv *Server) AddMsgService(ms MsgService) error {
	if !strings.HasPrefix(ms.ID, "MG") {
		return fmt.Errorf("invalid MsgService SID %s", ms.ID)
	}

	if ms.SMSWebhookURL != "" {
		err := validateURL(ms.SMSWebhookURL)
		if err != nil {
			return err
		}
	}
	for _, nStr := range ms.Numbers {
		_, err := libphonenumber.Parse(nStr, "")
		if err != nil {
			return fmt.Errorf("invalid phone number %s: %v", nStr, err)
		}
	}

	msDB := <-srv.msgSvcDB
	if _, ok := msDB[ms.ID]; ok {
		srv.msgSvcDB <- msDB
		return fmt.Errorf("MsgService SID %s already exists", ms.ID)
	}

	numDB := <-srv.numbersDB
	for _, nStr := range ms.Numbers {
		n := numDB[nStr]
		if n == nil {
			n = &Number{Number: nStr}
			numDB[nStr] = n
		}
		msDB[ms.ID] = append(msDB[ms.ID], n)

		if ms.SMSWebhookURL == "" {
			continue
		}

		n.SMSWebhookURL = ms.SMSWebhookURL
	}
	srv.numbersDB <- numDB
	srv.msgSvcDB <- msDB

	return nil
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

// Close shuts down the server.
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
