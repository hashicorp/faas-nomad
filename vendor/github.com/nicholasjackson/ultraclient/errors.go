package ultraclient

import (
	"fmt"
	"net/url"
)

const (
	// ErrorTimeout is a constant to be used for an error message when the client
	// experiences a timeout.
	ErrorTimeout = "timeout"

	// ErrorCircuitOpen is a constant to be used for an error message when the
	// client opens a circuit
	ErrorCircuitOpen = "circuit open"

	// ErrorGeneral is a constant to be used for an error message when the
	// client returns a general unhandled error.
	ErrorGeneral = "general error"

	// ErrorUnableToCompleteRequest is a constant to be used for an error message
	// when the client is unable to complete the request.
	ErrorUnableToCompleteRequest = "unable to complete request"
)

// ClientError implements the Error interface and is a generic client error
type ClientError struct {
	// Message is the error message
	Message string

	// URL is the endpoint from which the message orginated
	URL url.URL
}

// Error implements the error interface
func (s ClientError) Error() string {
	return fmt.Sprintf("%v for url: %v", s.Message, s.URL.String())
}
