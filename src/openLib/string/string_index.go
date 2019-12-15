package main

import (
	"fmt"
	"strings"
)

func main() {
	s := "hello go world"

	fmt.Println(strings.Contains(s, "hell"))

	fmt.Println(strings.Index(s, "o"))

	fmt.Println(strings.LastIndex(s, "o"))

	fmt.Println(strings.Count(s, "o"))
}
