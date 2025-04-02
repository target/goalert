package mocktwilio

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/validation/validate"
)

// Config is used to configure the mock server.
type Config struct {
	// The SID and token should match values given to the backend
	// as the mock server will send and validate signatures.
	AccountSID string
	AuthToken  string

	// MinQueueTime determines the minimum amount of time an SMS or voice
	// call will sit in the queue before being processed/delivered.
	MinQueueTime time.Duration
}

// Server implements the Twilio API for SMS and Voice calls
// via the http.Handler interface.
type Server struct {
	mx        sync.RWMutex
	callbacks map[string]string

	smsInCh  chan *SMS
	callInCh chan *VoiceCall

	smsCh  chan *SMS
	callCh chan *VoiceCall

	errs chan error

	cfg Config

	messages map[string]*SMS
	calls    map[string]*VoiceCall
	msgSvc   map[string][]string
	rcs      map[string]string

	mux *http.ServeMux

	shutdown chan struct{}

	sidSeq uint64

	workers sync.WaitGroup

	carrierInfo   map[string]twilio.CarrierInfo
	carrierInfoMx sync.Mutex
}

// NewServer creates a new Server.
func NewServer(cfg Config) *Server {
	if cfg.MinQueueTime == 0 {
		cfg.MinQueueTime = 100 * time.Millisecond
	}
	s := &Server{
		cfg:         cfg,
		callbacks:   make(map[string]string),
		mux:         http.NewServeMux(),
		messages:    make(map[string]*SMS),
		calls:       make(map[string]*VoiceCall),
		msgSvc:      make(map[string][]string),
		smsCh:       make(chan *SMS),
		smsInCh:     make(chan *SMS),
		callCh:      make(chan *VoiceCall),
		callInCh:    make(chan *VoiceCall),
		errs:        make(chan error, 10000),
		shutdown:    make(chan struct{}),
		carrierInfo: make(map[string]twilio.CarrierInfo),
		rcs:         make(map[string]string),
	}

	base := "/2010-04-01/Accounts/" + cfg.AccountSID

	s.mux.HandleFunc(base+"/Calls.json", s.serveNewCall)
	s.mux.HandleFunc(base+"/Messages.json", s.serveNewMessage)
	s.mux.HandleFunc(base+"/Calls/", s.serveCallStatus)
	s.mux.HandleFunc(base+"/Messages/", s.serveMessageStatus)
	s.mux.HandleFunc("/v1/PhoneNumbers/", s.serveLookup)

	s.workers.Add(1)
	go s.loop()

	return s
}

// Errors returns a channel that gets fed all errors when calling
// the backend.
func (s *Server) Errors() chan error {
	return s.errs
}

func (s *Server) post(url string, v url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Twilio-Signature", string(twilio.Signature(s.cfg.AuthToken, url, v)))
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

func (s *Server) id(prefix string) string {
	return fmt.Sprintf("%s%032d", prefix, atomic.AddUint64(&s.sidSeq, 1))
}

// Close will shutdown the server loop.
func (s *Server) Close() {
	close(s.shutdown)
	s.workers.Wait()
}

// wait will wait the specified amount of time, but return
// true if aborted due to shutdown.
func (s *Server) wait(dur time.Duration) bool {
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-t.C:
		return false
	case <-s.shutdown:
		return true
	}
}

func (s *Server) loop() {
	defer s.workers.Done()

	for {
		select {
		case <-s.shutdown:
			return
		default:
		}

		select {
		case <-s.shutdown:
			return
		case sms := <-s.smsInCh:
			s.workers.Add(1)
			go sms.process()
		case vc := <-s.callInCh:
			s.workers.Add(1)
			go vc.process()
		}
	}
}

func apiError(status int, w http.ResponseWriter, e *twilio.Exception) {
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(e)
	if err != nil {
		panic(err)
	}
}

// ServeHTTP implements the http.Handler interface for serving [mock] API requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}

// SetCarrierInfo will set/update the carrier info (used for the Lookup API) for the given number.
func (s *Server) SetCarrierInfo(number string, info twilio.CarrierInfo) {
	s.carrierInfoMx.Lock()
	defer s.carrierInfoMx.Unlock()

	s.carrierInfo[number] = info
}

// getFromNumber will return a random number from the messaging service if ID is a
// messaging SID, or the value itself otherwise.
func (s *Server) getFromNumber(id string) string {
	if !strings.HasPrefix(id, "MG") {
		return id
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// select a random number from the message service
	if len(s.msgSvc[id]) == 0 {
		return ""
	}

	return s.msgSvc[id][rand.Intn(len(s.msgSvc[id]))]
}

// NewMessagingService registers a new Messaging SID for the given numbers.
func (s *Server) NewMessagingService(url string, numbers ...string) (string, error) {
	err := validate.URL("URL", url)
	for i, n := range numbers {
		err = validate.Many(err, validate.Phone(fmt.Sprintf("Number[%d]", i), n))
	}
	if err != nil {
		return "", err
	}
	svcID := s.id("MG")

	s.mx.Lock()
	defer s.mx.Unlock()
	for _, num := range numbers {
		s.callbacks["SMS:"+num] = url
	}
	s.msgSvc[svcID] = numbers

	return svcID, nil
}

// EnableRCS enables RCS for the given messaging service ID, returning the RCS sender ID.
func (s *Server) EnableRCS(msgSvcID string) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.msgSvc[msgSvcID]; !ok {
		return "", errors.New("messaging service not found")
	}
	seq := atomic.AddUint64(&s.sidSeq, 1)
	rcsID := fmt.Sprintf("test_%04d_agent", seq)
	s.rcs[msgSvcID] = rcsID

	url := s.msgSvc[msgSvcID][0]
	s.callbacks["SMS:"+rcsID] = url

	return rcsID, nil
}

// RegisterSMSCallback will set/update a callback URL for SMS calls made to the given number.
func (s *Server) RegisterSMSCallback(number, url string) error {
	err := validate.URL("URL", url)
	if err != nil {
		return err
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.callbacks["SMS:"+number] = url
	return nil
}

// RegisterVoiceCallback will set/update a callback URL for voice calls made to the given number.
func (s *Server) RegisterVoiceCallback(number, url string) error {
	err := validate.URL("URL", url)
	if err != nil {
		return err
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	s.callbacks["VOICE:"+number] = url
	return nil
}
