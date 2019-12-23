package main

import (
	"fmt"
	"reflect"
)

type Foo struct {
	FirstName string
	LastNme string
	Age int
}

func (f *Foo)reflect() {
	//  reflect.Elem() 方法获取这个指针指向的元素类型,被称为 取元素
	val := reflect.ValueOf(f).Elem()

	fmt.Println(val)

	// NumField()获取字段数量；  Field(i)根据索引获取字段
	for i := 0; i<val.NumField();i++{
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		fmt.Printf("Field: %s, Value: %v\n", typeField.Name, valueField.Interface())
	}
}

func main()  {
	f := &Foo{
		FirstName: "he",
		LastNme: "shushun",
		Age: 30,
	}

	f.reflect()
}