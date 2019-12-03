package main

import (
	"fmt"
	"reflect"
)

func main()  {
	s := "I am string"

	// 对象类型
	fmt.Println(reflect.TypeOf(s))

	// 对象值
	fmt.Println(reflect.ValueOf(s))

	x  := 3.4
	fmt.Println(reflect.TypeOf(x))
}
