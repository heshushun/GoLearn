package calc

type Dispatcher struct {
	WorkerPool chan chan Job //worker的池子，控制worker的数量
	WorkerList []Worker      //worker的切片
}

//根据传入的值，创建对应数量的channel
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{
		WorkerPool: pool,
	}
}

//根据最大值，创建对应数量的worker
func (d *Dispatcher) Run() {
	for i := 0; i < MaxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
		d.WorkerList = append(d.WorkerList, worker)
	}
	//监听工作队列
	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				jobChannel := <-d.WorkerPool
				jobChannel <- job
			}(job)
		}
	}
}

//停止所有的worker
func (d *Dispatcher) Stop() {
	for _, worker := range d.WorkerList {
		worker.Stop()
	}
}
