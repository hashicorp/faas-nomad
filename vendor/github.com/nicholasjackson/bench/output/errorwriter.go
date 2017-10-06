package output

import (
	"fmt"
	"io"
	"time"

	"github.com/nicholasjackson/bench/results"
)

func WriteErrorLogs(internal time.Duration, r results.ResultSet, w io.Writer) {

	for _, row := range r {
		if row.Error != nil {
			w.Write([]byte(fmt.Sprintf("%v: %v\n", row.Timestamp, row.Error)))
		}
	}
}
