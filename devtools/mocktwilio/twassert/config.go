package twassert

import "time"

type Config struct {
	ServerAPI
	// Timeout is used to set the timeout for all operations, expected messages/calls as well as API calls for things like answering a call.
	Timeout time.Duration

	// AppPhoneNumber is the phone number that the application will use to make calls and send messages.
	AppPhoneNumber string

	// RefreshFunc will be called before waiting for new messages or calls to arrive.
	//
	// It is useful for testing purposes to ensure pending messages/calls are sent from the application.
	//
	// Implementations should not return until requests to mocktwilio are complete.
	RefreshFunc func()
}
