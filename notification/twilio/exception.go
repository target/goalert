package twilio

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/target/goalert/notification"
)

// Exception contains information on a Twilio error.
type Exception struct {
	Status   int
	Message  string
	Code     int
	MoreInfo string `json:"more_info"`
}

func (e Exception) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// wrapStatusErr will wrap a 404 Exception with notification.ErrStatusNotFound
// so callers can stop checking the status of resources that no longer exist
func wrapStatusErr(err error) error {
	var e *Exception
	if errors.As(err, &e) && e.Status == http.StatusNotFound {
		return fmt.Errorf("%w: %w", notification.ErrStatusNotFound, err)
	}

	return err
}
