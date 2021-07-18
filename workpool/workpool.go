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

func New(workerCount, poolLen int) *WorkPool {
	return &WorkPool{
		workerCount: workerCount,
		pool:        make(chan *Task, poolLen),
	}
}
