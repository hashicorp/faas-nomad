package results

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSortsResultSet(t *testing.T) {

	r1 := ResultSet{
		Result{Timestamp: time.Now().Add(1 * time.Second)},
		Result{Timestamp: time.Now()},
	}
	sort.Sort(r1)

	assert.True(t, r1[0].Timestamp.Unix() < r1[1].Timestamp.Unix())
}

func TestDurationOf0ShouldReturn1Bucket(t *testing.T) {

	r1 := ResultSet{
		Result{Timestamp: time.Now()},
		Result{Timestamp: time.Now().Add(20 * time.Millisecond)},
	}

	buckets := r1.Reduce(0 * time.Minute)

	assert.Equal(t, 1, len(buckets))
}

func TestDurationSmallerThanResultsSpanShouldReturn2Buckets(t *testing.T) {

	r1 := ResultSet{
		Result{Timestamp: time.Now().Add(1103 * time.Millisecond)},
		Result{Timestamp: time.Now()},
	}

	buckets := r1.Reduce(1000 * time.Millisecond)

	assert.Equal(t, 2, len(buckets))
}

func TestResultsWithGapShouldReturn3Buckets(t *testing.T) {

	r1 := ResultSet{
		Result{Timestamp: time.Now().Add(2001 * time.Millisecond)},
		Result{Timestamp: time.Now()},
	}

	buckets := r1.Reduce(1 * time.Second)

	assert.Equal(t, 3, len(buckets))
	assert.Equal(t, 1, len(buckets[0]))
	assert.Equal(t, 0, len(buckets[1])) //second bucket should be blank
}

func TestBucketNumberIsReturnedCorrectly(t *testing.T) {

	now := time.Now()

	bucket := getBucketNumber(
		now,
		now.Add(-3*time.Second),
		now.Add(2*time.Second),
		1*time.Second,
		getBucketCount(now.Add(-3*time.Second), now.Add(2*time.Second), 1*time.Second),
	)

	assert.Equal(t, 3, bucket)
}

func TestBucketNumberIsReturnedCorrectly2(t *testing.T) {

	now := time.Now()

	bucket := getBucketNumber(
		now.Add(-1*time.Second),
		now.Add(-3*time.Second),
		now.Add(2*time.Second),
		1*time.Second,
		getBucketCount(now.Add(-3*time.Second), now.Add(2*time.Second), 1*time.Second),
	)

	assert.Equal(t, 2, bucket)
}

func TestBucketNumberIsReturnedCorrectly3(t *testing.T) {

	now := time.Now()

	bucket := getBucketNumber(
		now.Add(-3*time.Second),
		now.Add(-3*time.Second),
		now.Add(2*time.Second),
		1*time.Second,
		getBucketCount(now.Add(-3*time.Second), now.Add(2*time.Second), 1*time.Second),
	)

	assert.Equal(t, 0, bucket)
}

func TestBucketNumberIsReturnedCorrectly4(t *testing.T) {

	now := time.Now()

	bucket := getBucketNumber(
		now.Add(2*time.Second),
		now.Add(-3*time.Second),
		now.Add(2*time.Second),
		1*time.Second,
		getBucketCount(now.Add(-3*time.Second), now.Add(2*time.Second), 1*time.Second),
	)

	assert.Equal(t, 5, bucket)
}
func TestBucketNumberIsReturnedCorrectly5(t *testing.T) {

	now := time.Now()

	bucket := getBucketNumber(
		now.Add(2*time.Minute),
		now.Add(-3*time.Minute),
		now.Add(2*time.Minute),
		1*time.Minute,
		getBucketCount(now.Add(-3*time.Second), now.Add(2*time.Second), 1*time.Second),
	)

	assert.Equal(t, 5, bucket)
}
