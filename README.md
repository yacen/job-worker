# job-worker 示例

## Payload
```go
type Payload interface {
	Do() error
}
```
提供具体 业务代码 的接口

## Job
```go
type Job struct {
	Payload Payload
}
```
抽像任务，Payload代表需要实现的业务实体，给 worker 调用

## Worker
```go
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
```
需要执行 job 的worker。
* WorkerPool 空闲工人池
* JobChannel 工人要做的任务通道
* quit 退出时，发信号

### Worker.Start()
```go
func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w
			select {
			case job := <-w.JobChannel:
				if err := job.Payload.Do(); err != nil {
					log.Printf("Error uploading to S3: %s", err.Error())
				}
			case <-w.quit:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

```
Start() 方法开一个 goroutine ，死循环从 任务通道里取任务做， 任务完成后 放入 WorkerPool。

当调用 Stop() 方法时，退出死循环。

## Dispatcher
任务分发器
```go
type Dispatcher struct {
	WorkerPool chan chan Job
	maxWorkers int
	JobQueue chan Job
	workers []Worker
}

func NewDispatcher(maxWorkers, maxQueueSize int) *Dispatcher {
	pool := make(chan Worker, maxWorkers)
	jobQueue := make(chan Job, maxQueueSize)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers, JobQueue: jobQueue}
}
```
* WorkerPool 空闲工人池
* maxWorkers 工人总数
* JobQueue 总任务队列
* workers 工人

## Dispatcher.Run()
```go
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		d.workers = append(d.workers, worker)
		worker.Start()
	}

	go d.dispatch()
}

```
创建worker， 开始分发

## Dispatcher.dispatch()
```go
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
```
循环的从任务队列中获取任务，从空闲工人池中得到一个工人，把任务交给它做。工人会在 Worker.Start()方法里的 goroutine 里获取任务并执行。

## Dispatcher.Submit(job Job)
```go
func (d *Dispatcher) Submit(job Job) {
	d.JobQueue <- job
}
``` 
提交一个任务到任务队列中去。 dispatch()方法就会获取到它。