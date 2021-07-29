package calc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TaskFunc func(args ...interface{}) error

type Task struct {
	f    TaskFunc
	args []interface{}
}

func NewTask(f TaskFunc, args ...interface{}) *Task {
	t := Task{
		f:    f,
		args: args,
	}
	return &t
}

func (t *Task) Execute() error {
	err := t.f(t.args...)
	return err
}

type WorkerRecover struct {
	recoverList []int
	lock        sync.RWMutex
}

type Pool struct {
	//协程池worker数量 (Goroutine个数)
	workerNum int

	//协程池任务队列
	JobsChannel chan *Task

	workerRecover *WorkerRecover

	stopCtx        context.Context
	stopCancelFunc context.CancelFunc
	wg             sync.WaitGroup
}

// 创建一个协程池
func NewPool(num int, poolLen int) *Pool {
	p := Pool{
		workerNum:     num,
		JobsChannel:   make(chan *Task, poolLen),
		workerRecover: &WorkerRecover{},
		wg:            sync.WaitGroup{},
	}
	return &p
}

// 加一个任务
func (p *Pool) pushTask(t *Task) {
	select {
	case p.JobsChannel <- t:
	default: // 阻塞了，重新启动worker
		p.Restart()
		time.Sleep(10 * time.Millisecond)
		p.JobsChannel <- t
	}
}

// 协程池创建一个worker并且开始工作
func (p *Pool) worker(workerID int) {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()
	for {
		select {
		case <-p.stopCtx.Done():
			p.wg.Done()
			fmt.Println("协程池Pool停止")
			return
		case task := <-p.JobsChannel:
			err := task.Execute()
			if err != nil {
				p.pushTask(task)
				continue
			}
			fmt.Println("worker ID ", workerID, " 执行 task ID", task.args[0].(int), "任务完毕")
			// 重新设置等待15s
			ticker.Reset(time.Second * 15)
		case <-ticker.C: // 15s 都收不到任何处理就回收
			p.workerRecover.lock.Lock()
			if p.workerNum-len(p.workerRecover.recoverList) > 1 {
				fmt.Println("worker ID ", workerID, " 超时回收")
				p.workerRecover.recoverList = append(p.workerRecover.recoverList, workerID)
			} else {
				ticker.Reset(time.Second * 15)
			}
			p.workerRecover.lock.Unlock()
		}
	}
}

// 协程池Pool开始工作
func (p *Pool) Start() {
	p.wg.Add(p.workerNum)
	p.stopCtx, p.stopCancelFunc = context.WithCancel(context.Background())
	for i := 0; i < p.workerNum; i++ {
		go p.worker(i)
	}
}

// 重启回收的worker
func (p *Pool) Restart() {
	p.workerRecover.lock.Lock()
	defer p.workerRecover.lock.Unlock()
	for _, workerID := range p.workerRecover.recoverList {
		fmt.Println("worker ID ", workerID, " 重开")
		go p.worker(workerID)
	}
}

// 协程池Pool停止工作
func (p *Pool) Stop() {
	p.stopCancelFunc()
	p.wg.Wait()
}
