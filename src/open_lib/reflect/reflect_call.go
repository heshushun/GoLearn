package main

import (
	"fmt"
	"reflect"
)

// 通过反射动态调用

func Method(in interface{}) (ok bool) {
	v := reflect.ValueOf(in)

	fmt.Println(v.Kind())
	// Kind()获取类型
	if v.Kind() == reflect.Slice{
		ok = true
	}else {
		panic("error")
	}

	num := v.Len()
	for i := 0; i< num; i++{
		fmt.Println(v.Index(i).Interface())
	}
	return ok
}

func main()  {
	s := []int{1, 3, 5, 7, 9}
	b := []float64{1.2, 2.4, 9.8, 7.9}
	Method(s)
	Method(b)
}