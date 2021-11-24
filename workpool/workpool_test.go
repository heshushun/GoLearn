package workpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 这里模拟ServerH服务的限流操作
var serverh = &server{max: 10, interval: 5}

type server struct {
	count    int
	max      int
	lasttime time.Time
	interval int64
	mu       sync.Mutex
}

func (s *server) Access(i int) bool {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	if s.lasttime.Unix() <= 0 || s.count >= s.max {
		if now.After(s.lasttime) {
			s.count = 1
			s.lasttime = time.Unix(now.Unix()+s.interval, 0)
			return true
		}
		fmt.Printf("Access false,i=%d \n", i)
		return false
	} else {
		s.count++
		fmt.Printf("Access true,i=%d s.count %d\n", i, s.count)
		return true
	}
}

// 这里是笔者服务的逻辑
func TestWorkPool_Start(t *testing.T) {
	wp := NewWorkPool(3, 100).Start()
	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		wp.PushTaskFunc(func(w *WorkPool, args ...interface{}) bool {
			if !serverh.Access(args[0].(int)) {
				// 发送睡眠5秒的通知
				w.SleepNotify(5)
				// 此次未执行成功，要将该任务放回协程池
				return false
			}
			return true
		}, i)
	}
	time.Sleep(100 * time.Second)
}

var closeChan chan bool
var resultQueue chan int

func receiveAntiPlayResult() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		<-t.C
		select {
		case <-closeChan:
			println("!!!!! close2")
			close(resultQueue)
			return
		case resultQueue <- 9: // 接收
		}
	}
}

// 反作弊战斗结果 统一分发
func dispatchAntiPlayResult() {
	for {
		select {
		//case <-closeChan:
		//	println("!!!!! close3")
		//	return
		case r, ok := <-resultQueue: // 分发
			if !ok {
				println("!!!!! close3")
				return
			}
			if ok && r == 9 {
				println("!!!!! ", r)
			}
		default:

		}
	}
}

//
func TestWorkPool_ooo(t *testing.T) {
	closeChan = make(chan bool)
	resultQueue = make(chan int, 1024)
	go receiveAntiPlayResult()
	time.Sleep(1 * time.Second)
	go dispatchAntiPlayResult()
	time.Sleep(1 * time.Second)
	closeChan <- true
	time.Sleep(5 * time.Second)
}
