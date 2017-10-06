package bench

// Request interface represents a load test which would be executed by Bench
type Request interface {
	// Do executes the request and returns an error on failure or nil on success
	Do() error
}
