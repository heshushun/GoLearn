package main

import (
	"fmt"
	"sync"
	"time"
)

/**
Once 的作用是多次调用但只执行一次，Once 只有一个方法，Once.Do()，向 Do 传入一个函数
 */

func main ()  {

	o := &sync.Once{}
	go myfun(o)
	go myfun(o)
	time.Sleep(time.Second * 2)

}

func myfun(o *sync.Once){
	fmt.Println("begin")
	o.Do(func() {
		fmt.Println("Working ...")
	})
	fmt.Println("end")
}
