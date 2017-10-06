package semaphore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSemaphoreHasCapacity1(t *testing.T) {
	th := NewSemaphore(1, 0)

	assert.Equal(t, 1, th.Capacity())
}

func TestResizeIncreasesCapacityTo2(t *testing.T) {
	th := NewSemaphore(1, 0)
	th.Resize(2)
	assert.Equal(t, 2, th.Capacity())
}

func TestResizeWhenBlockingCreatesNewThread(t *testing.T) {

	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	th := NewSemaphore(1, 0)
	go func() {
		th.Lock()
		th.Resize(2)
		th.Lock()
		done <- true
	}()

	select {
	case <-timeout:
		assert.Fail(t, "Timed out waiting for lock")
	case <-done:
		assert.Equal(t, 2, len(th.s), "We should have two elements in the semaphore as we called lock twice")
	}
}

func TestTimesoutWhenBlocking(t *testing.T) {

	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	th := NewSemaphore(1, 0)

	go func() {
		th.Lock()
		th.Resize(1) // ensure resizing to the same size does not kill threads
		th.Lock()    //should block
		done <- true
	}()

	select {
	case <-timeout:
		assert.Equal(t, 1, th.Length(), "We should have only one element in the semaphore as the second call to lock should block")
	case <-done:
		assert.Fail(t, "Test should not have completed")
	}
}

func TestReleaseReducesLength(t *testing.T) {

	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	th := NewSemaphore(1, 0)

	go func() {
		th.Lock()
		th.Release()
		done <- true
	}()

	select {
	case <-timeout:
		assert.Fail(t, "Test should have completed")
	case <-done:
		assert.Equal(t, 0, th.Length(), "We should have 0 elements in the semaphore")
	}
}

func Test2ReleasesWithOneLockDoesNotBlock(t *testing.T) {

	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	th := NewSemaphore(1, 0)

	go func() {
		th.Lock()
		th.Resize(2)
		th.Release()
		th.Release()
		done <- true
	}()

	select {
	case <-timeout:
		assert.Fail(t, "Timed out waiting for test to complete")
	case <-done:
		assert.Equal(t, 0, th.Length(), "We should have 0 elements in the semaphore")
	}
}

func Test2ReleasesAfterResizeBehavesCorrectly(t *testing.T) {

	timeout := time.After(2 * time.Second)
	done := make(chan bool)

	th := NewSemaphore(1, 0)

	go func() {
		th.Lock()
		th.Resize(2)
		th.Lock()
		th.Release()
		th.Release()
		done <- true
	}()

	select {
	case <-timeout:
		assert.Fail(t, "Timed out waiting for test to complete")
	case <-done:
		assert.Equal(t, 0, th.Length(), "We should have 0 elements in the semaphore")
	}
}

func TestResizeAndLockDoesNotCauseDeadlock(t *testing.T) {

	th := NewSemaphore(5, 0)

	go func() {
		for {
			th.Lock()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			th.Release()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	th.Resize(10)

	assert.Equal(t, 10, th.Capacity())
}

func TestSemaphoreIncreasesOverTime(t *testing.T) {

	th := NewSemaphore(5, 500*time.Millisecond)

	go func() {
		for {
			th.Lock()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			th.Release()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	timer := time.NewTimer(600 * time.Millisecond)
	<-timer.C
	assert.Equal(t, 5, th.Capacity())
}
func TestResizeToSameSizeDoesNothing(t *testing.T) {

	th := NewSemaphore(5, 0)

	go func() {
		for {
			th.Lock()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			th.Release()
			time.Sleep(100 * time.Millisecond)
		}
	}()

	th.Resize(5)
	assert.Equal(t, 5, th.Capacity())
}
