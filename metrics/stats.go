package metrics

// StatsD defines an interface for logging data with StatsD
type StatsD interface {
	Incr(name string, tags []string, rate float64) error
}
