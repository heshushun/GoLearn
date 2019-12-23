package main

import "fmt"

func main()  {
	b := 3
	if a := b; a > 2 {
		fmt.Println("a > 2")
	}else if a < 2 {
		fmt.Println("a < 2")
	}else {
		fmt.Println("a == 2")
	}
}
