package main

import (
	"fmt"
	"reflect"
)

// 通过反射来修改对象

func main(){
	y := 110
	fmt.Println("1: ", y)

	v := reflect.ValueOf(&y)
	v.Elem().SetInt(119)

	fmt.Println("2: ", y)
}
