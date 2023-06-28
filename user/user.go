package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

// A User is the base information of a user of the system. Authentication details are stored
// separately based on the auth provider.
type User struct {
	// ID is the unique identifier for the user
	ID string

	// Name is the full name of the user
	Name string

	// Email is the primary contact email for the user. It is used for account-related communications
	Email string

	// AvatarURL is an absolute address for an image to be used as the avatar.
	AvatarURL string

	// AlertStatusCMID defines a contact method ID for alert status updates.
	//
	// Deprecated: No longer used.
	AlertStatusCMID string

	// The Role of the user
	Role permission.Role

	// isUserFavorite returns true if a user is favorited by the current user.
	isUserFavorite bool
}

func (User) TableName() string { return "users" }

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
	err := fn(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.AvatarURL,
		&u.Role,
		&u.isUserFavorite,
	)
	return err
}

func (u *User) userUpdateFields() []interface{} {
	return []interface{}{
		u.ID,
		u.Name,
		u.Email,
	}
}

func (u *User) fields() []interface{} {
	return []interface{}{
		u.ID,
		u.Name,
		u.Email,
		u.AvatarURL,
		u.Role,
	}
}

// Normalize will produce a normalized/validated User struct.
func (u User) Normalize() (*User, error) {
	var err error
	if u.ID == "" {
		u.ID = uuid.New().String()
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

// IsUserFavorite returns true if a user is a favorite of the current user.
func (u User) IsUserFavorite() bool {
	return u.isUserFavorite
}
