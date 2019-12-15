package main

import (
	"fmt"
	"strings"
)

func main() {
	value := "Your cat is cute"
	fmt.Println(value)
	result := strings.Replace(value, "cat", "dog", -1)
	fmt.Println(result)

	value = "bird bird bird"
	fmt.Println(value)
	result = strings.Replace(value, "bird", "fish", 1)
	fmt.Println(result)

}
