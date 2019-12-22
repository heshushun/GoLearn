package main

import "fmt"

func main()  {
	var x [5]int
	x[4] = 100
	fmt.Println(x)

	total := 0
	for i:=0; i<len(x); i++{
		total += x[i]
	}
	fmt.Println(total/len(x))

	x = [5]int{23,34,87,90,12}
	total2 := 0
	for _, value := range x{
		total2 += value
	}
	fmt.Println(total2/len(x))
}
