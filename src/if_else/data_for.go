package main

import "fmt"

func main()  {
	// for后的表达式也可以省略圆括号
	for i, max := 0, 3; i < max; i++{
		fmt.Println(i)
	}

	// 使用for代替while
	var x int
	for x < 5{
		fmt.Println(x)
		x++
	}

}