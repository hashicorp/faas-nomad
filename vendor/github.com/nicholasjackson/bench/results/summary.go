package results

import (
	"fmt"
	"time"

	"github.com/nicholasjackson/bench/errors"
)

// Row represents a computed set of benchmark results
type Row struct {
	StartTime      time.Time
	ElapsedTime    time.Duration
	TotalRequests  int
	Threads        int
	AvgRequestTime time.Duration
	TotalSuccess   int
	TotalFailures  int
	TotalTimeouts  int
}

// String implements String from the Stringer interface and
// allows results to be serialized to a sting
func (r Row) String() string {

	rStr := fmt.Sprintf("Start Time: %v\n", r.StartTime.UTC())
	rStr = fmt.Sprintf("%vElapsed Time: %v\n", rStr, r.ElapsedTime)
	rStr = fmt.Sprintf("%vThreads: %v\n", rStr, r.Threads)
	rStr = fmt.Sprintf("%vTotal Requests: %v\n", rStr, r.TotalRequests)
	rStr = fmt.Sprintf("%vAvg Request Time: %v\n", rStr, r.AvgRequestTime)
	rStr = fmt.Sprintf("%vTotal Success: %v\n", rStr, r.TotalSuccess)
	rStr = fmt.Sprintf("%vTotal Timeouts: %v\n", rStr, r.TotalTimeouts)
	return fmt.Sprintf("%vTotal Failures: %v\n", rStr, r.TotalFailures)
}

// TabularResults processes a set of results and writes a tabular
// summary to the standard output
type TabularResults struct{}

// Tabulate transforms the ResultsSets and returns a slice of Row
func (t *TabularResults) Tabulate(results []ResultSet) []Row {

	var rows []Row
	startTime := time.Unix(0, 0)

	for _, bucket := range results {

		if len(bucket) > 0 {

			var elapsedTime time.Duration

			if startTime == time.Unix(0, 0) {
				startTime = bucket[0].Timestamp
			}
			elapsedTime = bucket[0].Timestamp.Sub(startTime)

			row := Row{
				StartTime:      bucket[0].Timestamp,
				ElapsedTime:    elapsedTime,
				Threads:        0,
				TotalRequests:  0,
				TotalFailures:  0,
				TotalSuccess:   0,
				TotalTimeouts:  0,
				AvgRequestTime: 0,
			}

			totalRequestTime := 0 * time.Second
			maxThreads := 0

			for _, r := range bucket {

				row.TotalRequests++

				if r.Error != nil {
					if _, ok := r.Error.(errors.Timeout); ok {
						row.TotalTimeouts++
					}

					row.TotalFailures++
				} else {
					row.TotalSuccess++
					totalRequestTime += r.RequestTime
				}

				if r.Threads > maxThreads {
					maxThreads = r.Threads
					row.Threads = maxThreads
				}
			}

			if totalRequestTime != 0 && row.TotalSuccess != 0 {
				avgTime := int64(totalRequestTime) / int64(row.TotalSuccess)
				row.AvgRequestTime = time.Duration(avgTime)
			}

			rows = append(rows, row)
		}
	}

	return rows
}
