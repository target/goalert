package permission

//go:generate go run golang.org/x/tools/cmd/stringer -type SourceType

// SourceType describes a type of authentication used to authorize a context.
type SourceType int

const (
	// SourceTypeNotificationCallback is set when a context is authenticated via the response to an outgoing notification.
	SourceTypeNotificationCallback SourceType = iota

	// SourceTypeIntegrationKey is set when an integration key is used to provide permission on a context.
	SourceTypeIntegrationKey

	// SourceTypeAuthProvider is set when a provider from the auth package is used (e.g. the web UI).
	SourceTypeAuthProvider

	// SourceTypeContactMethod is set when a context is authorized for use of a user's contact method.
	SourceTypeContactMethod

	// SourceTypeHeartbeat is set when a context is authorized for use of a service's heartbeat.
	SourceTypeHeartbeat

	// SourceTypeNotificationChannel is set when a context is authorized for use of a notification channel.
	SourceTypeNotificationChannel

	// SourceTypeCalendarSubscription is set when a context is authorized for use of a calendar subscription.
	SourceTypeCalendarSubscription

	// SourceTypeGQLAPIKey is set when a context is authorized for use of the GraphQL API.
	SourceTypeGQLAPIKey
)

// SourceInfo provides information about the source of a context's authorization.
type SourceInfo struct {
	Type SourceType
	ID   string
}

func (s SourceInfo) String() string {
	str := s.Type.String()
	if s.ID != "" {
		// using curly-braces so that it doesn't look too confusing
		// if we ever run into an unknown source type
		//
		// unknown will show up as SourceType(n) where n is the int value.
		str += "{" + s.ID + "}"
	}
	return str
}
