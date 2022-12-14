package twassert

import (
	"testing"

	"github.com/target/goalert/devtools/mocktwilio"
)

// Assertions is used to assert voice and SMS behavior.
type Assertions interface {
	// Device returns a TwilioDevice for the given number.
	//
	// It is safe to call multiple times for the same device.
	Device(number string) Device

	// WaitAndAssert will fail the test if there are any unexpected messages received.
	WaitAndAssert()

	// WithT will return a new PhoneAssertions with a separate text context.
	WithT(*testing.T) Assertions
}

// A Device immitates a device (i.e. a phone) for testing interactions.
type Device interface {
	// SendSMS will send a message to GoAlert from the device.
	SendSMS(text string)

	// ExpectSMS will match against an SMS that matches ALL provided keywords (case-insensitive).
	// Each call to ExpectSMS results in the requirement that an additional SMS is received.
	ExpectSMS(keywords ...string) ExpectedSMS

	// RejectSMS will match against an SMS that matches ALL provided keywords (case-insensitive) and tell the server that delivery failed.
	RejectSMS(keywords ...string)

	// ExpectCall asserts for and returns the next phone call to the device.
	ExpectCall() RingingCall

	// ExpectVoice is a convenience method for PhoneDevice.ExpectCall().Answer().ExpectSay(keywords...).Hangup()
	ExpectVoice(keywords ...string)

	// IgnoreUnexpectedSMS will cause any extra SMS messages (after processing ExpectSMS calls) that match
	// ALL keywords (case-insensitive) to not fail the test.
	IgnoreUnexpectedSMS(keywords ...string)
}

// RingingCall is a call that is ringing and not yet answered or rejected.
type RingingCall interface {
	Answer() ExpectedCall
	Reject()
	RejectWith(mocktwilio.FinalCallStatus)
}

// ExpectedCall represents a phone call.
type ExpectedCall interface {
	// Press imitates a user entering a key on the phone.
	Press(digits string) ExpectedCall

	// IdleForever imitates a user waiting for a timeout (without pressing anything) on the phone.
	IdleForever() ExpectedCall

	// ExpectSay asserts that the spoken message matches ALL keywords (case-insensitive).
	ExpectSay(keywords ...string) ExpectedCall

	// Text will return the last full spoken message as text. Separate stanzas (e.g. multiple `<Say>`) are
	// separated by newline.
	Text() string

	// Hangup will hangup the active call.
	Hangup()
}

// ExpectedSMS represents an SMS message.
type ExpectedSMS interface {
	// ThenReply will respond with an SMS with the given text.
	ThenReply(text string) SMSReply

	// Text is the text of the SMS message.
	Text() string

	// From is the source number of the SMS message.
	From() string
}

// SMSReply represents a reply to a received SMS message.
type SMSReply interface {
	// ThenExpect will match against an SMS that matches ALL provided keywords (case-insensitive).
	// The message must be received AFTER the reply is sent or the assertion will fail.
	ThenExpect(keywords ...string) ExpectedSMS
}
