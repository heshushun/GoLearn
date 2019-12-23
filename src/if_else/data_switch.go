package main

import "fmt"

// 不需要显示地写 break
// 如果不想break，可以使用fallthrough关键字来向下继续执行case
func main()  {
	a, b, c := 1, 2, 3
	x := 2
	switch x {
	case a, b:
		fmt.Println("a|b")
	case c:
		fmt.Println("c")
	default:
		fmt.Println("default")
	}
}
