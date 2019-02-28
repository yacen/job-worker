package main

import (
	"fmt"
	"github.com/yacen/job-worker"
	"time"
)

var (
	MaxWorker = 100 //os.Getenv("MAX_WORKERS")
	MaxQueue  = 10  //os.Getenv("MAX_QUEUE")
)

type PayloadImpl struct {
	number int
}

func (p *PayloadImpl) Do() error {
	fmt.Println("do business ...", p.number)
	return nil
}

func main() {
	dispatcher := job_worker.NewDispatcher(MaxWorker, MaxQueue)
	dispatcher.Run()

	for i := 10000000000; i > 0; i-- {
		job := job_worker.Job{Payload: &PayloadImpl{number: i}}
		// Submit the job onto the queue.
		dispatcher.Submit(job)
	}

	time.Sleep(100 * time.Hour)
}
