package output

import (
	"fmt"
	"io"
	"time"

	"github.com/nicholasjackson/bench/results"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// PlotData plots the results onto a graph and saves the output to the given writer
func PlotData(interval time.Duration, r results.ResultSet, w io.Writer) {
	set := r.Reduce(interval)
	t := results.TabularResults{}
	rows := t.Tabulate(set)

	seriesY := createYSeries(rows)
	seriesX := createXSeries(rows)
	ticks := createXTicks(rows)

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   25,
				Right:  25,
				Bottom: 25,
			},
			FillColor: drawing.ColorFromHex("efefef"),
		},
		XAxis: chart.XAxis{
			Name:      "Elapsed Time (s)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f", v)
			},
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			Name:      "Count",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f", v)
			},
		},
		YAxisSecondary: chart.YAxis{
			Name:      "Time (ms)",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.2f", v)
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "Success",
				XValues: seriesX["x"],
				YValues: seriesY["y.success"],
				Style: chart.Style{
					Show:        true,                             //note; if we set ANY other properties, we must set this to true.
					StrokeColor: drawing.ColorGreen,               // will supercede defaults
					FillColor:   drawing.ColorGreen.WithAlpha(64), // will supercede defaults
				},
			},
			chart.ContinuousSeries{
				Name:    "Failure",
				XValues: seriesX["x"],
				YValues: seriesY["y.failure"],
				Style: chart.Style{
					Show:        true,                           //note; if we set ANY other properties, we must set this to true.
					StrokeColor: drawing.ColorRed,               // will supercede defaults
					FillColor:   drawing.ColorRed.WithAlpha(64), // will supercede defaults
				},
			},
			chart.ContinuousSeries{
				Name:    "Timeout",
				XValues: seriesX["x"],
				YValues: seriesY["y.timeout"],
				Style: chart.Style{
					Show:        true,                                         //note; if we set ANY other properties, we must set this to true.
					StrokeColor: drawing.ColorFromHex("FFD133"),               // will supercede defaults
					FillColor:   drawing.ColorFromHex("FFD133").WithAlpha(64), // will supercede defaults
				},
			},
			chart.ContinuousSeries{
				Name:    "Threads",
				XValues: seriesX["x"],
				YValues: seriesY["y.threads"],
				Style: chart.Style{
					Show:        true,                           //note; if we set ANY other properties, we must set this to true.
					StrokeColor: drawing.ColorFromHex("FF338D"), // will supercede defaults
				},
			},
			chart.ContinuousSeries{
				YAxis:   chart.YAxisSecondary,
				Name:    "Request time (ms)",
				XValues: seriesX["x"],
				YValues: seriesY["y.request"],
				Style: chart.Style{
					Show:        true,              //note; if we set ANY other properties, we must set this to true.
					StrokeColor: drawing.ColorBlue, // will supercede defaults
				},
			},
		},
	}

	//note we have to do this as a separate step because we need a reference to graph
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	graph.Render(chart.PNG, w)
}

func createXTicks(results []results.Row) []chart.Tick {
	var ticks []chart.Tick

	maxTicks := 20
	totalTime := results[len(results)-1].ElapsedTime.Seconds()
	tickInterval := float64(totalTime) / float64(maxTicks)

	fmt.Println(tickInterval)
	nextTick := 0.0

	for i := range results {
		if tickInterval < 1 || float64(i) >= nextTick {

			tick := chart.Tick{
				Value: nextTick,
				Label: fmt.Sprintf("%.0f", nextTick),
			}
			ticks = append(ticks, tick)
			nextTick += tickInterval
		}
	}

	// add one last tick for the end
	lastTick := results[len(results)-1].ElapsedTime
	tick := chart.Tick{
		Value: lastTick.Seconds(),
		Label: fmt.Sprintf("%.0f", lastTick.Seconds()),
	}
	ticks = append(ticks, tick)

	return ticks
}

func createXSeries(results []results.Row) map[string][]float64 {
	series := map[string][]float64{
		"x": make([]float64, 0),
	}

	for _, row := range results {
		series["x"] = append(series["x"], row.ElapsedTime.Seconds())
	}

	return series
}

func createYSeries(results []results.Row) map[string][]float64 {
	series := map[string][]float64{
		"y.success": make([]float64, 0),
		"y.failure": make([]float64, 0),
		"y.timeout": make([]float64, 0),
		"y.request": make([]float64, 0),
		"y.threads": make([]float64, 0),
	}

	for _, row := range results {
		series["y.success"] = append(series["y.success"], float64(row.TotalSuccess))
		series["y.failure"] = append(series["y.failure"], float64(row.TotalFailures))
		series["y.timeout"] = append(series["y.timeout"], float64(row.TotalTimeouts))
		series["y.threads"] = append(series["y.threads"], float64(row.Threads))
		series["y.request"] = append(series["y.request"], float64(row.AvgRequestTime.Nanoseconds()/1e6))
	}

	return series
}
