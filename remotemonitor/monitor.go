package remotemonitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/retry"
)

// Monitor will check for functionality and communication between itself and one or more instances.
// Each monitor should have a unique phone number and location.
type Monitor struct {
	appCfg     config.Config
	cfg        Config
	tw         twilio.Config
	shutdownCh chan struct{}
	startCh    chan string
	finishCh   chan string
	pendingCh  chan int
	pending    map[string]time.Time
	srv        *http.Server
}

func setRequestScheme(scheme string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// required for Twilio sig validation to work
		req.URL.Host = req.Host
		req.URL.Scheme = scheme

		h.ServeHTTP(w, req)
	})
}

// NewMonitor creates and starts a new Monitor with the given Config.
func NewMonitor(cfg Config) (*Monitor, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	http.DefaultTransport = &requestIDTransport{
		RoundTripper: http.DefaultTransport,
	}
	u, err := url.Parse(cfg.PublicURL)
	if err != nil {
		return nil, err
	}
	m := &Monitor{
		cfg:        cfg,
		tw:         twilio.Config{},
		shutdownCh: make(chan struct{}),
		startCh:    make(chan string),
		finishCh:   make(chan string),
		pendingCh:  make(chan int),
		pending:    make(map[string]time.Time),
	}
	l, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return nil, err
	}
	h := twilio.WrapValidation(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, u.Path)

		m.ServeHTTP(w, req)
	}), m.tw)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.WriteString(w, "ok")
	})
	m.appCfg.General.PublicURL = cfg.PublicURL
	m.appCfg.Twilio.Enable = true
	m.appCfg.Twilio.AccountSID = cfg.Twilio.AccountSID
	m.appCfg.Twilio.AuthToken = cfg.Twilio.AuthToken
	m.appCfg.Twilio.FromNumber = cfg.Twilio.FromNumber
	m.appCfg.Twilio.MessagingServiceSID = cfg.Twilio.MessageSID
	mux.Handle("/", twilio.WrapHeaderHack(h))
	m.srv = &http.Server{
		Handler: config.Handler(
			setRequestScheme(u.Scheme, mux),
			config.Static(m.appCfg),
		),
		IdleTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		MaxHeaderBytes:    1024 * 1024,
	}

	m.srv.SetKeepAlivesEnabled(false)

	log.Println("Listening:", l.Addr())

	go m.serve(l)
	go m.loop()
	go m.waitLoop()

	return m, nil
}

func (m *Monitor) serve(l net.Listener) {
	err := m.srv.Serve(l)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln("ERROR:", err)
	}
}

func (m *Monitor) reportErr(i Instance, err error, action string) {
	if err == nil {
		return
	}
	summary := fmt.Sprintf("Remote Monitor in %s failed to %s in %s", m.cfg.Location, action, i.Location)
	details := fmt.Sprintf("Monitor Location: %s\nInstance Location: %s\nAction: %s\nError: %s", m.cfg.Location, i.Location, action, err.Error())
	for _, ins := range m.cfg.Instances {
		if ins.ErrorAPIKey == "" {
			log.Println("No ErrorAPIKey for", ins.Location)
			continue
		}
		ins := ins // copy
		go func() {
			if err := ins.createGenericAlert(ins.ErrorAPIKey, "", summary, details); err != nil {
				log.Printf("ERROR: create generic alert: %v", err)
			}
		}()
	}
	log.Println("ERROR:", summary)
}

func (m *Monitor) waitLoop() {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-t.C:
			for k, v := range m.pending {
				if time.Since(v) > time.Minute {
					delete(m.pending, k)
				}
			}
		case name := <-m.startCh:
			m.pending[name] = time.Now()
		case name := <-m.finishCh:
			delete(m.pending, name)
		}

		select {
		case m.pendingCh <- len(m.pending):
		default:
		}
	}
}

func (m *Monitor) createEmailAlert(i Instance, dedup, summary, details string) error {
	addr, err := mail.ParseAddress(i.EmailAPIKey)
	if err != nil {
		return err
	}
	key, domain, found := strings.Cut(addr.Address, "@")
	if !found {
		return fmt.Errorf("invalid email address: %s", i.EmailAPIKey)
	}
	addr.Address = key + "+" + dedup + "@" + domain

	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.cfg.SMTP.From)
	msg += fmt.Sprintf("To: %s\r\n", addr.String())
	msg += fmt.Sprintf("Subject: %s\r\n", summary)
	msg += fmt.Sprintf("\r\n%s\r\n", details)

	host, _, err := net.SplitHostPort(m.cfg.SMTP.ServerAddr)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if m.cfg.SMTP.User != "" || m.cfg.SMTP.Pass != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTP.User, m.cfg.SMTP.Pass, host)
	}

	err = retry.DoTemporaryError(func(_ int) error {
		err = smtp.SendMail(m.cfg.SMTP.ServerAddr, auth, m.cfg.SMTP.From, []string{addr.Address}, []byte(msg))
		err = errors.Wrap(err, "send email")
		return err
	},
		retry.Log(m.context()),
		retry.Limit(m.cfg.SMTP.Retries),
		retry.FibBackoff(time.Second),
	)

	return err
}

func (m *Monitor) loop() {
	delay := time.Duration(m.cfg.CheckMinutes) * time.Minute
	t := time.NewTicker(delay)

	dedup := fmt.Sprintf("RM-%s", m.cfg.Location)
	summary := fmt.Sprintf("Remote Monitor Communication Test from %s", m.cfg.Location)
	details := fmt.Sprintf(`This alert was generated by a GoAlert Remote Monitor running in %s.

These alerts are generated periodically to monitor actual system functionality and communication.

If it is not automatically closed within a minute, there may be a problem with SMS or network connectivity.
`, m.cfg.Location)

	doCheck := func(preferEmail bool) {
		for _, i := range m.cfg.Instances {
			if i.ErrorsOnly {
				continue
			}
			m.startCh <- i.Location
			go func(i Instance) {
				var err error
				switch {
				case i.EmailAPIKey != "" && (i.GenericAPIKey == "" || preferEmail):
					err = m.createEmailAlert(i, dedup, summary, details)
				case i.GenericAPIKey != "" && (i.EmailAPIKey == "" || !preferEmail):
					err = i.createGenericAlert(i.GenericAPIKey, dedup, summary, details)
				}

				if err != nil {
					m.reportErr(i, err, "create new alert")
				}
			}(i)
		}
	}

	var preferEmail bool
	doCheck(preferEmail)
	for {
		select {
		case <-m.shutdownCh:
			return
		case <-t.C:
			preferEmail = !preferEmail
			doCheck(preferEmail)
		}
	}
}

// context will return a new background context with config applied.
func (m *Monitor) context() context.Context {
	return m.appCfg.Context(context.Background())
}

// Shutdown gracefully shuts down the monitor, waiting for any in-flight checks to complete.
func (m *Monitor) Shutdown(ctx context.Context) error {
	log.Println("Beginning shutdown...")
	close(m.shutdownCh)

wait:
	for {
		select {
		case n := <-m.pendingCh:
			if n == 0 {
				// wait for all pending operations to finish or timeout
				break wait
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return m.srv.Shutdown(ctx)
}
