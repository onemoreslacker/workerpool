package workerpool

import "fmt"

type workerPoolError struct{ msg string }

func (e workerPoolError) Error() string {
	return fmt.Sprintf("error: %s", e.msg)
}

var (
	ErrPoolClosed = workerPoolError{msg: "worker pool closed"}
	ErrNoWorkers  = workerPoolError{msg: "no workers to remove"}
)
