package errors

// Timeout represents an error from a timed out request
type Timeout struct {
	Message string
}

func (t Timeout) Error() string {
	return t.Message
}
