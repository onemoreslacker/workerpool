package main

import (
	"fmt"
	"log"

	workerpool "github.com/onemoreslacker/workerpool/pkg"
)

func main() {
	wp := workerpool.New() // Начальное число воркеров = 0.
	defer wp.Close()

	const numWorkers = 10
	for range numWorkers {
		if err := wp.AddWorker(); err != nil {
			log.Fatalf("failed to add worker: %s", err.Error())
		}
	}

	const numTasks = 5
	for i := range numTasks {
		if err := wp.Submit(fmt.Sprintf("задача %d", i)); err != nil {
			log.Fatalf("failed to submit task: %s", err.Error())
		}
	}
}
