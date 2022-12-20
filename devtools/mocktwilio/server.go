package mocktwilio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
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

	msgCh         chan Message
	msgStateDB    chan map[string]*msgState
	outboundMsgCh chan *msgState

	callCh         chan Call
	callStateDB    chan map[string]*callState
	outboundCallCh chan *callState

	numInfoCh chan map[string]*CarrierInfo

	numbersDB chan *numberDB

	waitInFlight chan chan struct{}

	mux *http.ServeMux

	once         sync.Once
	shutdown     chan struct{}
	shutdownDone chan struct{}

	id uint64
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
		numbersDB: make(chan *numberDB, 1),
		mux:       http.NewServeMux(),

		msgCh:         make(chan Message, 10000),
		msgStateDB:    make(chan map[string]*msgState, 1),
		outboundMsgCh: make(chan *msgState),

		callCh:         make(chan Call, 10000),
		callStateDB:    make(chan map[string]*callState, 1),
		outboundCallCh: make(chan *callState),

		shutdown:     make(chan struct{}),
		shutdownDone: make(chan struct{}),

		waitInFlight: make(chan chan struct{}),

		numInfoCh: make(chan map[string]*CarrierInfo, 1),
	}
	srv.numbersDB <- newNumberDB()
	srv.msgStateDB <- make(map[string]*msgState)
	srv.callStateDB <- make(map[string]*callState)
	srv.numInfoCh <- make(map[string]*CarrierInfo)

	srv.initHTTP()

	go srv.loop()

	return srv
}

func (srv *Server) SetCarrierInfo(number string, info CarrierInfo) {
	db := <-srv.numInfoCh
	db[number] = &info
	srv.numInfoCh <- db
}

func (srv *Server) number(s string) *Number {
	db := <-srv.numbersDB
	if !db.NumberExists(s) {
		srv.numbersDB <- db
		return nil
	}

	n := &Number{
		Number:          s,
		VoiceWebhookURL: db.VoiceWebhookURL(s),
		SMSWebhookURL:   db.SMSWebhookURL(s),
	}
	srv.numbersDB <- db
	return n
}

func (srv *Server) svcNumbers(id string) []string {
	db := <-srv.numbersDB
	svc := db.MsgSvcNumbers(id)
	srv.numbersDB <- db
	return svc
}

// AddUpdateNumber adds or updates a number.
func (srv *Server) AddUpdateNumber(n Number) error {
	db := <-srv.numbersDB
	err := db.AddUpdateNumber(n)
	srv.numbersDB <- db

	return err
}

// AddUpdateMsgService or updates a messaging service.
func (srv *Server) AddUpdateMsgService(ms MsgService) error {
	db := <-srv.numbersDB
	err := db.AddUpdateMsgService(ms)
	srv.numbersDB <- db
	return err
}

func (srv *Server) nextID(prefix string) string {
	return fmt.Sprintf("%s%032d", prefix, atomic.AddUint64(&srv.id, 1))
}

func (srv *Server) logErr(ctx context.Context, err error) {
	if err == nil {
		return
	}
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
	defer close(srv.msgCh)
	defer close(srv.callCh)
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-srv.shutdown:
			return
		case sms := <-srv.outboundMsgCh:
			wg.Add(1)
			go func() {
				sms.lifecycle(ctx)
				wg.Done()
			}()
		case call := <-srv.outboundCallCh:
			wg.Add(1)
			go func() {
				call.lifecycle(ctx)
				wg.Done()
			}()
		case ch := <-srv.waitInFlight:
			go func() {
				wg.Wait()
				close(ch)
			}()
		}
	}
}

// WaitInFlight waits for all in-flight requests/messages/calls to complete.
func (srv *Server) WaitInFlight(ctx context.Context) error {
	ch := make(chan struct{})
	select {
	case srv.waitInFlight <- ch:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-ch:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
