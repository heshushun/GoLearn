package main

import "fmt"

func main()  {
	var a []int
	a = append(a, 1,2,3)
	a = append(a, []int{1,2,3}...)

	var numbers  []int
	for i:=0; i<10; i++{
		numbers = append(numbers, i)
		fmt.Printf("len: %d  cap: %d pointer: %p\n", len(numbers), cap(numbers), numbers)
	}

	var b = []int{1,2,3}
	fmt.Println(b)

	b = append([]int{0}, b...) // 在开头添加1个元素
	fmt.Println(b)

	b = append([]int{-3,-2,-1}, b...)  // 在开头添加1个切片
	fmt.Println(b)

	b = append(b[:2], append([]int{999}, a[2:]...)...) // 在第2个位置插入999
	fmt.Println(b)

}
