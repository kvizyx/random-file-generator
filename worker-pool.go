package main

import (
	"sync"
)

type worker struct {
	jobs chan func()
	stop chan bool
}

func (w *worker) run(readyWg *sync.WaitGroup) {
	readyWg.Done()

	for {
		select {
		case job := <-w.jobs:
			job()
		case <-w.stop:
			return
		}
	}
}

type workerPool struct {
	size    int
	workers map[int]worker
	jobs    chan func()
}

func newWorkerPool(size int) *workerPool {
	readyWg := &sync.WaitGroup{}

	pool := &workerPool{
		size:    size,
		workers: make(map[int]worker),
		jobs:    make(chan func()),
	}

	readyWg.Add(size)

	for i := range size {
		w := worker{
			jobs: pool.jobs,
			stop: make(chan bool),
		}

		pool.workers[i] = w

		go w.run(readyWg)
	}

	readyWg.Wait()

	return pool
}

func (wp *workerPool) submit(job func()) {
	wp.jobs <- job
}

func (wp *workerPool) stop() {
	for _, wk := range wp.workers {
		wk.stop <- true
	}

	close(wp.jobs)
}
