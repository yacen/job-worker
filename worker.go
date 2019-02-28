package job_worker

import "log"

// Worker represents the worker that executes the job
type Worker struct {
	WorkerPool chan Worker
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan Worker) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.WorkerPool <- w

			select {
			case job := <-w.JobChannel:
				// we have received a work request.
				if err := job.Payload.Do(); err != nil {
					log.Printf("Error uploading to S3: %s", err.Error())
				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
