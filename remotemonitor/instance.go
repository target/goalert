package remotemonitor

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/target/goalert/util"
)

// An Instance represents a running remote GoAlert instance to monitor.
type Instance struct {
	// Location must be unique.
	Location string

	// GenericAPIKey is used to create test alerts.
	// The service it points to should have an escalation policy that allows at least 60 seconds
	// before escalating to a human. It should send initial notifications to the monitor via SMS.
	GenericAPIKey string

	// EmailAPIKey is used to create test alerts.
	// The service it points to should have an escalation policy that allows at least 60 seconds
	// before escalating to a human. It should send initial notifications to the monitor via SMS.
	EmailAPIKey string

	// ErrorAPIKey is the key used to create new alerts for encountered errors.
	ErrorAPIKey string

	// HeartbeatURLs are sent a POST request after a successful test cycle for this instance.
	HeartbeatURLs []string

	// PublicURL should point to the publicly-routable base of the instance.
	PublicURL string

	// Phone is the number that incoming SMS messages from this instances will be from.
	// Must be unique between all instances.
	Phone string

	// ErrorsOnly, if set, will disable creating test alerts for the instance. Any error-alerts will
	// still be generated, however.
	ErrorsOnly bool
}

func (i Instance) Validate() error {
	if i.Location == "" {
		return errors.New("location is required")
	}
	if i.PublicURL == "" {
		return errors.New("public URL is required")
	}
	_, err := url.Parse(i.PublicURL)
	if err != nil {
		return fmt.Errorf("parse public URL: %v", err)
	}
	if i.Phone == "" {
		return errors.New("phone is required")
	}
	if i.ErrorAPIKey == "" {
		return errors.New("error API key is required")
	}

	if i.GenericAPIKey == "" && i.EmailAPIKey == "" {
		return errors.New("at least one of Email or Generic API key is required")
	}

	if i.EmailAPIKey != "" {
		if _, err := mail.ParseAddress(i.EmailAPIKey); err != nil {
			return fmt.Errorf("parse email API key: %v", err)
		}
		addr, domain, _ := strings.Cut(i.EmailAPIKey, "@")
		if len(domain) > 255 {
			return fmt.Errorf("email domain is too long")
		}

		if len(i.Location)+len(addr)+len("+rm-") > 64 {
			return fmt.Errorf("location + email address is too long")
		}
	}

	return nil
}

func (i *Instance) doReq(path string, v url.Values) error {
	u, err := util.JoinURL(i.PublicURL, path)
	if err != nil {
		return err
	}
	resp, err := http.PostForm(u, v)
	if err != nil {
		return err
	}
	defer logClose(resp.Body.Close)
	if resp.StatusCode/100 != 2 {
		return errors.Errorf("non-200 response: %s", resp.Status)
	}
	return nil
}

func (i *Instance) createGenericAlert(key, dedup, summary, details string) error {
	v := make(url.Values)
	v.Set("token", key)
	v.Set("summary", summary)
	v.Set("details", details)
	v.Set("dedup", dedup)
	return i.doReq("/api/v2/generic/incoming", v)
}

func (i *Instance) heartbeat() []error {
	errCh := make(chan error, len(i.HeartbeatURLs))
	var wg sync.WaitGroup
	for _, u := range i.HeartbeatURLs {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			resp, err := http.Post(u, "", nil)
			if err != nil {
				errCh <- err
				return
			}
			defer logClose(resp.Body.Close)
			if resp.StatusCode/100 != 2 {
				errCh <- errors.Errorf("non-200 response: %s", resp.Status)
			}
		}(u)
	}
	wg.Wait()
	close(errCh)
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	return errs
}
