package workerpool

import (
	"fmt"
	"sync"
)

type worker struct {
	id    int
	input chan string
	quit  chan struct{}
}

func newWorker(id int, input chan string) *worker {
	return &worker{
		id:    id,
		input: input,
		quit:  make(chan struct{}),
	}
}

func (w *worker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.quit:
			return
		case data, ok := <-w.input:
			if !ok {
				return
			}
			fmt.Printf("worker %d is processing %s\n", w.id, data)
		}
	}
}

func (w *worker) stop() {
	close(w.quit)
}
