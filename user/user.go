package user

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

// A User is the base information of a user of the system. Authentication details are stored
// separately based on the auth provider.
//
type User struct {
	// ID is the unique identifier for the user
	ID string `json:"id"`

	// Name is the full name of the user
	Name string `json:"name"`

	// Email is the primary contact email for the user. It is used for account-related communications
	Email string `json:"email"`

	// AvatarURL is an absolute address for an image to be used as the avatar.
	AvatarURL string `json:"avatar_url"`

	// AlertStatusCMID defines a contact method ID for alert status updates.
	AlertStatusCMID string `json:"alert_status_log_contact_method_id"`

	// The Role of the user
	Role permission.Role `json:"role" store:"readonly"`
}

// ResolveAvatarURL will resolve the user avatar URL, using the email if none is set.
func (u User) ResolveAvatarURL(fullSize bool) string {
	if u.AvatarURL == "" {
		suffix := ""
		if fullSize {
			suffix = "&s=2048"
		}
		sum := md5.Sum([]byte(u.Email))
		u.AvatarURL = fmt.Sprintf("https://gravatar.com/avatar/%s?d=retro%s", hex.EncodeToString(sum[:]), suffix)
	}
	return u.AvatarURL
}

type scanFn func(...interface{}) error

func (u *User) scanFrom(fn scanFn) error {
	var statusCM sql.NullString
	err := fn(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.AvatarURL,
		&u.Role,
		&statusCM,
	)
	u.AlertStatusCMID = statusCM.String
	return err
}

func (u *User) userUpdateFields() []interface{} {
	var statusCM sql.NullString
	if u.AlertStatusCMID != "" {
		statusCM.Valid = true
		statusCM.String = u.AlertStatusCMID
	}
	return []interface{}{
		u.ID,
		u.Name,
		u.Email,
		statusCM,
	}
}
func (u *User) fields() []interface{} {
	var statusCM sql.NullString
	if u.AlertStatusCMID != "" {
		statusCM.Valid = true
		statusCM.String = u.AlertStatusCMID
	}
	return []interface{}{
		u.ID,
		u.Name,
		u.Email,
		u.AvatarURL,
		u.Role,
		statusCM,
	}
}

// Normalize will produce a normalized/validated User struct.
// Will only do the validate if email is not empty
func (u User) Normalize() (*User, error) {
	var err error
	if u.ID == "" {
		u.ID = uuid.NewV4().String()
	}
	if u.Email != "" {
		err = validate.Email("Email", u.Email)
		// Sanitize Email after it has been validated.
		u.Email = validate.SanitizeEmail(u.Email)
	}

	if u.AvatarURL != "" {
		err = validate.Many(
			err,
			validate.AbsoluteURL("AvatarURL", u.AvatarURL),
		)
	}

	if u.AlertStatusCMID != "" {
		err = validate.Many(
			err,
			validate.UUID("AlertStatusCMID", u.AlertStatusCMID),
		)
	}

	err = validate.Many(
		err,
		validate.Name("Name", u.Name),
		validate.OneOf("Role", u.Role, permission.RoleAdmin, permission.RoleUser),
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
