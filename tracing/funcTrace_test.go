package tracing

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_FuncTrace(t *testing.T) {
	// A1()
	// 起一个 goroutine 跑A1
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		A1()
		wg.Done()
	}()
	time.Sleep(time.Second)
	A2()
	wg.Wait()
}

func Test_FuncTraceRandom(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			Begin()
			wg.Done()
		}()
		time.Sleep(time.Millisecond *10)
	}
	wg.Wait()
}

func Test_FuncTraceRoute(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			Begin()
			wg.Done()
		}()
		time.Sleep(time.Millisecond *10)
	}
	wg.Wait()

	for gID, nodes := range routes {
		out := fmt.Sprintf("g[%02d] ", gID)
		for _, name := range nodes {
			out = out + " -> " + name
		}
		fmt.Println(out)
	}

	//rand.Seed(time.Now().Unix())
	//random := rand.Intn(2)+1
	//fmt.Println(random)
}