// Package task provides a work pool, for doing many things
// concurrently.
package task

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrPoolBusy = errors.New("task: work not done, pool not free in time")

// Worker defines the single method necessary
// to do work in the pool.
type Worker interface {
	Work()
}

// WorkerFunc is an adapter to allow the use of
// ordinary functions as Workers. If f is a function
// with an appropriate signature, WorkerFunc is a
// Worker which calls f.
type WorkerFunc func()

// Work calls WorkerFunc w().
func (w WorkerFunc) Work() {
	w()
}

// Task provides a pool of goroutines for Worker
// tasks to be submitted.
type Task struct {
	// the Worker channel is used to submit work
	// to the pool
	work      chan Worker
	wg        sync.WaitGroup
	submitted uint64
	completed uint64
}

// New creates a new work pool.
func New(poolSize int) *Task {
	t := Task{
		// Use unbuffered channel to guarantee work
		// submitted is being worked on after a call
		// to t.Do
		work: make(chan Worker),
	}

	t.wg.Add(poolSize)

	for i := 0; i < poolSize; i++ {
		go func() {
			for w := range t.work {
				w.Work()
				t.incrementCompleted()
			}

			t.wg.Done()
		}()
	}

	return &t
}

// Do submits work to the pool.
// An error is returned in the case the pool is saturated
// with existing work and does not respond before the
// context is canceled.
func (t *Task) Do(ctx context.Context, w Worker) error {
	t.incrementSubmitted()

	select {
	case <-ctx.Done():
		return ErrPoolBusy
	case t.work <- w:
		return nil
	}
}

// Shutdown waits for all the goroutines to shutdown.
func (t *Task) Shutdown() {
	close(t.work)
	t.wg.Wait()
}

// Submitted returns the number of submitted work items.
func (t *Task) Submitted() uint64 {
	submitted := atomic.LoadUint64(&t.submitted)

	return submitted
}

// Completed returns the number of completed work items.
func (t *Task) Completed() uint64 {
	completed := atomic.LoadUint64(&t.completed)

	return completed
}

func (t *Task) incrementSubmitted() {
	atomic.AddUint64(&t.submitted, 1)
}

func (t *Task) incrementCompleted() {
	atomic.AddUint64(&t.completed, 1)
}
