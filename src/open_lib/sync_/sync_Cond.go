package main

import (
	"fmt"
	"sync"
	"time"
)

/**
Cond用于在并发环境下routine的等待和通知
 */

func main(){
	m := sync.Mutex{}
	m.Lock()
	c := sync.NewCond(&m)

	go func() {
		m.Lock()
		defer m.Unlock()
		fmt.Println("3.goroutine is owner of lock")
		time.Sleep(1 * time.Second)
		c.Broadcast() // 唤醒所有等待的 wait
		fmt.Println("4.goroutine will release lock soon (deffer Unlock)")
	}()

	fmt.Println("1.main goroutine is owner of lock")
	time.Sleep(1 * time.Second)
	fmt.Println("2.main goroutine is still lock")
	c.Wait()
	m.Unlock()
	fmt.Println("Done")
}
