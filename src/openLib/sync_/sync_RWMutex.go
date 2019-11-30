package main

import (
	"fmt"
	"sync"
	"time"
)

/**
RWMutex 比 Mutex 多了一个“读锁定”和“读写解锁”，可以让多个例程同时读取某对象。
 */

var m *sync.RWMutex

func main() {
	m = new(sync.RWMutex)

	go read(1)
	go read(2)

	time.Sleep(2 * time.Second)
}

func read(i int)  {
	fmt.Println(i, "start")

	m.RLock()
	fmt.Println(i, "reading ......")
	time.Sleep(1 * time.Second)
	m.RUnlock()

	fmt.Println(i, "end")
}
