package main

import (
	"fmt"
	"reflect"
)

type Test struct {
	num int
}

func main ()  {
	test := new(Test)
	test.num = 110
	rev := reflect.ValueOf(test)
	ty := reflect.Indirect(rev).Type()
	fmt.Println(ty.Kind().String())
}
