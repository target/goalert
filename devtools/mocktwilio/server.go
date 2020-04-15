package mocktwilio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	smsInCh chan *SMS

	smsCh  chan *SMS
	callCh chan *VoiceCall

	errs chan error

	cfg Config

	messages map[string]*SMS
	calls    map[string]*VoiceCall

	mux *http.ServeMux

	shutdown chan struct{}

	sidSeq uint64

	workers sync.WaitGroup
}

// NewServer creates a new Server.
func NewServer(cfg Config) *Server {
	if cfg.MinQueueTime == 0 {
		cfg.MinQueueTime = 100 * time.Millisecond
	}
	s := &Server{
		cfg:       cfg,
		callbacks: make(map[string]string),
		mux:       http.NewServeMux(),
		messages:  make(map[string]*SMS),
		calls:     make(map[string]*VoiceCall),
		smsCh:     make(chan *SMS),
		smsInCh:   make(chan *SMS),
		callCh:    make(chan *VoiceCall),
		errs:      make(chan error, 10000),
		shutdown:  make(chan struct{}),
	}

	base := "/Accounts/" + cfg.AccountSID

	s.mux.HandleFunc(base+"/Calls.json", s.serveNewCall)
	s.mux.HandleFunc(base+"/Messages.json", s.serveNewMessage)
	s.mux.HandleFunc(base+"/Calls/", s.serveCallStatus)
	s.mux.HandleFunc(base+"/Messages/", s.serveMessageStatus)

	// start 20 senders/workers
	for i := 0; i < 20; i++ {
		s.workers.Add(1)
		go s.loop()
	}

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

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 && resp.StatusCode != 204 {
		return nil, errors.Errorf("non-204 response on empty body: %s", resp.Status)
	}

	return data, nil
}

func (s *Server) processMessages() {
	s.mx.Lock()
	for _, sms := range s.messages {
		if time.Since(sms.start) < s.cfg.MinQueueTime {
			continue
		}
		switch sms.msg.Status {
		case twilio.MessageStatusAccepted:
			defer sms.updateStatus(twilio.MessageStatusQueued)
		case twilio.MessageStatusQueued:
			// move to sending once it's been pulled from the channel
			select {
			case s.smsCh <- sms:
				sms.msg.Status = twilio.MessageStatusSending
			default:
			}
		}
	}
	s.mx.Unlock()
}
func (s *Server) id(prefix string) string {
	return fmt.Sprintf("%s%032d", prefix, atomic.AddUint64(&s.sidSeq, 1))
}
func (s *Server) processCalls() {
	for _, vc := range s.calls {
		if time.Since(vc.start) < s.cfg.MinQueueTime {
			continue
		}
		switch vc.call.Status {
		case twilio.CallStatusQueued:
			vc.updateStatus(twilio.CallStatusInitiated)
		case twilio.CallStatusInitiated:
			// move to ringing once it's been pulled from the channel
			s.mx.Lock()
			select {
			case s.callCh <- vc:
				vc.call.Status = twilio.CallStatusRinging
			default:
			}
			s.mx.Unlock()
		case twilio.CallStatusInProgress:
			s.mx.Lock()
			if vc.hangup || vc.needsProcessing {
				select {
				case s.callCh <- vc:
					vc.needsProcessing = false
					if vc.hangup {
						vc.call.Status = twilio.CallStatusCompleted
					}
				default:
				}
			}
			s.mx.Unlock()
		}
	}
}

// Close will shutdown the server loop.
func (s *Server) Close() error {
	close(s.shutdown)
	s.workers.Wait()
	return nil
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
			sms.process()
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
