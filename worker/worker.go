package worker

import (
	"context"
	"sync"
	"time"
)

// Workers define a number of goroutines that execute a given task
type Workers struct {
	tasks     chan func()
	waitGroup sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// Create the workers with a limit to the number define by maxWorkers
func NewWorkers(maxWorkers int) *Workers {
	ctx, cancel := context.WithCancel(context.Background())
	workers := &Workers{
		tasks:  make(chan func()),
		ctx:    ctx,
		cancel: cancel,
	}
	for i := 0; i < maxWorkers; i++ {
		go workers.worker()
	}
	return workers
}

// worker is create the worker for execute a task pass throught the Task function
func (w *Workers) worker() {
	for task := range w.tasks {
		w.waitGroup.Add(1)
		task()
		w.waitGroup.Done()
	}
}

// Pass a task to be executed
// If a worker is available the task will be taken and return true
// If not it wait 100ms to have a available worker to take the task,
// At the end of the wait the task is rejected and return false
func (w *Workers) Task(task func()) bool {

	select {
	case w.tasks <- task:
		return true
	case <-time.After(100 * time.Millisecond):
		return false
	default:
		return false
	}
}

// Shutdown close properly the task and the workers
func (w *Workers) Shutdown() {
	close(w.tasks)
	w.cancel()
}
