package alert

import (
	"strings"
	"time"

	"github.com/target/goalert/validation/validate"
)

// MaxCommentLength is the maximum length of an alert comment body.
//
// Kept in sync with the alert_comments_body_max_length CHECK constraint.
const MaxCommentLength = 4096

// A Comment is a free-form note left on an alert by a user.
//
// Comments capture triage context that does not belong in the system-generated
// alert log: what someone tried, what they found, who they handed it to.
type Comment struct {
	ID      int    `json:"id"`
	AlertID int    `json:"alert_id"`
	Body    string `json:"body"`

	// UserID is empty when the author's account has since been deleted. The
	// comment itself is retained.
	UserID string `json:"user_id"`

	CreatedAt time.Time `json:"created_at"`
}

func (c Comment) Normalize() (*Comment, error) {
	c.Body = strings.TrimSpace(c.Body)

	// RequiredText, not Text: Text returns nil for an empty string, which would
	// let a blank comment through to the alert_comments_body_not_empty CHECK
	// and surface as a raw SQL error instead of a field error.
	err := validate.RequiredText("Body", c.Body, 1, MaxCommentLength)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
