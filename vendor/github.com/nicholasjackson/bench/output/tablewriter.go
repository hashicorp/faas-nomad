package output

import (
	"io"
	"time"

	"github.com/nicholasjackson/bench/results"
)

// WriteTabularData writes the given results to the given output stream
func WriteTabularData(interval time.Duration, r results.ResultSet, w io.Writer) {

	set := r.Reduce(interval)
	t := results.TabularResults{}
	rows := t.Tabulate(set)

	for _, row := range rows {
		w.Write([]byte(row.String()))
		w.Write([]byte("\n"))
	}
}
