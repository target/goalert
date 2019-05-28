package auth

import (
	"context"
	"net/http"
)

// An IdentityProvider provides an option for a user to login (identify themselves).
//
// Examples include user/pass, OIDC, LDAP, etc..
type IdentityProvider interface {
	Info(context.Context) ProviderInfo

	ExtractIdentity(*RouteInfo, http.ResponseWriter, *http.Request) (*Identity, error)
}

// Identity represents a user's proven identity.
type Identity struct {
	// SubjectID should be a provider-specific identifier for an individual.
	SubjectID     string
	Email         string
	EmailVerified bool
	Name          string
}

// ProviderInfo holds the details for using a provider.
type ProviderInfo struct {
	// Title is a user-viewable string for identifying this provider.
	Title string

	// LogoURL is the optional URL of an icon to display with the provider.
	LogoURL string `json:",omitempty"`

	// Fields holds a list of fields to include with the request.
	// The order specified is the order displayed.
	Fields []Field `json:",omitempty"`

	// Hidden indicates that the provider is not intended for user visibility.
	Hidden bool `json:"-"`

	// Enabled indicates that the provider is currently turned on.
	Enabled bool `json:"-"`
}

// Field represents a single form field for authentication.
type Field struct {
	// ID is the unique name/identifier of the field.
	// It will be used as the key name in the POST request.
	ID string

	// Label is the text displayed to the user for the field.
	Label string

	// Required indicates a field that must not be empty.
	Required bool

	// Password indicates the field should be treated as a password (gererally masked).
	Password bool

	// Scannable indicates the field can be entered via QR-code scan.
	Scannable bool
}
