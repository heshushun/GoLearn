package main

import (
	"fmt"
	"strings"
)

func main() {
	var s = "Goodbye,world!"

	fmt.Println(strings.Trim(s, "!"))
	fmt.Println(strings.TrimRight(s, "!"))
	fmt.Println(strings.TrimLeft(s, "!"))

	fmt.Println(strings.TrimPrefix(s, "Go"))
	fmt.Println(strings.TrimSuffix(s, "d!"))

	fmt.Println(strings.TrimSpace("Goodbye world!"))
	fmt.Println(strings.TrimPrefix(s, "Goodbye"))
	fmt.Println(strings.TrimPrefix(s, "Howdy"))
}
