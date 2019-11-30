package main

import (
	"fmt"
	"sync"
)

/**
互斥锁用来保证在任一时刻，只能有一个例程访问某对象。Mutex 的初始值为解锁状态。
 */

type  SafeInt struct {
	sync.Mutex
	Num int
}

func main() {
	count := SafeInt{}
	done := make(chan bool)
	for i := 0; i<10; i++ {
		go func(i int) {
			count.Lock()
			count.Num += i
			fmt.Print(count.Num, "  ")
			count.Unlock()
			done <- true
		}(i)
	}

	for i := 0; i<8; i++ {
		<- done
	}
}
