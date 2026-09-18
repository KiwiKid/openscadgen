package server

import (
	"context"
	"fmt"
	"sync"
)

// processingQueue is a bounded, context-cancellable worker pool for render
// jobs. One worker is deliberate: the local OpenSCAD renderer is process-heavy
// and its include/cache behaviour is not safe to assume concurrent. A future
// RenderEngine with declared concurrency support can raise this without
// changing the HTTP/polling contract.
type processingQueue struct {
	jobs    chan *queuedProcessingJob
	mu      sync.Mutex
	running map[string]*queuedProcessingJob
}

type queuedProcessingJob struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	run    func(context.Context)
}

func newProcessingQueue(workerCount, capacity int) *processingQueue {
	if workerCount < 1 {
		workerCount = 1
	}
	if capacity < 1 {
		capacity = 1
	}
	queue := &processingQueue{
		jobs:    make(chan *queuedProcessingJob, capacity),
		running: make(map[string]*queuedProcessingJob),
	}
	for range workerCount {
		go queue.worker()
	}
	return queue
}

func (q *processingQueue) submit(id string, run func(context.Context)) error {
	ctx, cancel := context.WithCancel(context.Background())
	job := &queuedProcessingJob{id: id, ctx: ctx, cancel: cancel, run: run}
	q.mu.Lock()
	q.running[id] = job
	q.mu.Unlock()
	select {
	case q.jobs <- job:
		return nil
	default:
		q.mu.Lock()
		delete(q.running, id)
		q.mu.Unlock()
		cancel()
		return fmt.Errorf("render queue is full")
	}
}

func (q *processingQueue) cancel(id string) bool {
	q.mu.Lock()
	job, ok := q.running[id]
	q.mu.Unlock()
	if ok {
		job.cancel()
	}
	return ok
}

func (q *processingQueue) worker() {
	for job := range q.jobs {
		job.run(job.ctx)
		q.mu.Lock()
		delete(q.running, job.id)
		q.mu.Unlock()
	}
}

// Keep a finite queue so a burst of HTTP requests cannot spawn an unbounded
// number of OpenSCAD processes or retain arbitrary configs in memory.
var renderJobQueue = newProcessingQueue(1, 32)
