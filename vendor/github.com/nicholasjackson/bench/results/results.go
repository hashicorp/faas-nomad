package results

import (
	"sort"
	"time"
)

// Result is a structure that encapsulates processable result data
type Result struct {
	Timestamp   time.Time
	RequestTime time.Duration
	Error       error
	Threads     int
}

// ResultSet is a convenience mapping for type []Result
type ResultSet []Result

// Implement Sort interface
func (r ResultSet) Len() int {
	return len(r)
}

func (r ResultSet) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ResultSet) Less(i, j int) bool {
	return r[i].Timestamp.UnixNano() < r[j].Timestamp.UnixNano()
}

// Reduce reduces the ResultSet into buckets defined by the given interval
func (r ResultSet) Reduce(interval time.Duration) []ResultSet {
	sort.Sort(r)

	start := r[0].Timestamp
	end := r[len(r)-1].Timestamp

	// create the buckets
	bucketCount := getBucketCount(start, end, interval)
	buckets := make([]ResultSet, bucketCount)

	for _, result := range r {

		currentBucket := getBucketNumber(result.Timestamp, start, end, interval, bucketCount)
		buckets[currentBucket] = append(buckets[currentBucket], result)
	}

	return buckets
}

func getBucketCount(start time.Time, end time.Time, interval time.Duration) int {
	if interval == time.Duration(0) {
		return 1
	}

	totalDuration := end.UnixNano() - start.UnixNano()
	bucketCount := time.Duration(totalDuration) / interval // nanoseconds

	return int(bucketCount) + 1
}

func getBucketNumber(current time.Time, start time.Time, end time.Time, interval time.Duration, bucketCount int) int {

	if interval == time.Duration(0) {
		return 0
	}

	bucketSize := int64(interval)
	curr := end.UnixNano() - current.UnixNano()

	return bucketCount - int(curr/bucketSize) - 1
}
