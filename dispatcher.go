package job_worker

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	WorkerPool chan Worker

	maxWorkers int

	// A buffered channel that we can send work requests on.
	JobQueue chan Job

	workers []Worker
}

func NewDispatcher(maxWorkers, maxQueueSize int) *Dispatcher {
	pool := make(chan Worker, maxWorkers)
	jobQueue := make(chan Job, maxQueueSize)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers, JobQueue: jobQueue}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		d.workers = append(d.workers, worker)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) Submit(job Job) {
	d.JobQueue <- job
}

func (d *Dispatcher) Stop() {
	for _, worker := range d.workers {
		worker.Stop()
	}
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				worker := <-d.WorkerPool

				// dispatch the job to the worker job channel
				worker.JobChannel <- job
			}(job)
		}
	}
}
