package bench

import "github.com/nicholasjackson/bench/results"

func (b *Bench) processResults(r results.ResultSet) {

	for _, out := range b.outputs {
		out.function(out.interval, r, out.writer)
	}
}
