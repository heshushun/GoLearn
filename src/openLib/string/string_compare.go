package main

import (
	"fmt"
	"strings"
)

func main() {
	var s1 string = "Welcome to The WORld of go!"
	var s2 string = "Welcome go The WORld of go!"

	fmt.Println(strings.Compare(s1, s2)) // s1 > s2 返回值为int型，1

	fmt.Println(strings.EqualFold("Go", "go")) //返回值为bool型， true
}
