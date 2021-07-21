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
