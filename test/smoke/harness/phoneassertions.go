package harness

// PhoneAssertions is used to assert voice and SMS behavior.
type PhoneAssertions interface {

	// Device returns a TwilioDevice for the given number.
	//
	// It is safe to call multiple times for the same device.
	Device(number string) PhoneDevice

	// WaitAndAssert will fail the test if there are any unexpected messages received within the timeout interval.
	WaitAndAssert()
}

// A PhoneDevice immitates a device (i.e. a phone) for testing interactions.
type PhoneDevice interface {
	// SendSMS will send a message to GoAlert from the device.
	SendSMS(body string)

	// ExpectSMS will match against an SMS that matches ALL provided keywords (case-insensitive).
	// Each call to ExpectSMS results in the requirement that an additional SMS is received.
	ExpectSMS(keywords ...string) ExpectedSMS

	// RejectSMS will match against an SMS that matches ALL provided keywords (case-insensitive) and tell the server that delivery failed.
	RejectSMS(keywords ...string)

	// ExpectVoice will match against a voice call where the spoken text matches ALL provided keywords (case-insensitive).
	ExpectVoice(keywords ...string) ExpectedCall

	// RejectVoice will match against a voice call where the spoken text matches ALL provided keywords (case-insensitive) and tell the server that delivery failed.
	RejectVoice(keywords ...string)

	// IgnoreUnexpectedSMS will cause any extra SMS messages (after processing ExpectSMS calls) that match
	// ALL keywords (case-insensitive) to not fail the test.
	IgnoreUnexpectedSMS(keywords ...string)

	// IgnoreUnexpectedVoice will cause any extra voice calls (after processing ExpectVoice) that match
	// ALL keywords (case-insensitive) to not fail the test.
	IgnoreUnexpectedVoice(keywords ...string)
}

// ExpectedCall represents a phone call.
type ExpectedCall interface {
	// ThenPress imitates a user entering a key on the phone.
	ThenPress(digits string) ExpectedCall

	// ThenExpect asserts that the message matches ALL keywords (case-insensitive).
	//
	// Generally used as ThenPress().ThenExpect()
	ThenExpect(keywords ...string) ExpectedCall

	// Body will return the last full spoken message as text. Separate stanzas (e.g. multiple `<Say>`) are
	// separated by newline.
	Body() string

	// Hangup will hangup the active call.
	Hangup()
}

// ExpectedSMS represents an SMS message.
type ExpectedSMS interface {

	// ThenReply will respond with an SMS with the given body.
	ThenReply(body string) SMSReply

	// Body is the text of the SMS message.
	Body() string

	// From is the source number of the SMS message.
	From() string
}

// SMSReply represents a reply to a received SMS message.
type SMSReply interface {
	// ThenExpect will match against an SMS that matches ALL provided keywords (case-insensitive).
	// The message must be received AFTER the reply is sent or the assertion will fail.
	ThenExpect(keywords ...string) ExpectedSMS
}
