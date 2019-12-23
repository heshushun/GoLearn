package main

import (
	"fmt"
	"sync"
)

/**
Pool 用于存储临时对象，它将使用完毕的对象存入对象池中，在需要的时候取出来重复使用，
目的是为了避免重复创建相同的对象造成 GC 负担过重
 */

func main() {
	p := &sync.Pool{
		New : func() interface{} {
			return 2
		},
	}

	a := p.Get().(int) // 取
	p.Put(100)	   // 存
	b := p.Get().(int)

	fmt.Println(a, b)
}
