package semaphore

import (
	"sync"
	"time"
)

// Semaphore is an object which can be used to limit concurrency.
// It is a wrapper arround a buffered channel with convenience methods
// to be able to increase or decrease the capacity of the channel.
type Semaphore struct {
	s           chan struct{}
	lockLock    chan struct{}
	lockDone    chan struct{}
	resizeMutex sync.Mutex
	readMutex   sync.RWMutex
	rampUp      time.Duration
}

// NewSemaphore is used to create a new semaphore, initalised with a capacity
// this controls the number of locks which can be active at any one time.
func NewSemaphore(capacity int, rampUp time.Duration) *Semaphore {
	s := Semaphore{
		lockDone: make(chan struct{}),
		lockLock: make(chan struct{}),
		rampUp:   rampUp,
	}

	// if the rampup time is less than 1 then return immediately
	if rampUp < 1 {
		s.s = make(chan struct{}, capacity)
	} else {
		s.s = make(chan struct{}, 1)
		go s.rampUpThreads(capacity, rampUp)
	}

	s.resizeUnlock()

	return &s
}

// Lock takes out a lock, once the number of locks has reached the capacity
// Then Lock will block untill release is called.
func (t *Semaphore) Lock() {

	t.waitIfResizing()
	t.s <- struct{}{} // if we are already blocking here then we can cause problems
}

// Release unlocks the semaphore and allows new lock instances to be called without
// blocking if the number of locks currently equal the capacity.
// It is important to call Release at the end of any operation which aquires a lock.
func (t *Semaphore) Release() {
	t.waitIfResizing()

	// we need a read lock to ensure we do not resize whilst resizing
	t.readMutex.RLock()
	defer t.readMutex.RUnlock()

	// make sure we have not called Release without Lock
	if len(t.s) == 0 {
		return
	}
	<-t.s
}

// Resize allows dynamic resizing of the semaphore, it can be used if it desired
// to increase the current number of allowable concurent processes.
func (t *Semaphore) Resize(capacity int) {

	// only allow one resize to be called from one thread
	t.resizeMutex.Lock()

	if capacity == cap(t.s) {
		t.resizeMutex.Unlock()
		return
	}

	// lock the locks
	t.resizeLock()
	t.readMutex.Lock()
	defer t.resizeUnlock()
	defer t.resizeMutex.Unlock()
	defer t.readMutex.Unlock()

	new := make(chan struct{}, capacity) // create the new semaphore with the new capcity

	// copy the old values
	for n := len(t.s); n != 0; n = len(t.s) {
		new <- <-t.s // copy elements to the new channel
	}

	t.s = new
}

// Capacity returns the current capacity of the semaphore, or the number
// of permissable locks before Lock will block.
func (t *Semaphore) Capacity() int {

	t.waitIfResizing()
	return cap(t.s)
}

// Length is the current size of the semaphore, a size greater than 0 indicates that locks are
// still active
func (t *Semaphore) Length() int {

	t.waitIfResizing()
	return len(t.s)
}

func (t *Semaphore) resizeUnlock() {

	go func() {
		for {
			select {
			case <-t.lockDone:
				return
			case <-t.lockLock:
			}
		}
	}()
}

func (t *Semaphore) resizeLock() {
	t.lockDone <- struct{}{}
}

func (t *Semaphore) waitIfResizing() {
	t.lockLock <- struct{}{}
}

func (t *Semaphore) rampUpThreads(threads int, rampUp time.Duration) {

	timerInterval := rampUp / time.Duration(threads-len(t.s))
	timer := time.NewTicker(timerInterval)

	size := cap(t.s)

	for range timer.C {
		size++
		t.Resize(size)

		if size >= threads {
			timer.Stop()
		}
	}
}
