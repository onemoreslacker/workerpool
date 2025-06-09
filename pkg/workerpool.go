package workerpool

import (
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	data     chan string
	commands chan command
	stop     chan struct{}
	workers  map[int]*worker

	wg sync.WaitGroup
	mu sync.RWMutex

	nextWorkerID int
	released     *atomic.Bool
}

func New() *WorkerPool {
	wp := &WorkerPool{
		data:     make(chan string),
		commands: make(chan command),
		stop:     make(chan struct{}),
		workers:  make(map[int]*worker),
		released: &atomic.Bool{},
	}

	go wp.processor()
	wp.wg.Add(1)

	return wp
}

func (wp *WorkerPool) AddWorker() error {
	if wp.released.Load() {
		return ErrPoolClosed
	}

	cmd := newCommand(addWorker)

	select {
	case <-wp.stop:
		return ErrPoolClosed
	case wp.commands <- cmd:
		return <-cmd.done
	}
}

func (wp *WorkerPool) RemoveWorker() error {
	if wp.released.Load() {
		return ErrPoolClosed
	}

	cmd := newCommand(removeWorker)

	select {
	case <-wp.stop:
		return ErrPoolClosed
	case wp.commands <- cmd:
		return <-cmd.done
	}
}

func (wp *WorkerPool) Submit(job string) error {
	if wp.released.Load() {
		return ErrPoolClosed
	}

	select {
	case <-wp.stop:
		return ErrPoolClosed
	case wp.data <- job:
		return nil
	}
}

func (wp *WorkerPool) Running() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return len(wp.workers)
}

func (wp *WorkerPool) Close() {
	if !wp.released.CompareAndSwap(false, true) {
		return
	}

	close(wp.stop)

	wp.wg.Wait()

	close(wp.data)
	close(wp.commands)
}

func (wp *WorkerPool) IsClosed() bool {
	return wp.released.Load()
}

func (wp *WorkerPool) processor() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.stop:
			wp.releaseAllWorkers()
			return
		case cmd := <-wp.commands:
			var err error
			switch cmd.action {
			case addWorker:
				wp.addWorker()
			case removeWorker:
				err = wp.removeWorker()
			}
			cmd.done <- err
		}
	}
}

func (wp *WorkerPool) addWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.nextWorkerID++
	w := newWorker(wp.nextWorkerID, wp.data)
	wp.workers[wp.nextWorkerID] = w

	wp.wg.Add(1)
	go w.start(&wp.wg)
}

func (wp *WorkerPool) removeWorker() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if len(wp.workers) == 0 {
		return ErrNoWorkers
	}

	var (
		workerID    int
		eraseWorker *worker
	)

	for id, w := range wp.workers {
		workerID = id
		eraseWorker = w
		break
	}

	eraseWorker.stop()
	delete(wp.workers, workerID)

	return nil
}

func (wp *WorkerPool) releaseAllWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for id, w := range wp.workers {
		w.stop()
		delete(wp.workers, id)
	}
}
