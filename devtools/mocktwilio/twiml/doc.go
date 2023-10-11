// Package twiml provides a type-safe way of building and parsing the TwiML markup language in Go.
//
// TwiML is an XML-based language used by Twilio to control phone calls and SMS messages. The package provides
// Go structs that can be used to generate TwiML documents or parse incoming TwiML messages.
//
// The `Response` struct is the root element of a TwiML document. It can contain a sequence of `Verb`s, such as
// `Say`, `Gather`, `Pause`, `Redirect`, or `Hangup`. Each `Verb` is represented by a Go struct that contains
// the corresponding TwiML attributes and child elements.
//
// Example usage:
//
//	r := &twiml.Response{
//	    Verbs: []twiml.Verb{
//	        &twiml.Say{
//	            Content: "Hello, world!",
//	        },
//	        &twiml.Gather{
//	            Action: "/process_gather",
//	            NumDigitsCount: 1,
//	            Verbs: []twiml.GatherVerb{
//	                &twiml.Say{
//	                    Content: "Press 1 to confirm, or any other key to try again.",
//	                },
//	            },
//	        },
//	    },
//	}
//	b, _ := xml.MarshalIndent(r, "", "  ")
//	fmt.Println(string(b))
//
// Output:
//
// <Response>
//
//	<Say>Hello, world!</Say>
//	<Gather action="/process_gather">
//	  <Say>Press 1 to confirm, or any other key to try again.</Say>
//	</Gather>
//
// </Response>
//
// Incoming TwiML messages can be parsed using the `Interpreter` struct, which reads a sequence of bytes and
// returns the next `Verb` in the message until the end is reached. The `Interpreter` can be used to implement
// server-side TwiML applications that respond to user input.
package twiml
