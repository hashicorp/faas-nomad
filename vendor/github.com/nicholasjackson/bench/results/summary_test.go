package results

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
)

func createResultSets() []ResultSet {
	startTime := time.Now()

	return []ResultSet{
		ResultSet{
			Result{
				Timestamp: startTime,
			},
			Result{
				Timestamp: startTime.Add(1 * time.Second),
			},
		},
		ResultSet{
			Result{
				Timestamp: startTime.Add(2 * time.Second),
			},
			Result{
				Timestamp: startTime.Add(3 * time.Second),
			},
		},
	}
}

func TestReturns2RowsWhenIHave2ResultSet(t *testing.T) {
	sets := createResultSets()
	tabs := TabularResults{}

	rows := tabs.Tabulate(sets)

	assert.Equal(t, len(sets), len(rows))
}

func TestSetsTheStartTimeOfTheFirstRowToBeEqualToTheFirstResult(t *testing.T) {
	sets := createResultSets()
	tabs := TabularResults{}

	rows := tabs.Tabulate(sets)

	assert.Equal(t, sets[0][0].Timestamp, rows[0].StartTime)
}

func TestSetsTheStartTimeOfTheSecondRowToBeEqualToTheSecondResult(t *testing.T) {
	sets := createResultSets()
	tabs := TabularResults{}

	rows := tabs.Tabulate(sets)

	assert.Equal(t, sets[1][0].Timestamp, rows[1].StartTime)
}

func TestSetsTheElapsedTimeOfTheFirstRowTo0(t *testing.T) {
	sets := createResultSets()
	tabs := TabularResults{}

	rows := tabs.Tabulate(sets)

	assert.Equal(t, time.Duration(0), rows[0].ElapsedTime)
}

func TestSetsTheElapsedTimeOfTheSecondRowTo2(t *testing.T) {
	sets := createResultSets()
	tabs := TabularResults{}

	rows := tabs.Tabulate(sets)

	assert.Equal(t, time.Duration(2*time.Second), rows[1].ElapsedTime)
}
