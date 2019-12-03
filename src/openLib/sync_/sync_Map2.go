package main

import (
	"fmt"
	"sync"
)

func main() {

	var m sync.Map

	// Store
	m.Store(1,"a")
	m.Store(2,"b")

	// LoadOrStore
	v, ok := m.LoadOrStore("1", "aaa")
	fmt.Println(ok, v)
	v, ok = m.LoadOrStore(1, "aaa")
	fmt.Println(ok, v)

	//Load
	v, ok = m.Load(1)
	if ok{
		fmt.Println("it's an existing key, value is ", v)
	}else {
		fmt.Println("it's an unknown key")
	}

	// Range  要求参数是 func
	f := func(k, v interface{}) bool{
		fmt.Println(k, v)
		return true
	}
	m.Range(f)

	//Delete
	m.Delete(1)
	fmt.Println(m.Load(1))

}