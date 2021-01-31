package main

import (
	"fmt"
	"reflect"
)

type People struct {
	name string
	age int
}
func (p People)SayBye() {
	fmt.Println("Bye")
}
func (p People)SayHello() {
	fmt.Println("Hello")
}

func main() {
	var age interface{} = 20
	fmt.Println("age: ", age)

	// 1、TypeOf ValueOf 从接口变量到反射对象
	t := reflect.TypeOf(age)
	v := reflect.ValueOf(age)
	fmt.Printf("TypeOf %T \n", t)
	fmt.Printf("ValueOf %T \n", v)
	// TypeOf *reflect.rtype
	// ValueOf reflect.Value

	// 2、Interface 从反射对象到接口变量
	v = reflect.ValueOf(age)
	i := v.Interface()
	fmt.Printf("Interface %T %v\n", i, i)
	// Interface int 20

	// 3、Elem 获取指针指向的数据（用来修改反射数据）
	v1 := reflect.ValueOf(&age)
	fmt.Println("v1 可写性为:", v1.CanSet())
	v2 := v1.Elem()
	fmt.Println("v2 可写性为:", v2.CanSet())
	// v1 可写性为: false
	// v2 可写性为: true

	// 4、SetString() 设置值
	name := "hello"
	v1 = reflect.ValueOf(&name)
	v2 = v1.Elem()
	v2.SetString("world")
	fmt.Printf("SetString %T %v\n", name, name)
	// SetString string world

	// 5、Kind() 获取类别（相对Type 更广义）
	m := People{}
	t = reflect.TypeOf(m)
	fmt.Printf("Type: %v\n",t)
	fmt.Printf("Kind: %v\n",t.Kind())
	// Type: main.People
	// Kind: struct

	// 6、Int()、Float()、String()、Bool()、Interface() 类型转换
	age = 88
	ageV := reflect.ValueOf(age)
	fmt.Printf("Int 转换前， type: %T, value: %v \n", ageV, ageV)
	ageV2 := ageV.Int()
	fmt.Printf("Int 转换后， type: %T, value: %v \n", ageV2, ageV2)
	// Int 转换前， type: reflect.Value, value: 88
	// Int 转换后， type: int64, value: 88

	// 7、NumField() 获取结构体属性个数
	p := People{name: "hss", age: 27}
	pv := reflect.ValueOf(p)
	fmt.Printf("NumField: %v\n", pv.NumField())
	// NumField: 2

	// 8、Field() 获取结构体属性
	p = People{name: "hss", age: 27}
	pv = reflect.ValueOf(p)
	fmt.Printf("Field 1: %v\n", pv.Field(0))
	fmt.Printf("Field 2: %v\n", pv.Field(1))
	// Field 1: hss
	// Field 2: 27

	// 9、NumMethod() 获取结构体方法个数
	p = People{name: "hss", age: 27}
	pv = reflect.ValueOf(p)
	fmt.Printf("NumMethod: %v\n", pv.NumMethod())
	// NumMethod: 2

	// 10、Method() 获取结构体方法
	p = People{name: "hss", age: 27}
	tp := reflect.TypeOf(p)
	fmt.Printf("Method 1: %v\n", tp.Method(0).Name)
	fmt.Printf("Method 2: %v\n", tp.Method(1).Name)
	// Method 1: SayBye
	// Method 2: SayHello


}
