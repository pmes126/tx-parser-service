package conc

import (
	"context"
	"sync"
)

// In is used from the WorkerPool for adding tasks to the worker pool.
type In[I any] chan I

// Out is used from the WorkerPool for returning Result.
type Out chan error

// Job is the job run by the workers in the pool. I is the input of the job and T is the output type.
type Job[I any] func(ctx context.Context, task I) error

// WorkerPool creates maxWorkers goroutines to handle incoming tasks.
type WorkerPool[I any] struct {
	maxWorkers int
	job        Job[I]
	wg         *sync.WaitGroup
	in         In[I]
	out        Out
}

// NewWorkerPool creates new WorkerPool with max workers and buffer size for input/output channels.
func NewWorkerPool[I any](
	maxWorkers int,
	job Job[I],
	bufferSize int,
) *WorkerPool[I] {
	return &WorkerPool[I]{
		maxWorkers: maxWorkers,
		job:        job,
		wg:         &sync.WaitGroup{},
		in:         make(In[I], bufferSize),
		out:        make(Out, bufferSize),
	}
}

// PushTask used to push tasks into the worker pool.
func (wp *WorkerPool[I]) PushTask(task I) {
	wp.in <- task
}

// CloseInputChannel is used to indicate no more incoming tasks.
func (wp *WorkerPool[I]) CloseInputChannel() {
	close(wp.in)
}

// Start starts the workerPool by creating maxWorkers goroutines and a goroutine to watch the wait group
// it accepts a context to receive termination signals an returns the out channel that can be used to return
// results to the callers
func (wp *WorkerPool[I]) Start(ctx context.Context) Out {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for {
				select {
				case task, ok := <-wp.in:
					if !ok {
						return
					}
					err := wp.job(ctx, task)
					wp.out <- err
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wp.wg.Wait()
		close(wp.out)
	}()

	return wp.out
}
