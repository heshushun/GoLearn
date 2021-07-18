package workpool

import (
	"context"
	"sync"
)

type TaskFunc func()

type Task struct {
	f TaskFunc
}

func (t *Task) Execute() {
	t.f()
}

type WorkPool struct {
	pool        chan *Task
	workerCount int

	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
	wg             sync.WaitGroup
}

func NewWorkPool(workerCount, poolLen int) *WorkPool {
	return &WorkPool{
		workerCount: workerCount,
		pool:        make(chan *Task, poolLen),
	}
}

func (w *WorkPool) PushTask(t *Task) {
	w.pool <- t
}

func (w *WorkPool) PushTaskFunc(f TaskFunc) {
	w.pool <- &Task{
		f: f,
	}
}

func (w *WorkPool) work() {
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case t := <-w.pool:
			t.Execute()
		}
	}
}

func (w *WorkPool) Start() *WorkPool {
	w.wg.Add(w.workerCount)
	w.stopCtx, w.stopCancelFunc = context.WithCancel(context.Background())
	for i := 0; i < w.workerCount; i++ {
		go w.work()
	}
	return w
}

func (w *WorkPool) Stop() {
	w.stopCancelFunc()
	w.wg.Wait()
}