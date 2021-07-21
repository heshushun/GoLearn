package workpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Flag int64

type TaskFunc func(w *WorkPool, args ...interface{}) bool

type Task struct {
	f    TaskFunc
	args []interface{}
}

func (t *Task) Execute(w *WorkPool) bool {
	return t.f(w, t.args...)
}

type WorkPool struct {
	pool        chan *Task
	workerCount int

	// stop 相关
	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
	wg             sync.WaitGroup

	// sleep 相关
	sleepCtx        context.Context
	sleepCancelFunc context.CancelFunc
	sleepSeconds    int64
	sleepNotify     chan bool
}

func NewWorkPool(workerCount, poolLen int) *WorkPool {
	return &WorkPool{
		workerCount: workerCount,
		pool:        make(chan *Task, poolLen),
		sleepNotify: make(chan bool),
	}
}

func (w *WorkPool) PushTask(t *Task) {
	w.pool <- t
}

func (w *WorkPool) PushTaskFunc(f TaskFunc, args ...interface{}) {
	w.pool <- &Task{
		f:    f,
		args: args,
	}
}

func (w *WorkPool) work(i int) {
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case <-w.sleepCtx.Done():
			time.Sleep(time.Duration(w.sleepSeconds) * time.Second)
		case t := <-w.pool:
			flag := t.Execute(w)
			if !flag {
				w.PushTask(t)
				fmt.Printf("work %v PushTask, pool length %v \n", i, len(w.pool))
			}
		}
	}
}

func (w *WorkPool) Start() *WorkPool {
	fmt.Printf("workpool run %d worker\n", w.workerCount)
	w.wg.Add(w.workerCount)
	w.stopCtx, w.stopCancelFunc = context.WithCancel(context.Background())
	w.sleepCtx, w.sleepCancelFunc = context.WithCancel(context.Background())
	go w.sleepControl()
	for i := 0; i < w.workerCount; i++ {
		go w.work(i)
	}
	return w
}

func (w *WorkPool) Stop() {
	w.stopCancelFunc()
	w.wg.Wait()
}

func (w *WorkPool) sleepControl() {
	fmt.Println("sleepControl start...")
	for {
		select {
		case <-w.stopCtx.Done():
			w.wg.Done()
			return
		case <-w.sleepNotify:
			fmt.Printf("receive sleep notify start...\n")
			w.sleepCtx, w.sleepCancelFunc = context.WithCancel(context.Background())
			w.sleepCancelFunc()
			fmt.Printf("sleepControl will start sleep %v s\n", w.sleepSeconds)
			time.Sleep(time.Duration(w.sleepSeconds) * time.Second)
			w.sleepSeconds = 0
			fmt.Println("sleepControl was end sleep")
		}
	}
}

func (w *WorkPool) SleepNotify(seconds int64) {
	if atomic.CompareAndSwapInt64(&w.sleepSeconds, 0, seconds) {
		fmt.Printf("sleepSeconds set %v\n", seconds)
		w.sleepNotify <- true
	}
}
