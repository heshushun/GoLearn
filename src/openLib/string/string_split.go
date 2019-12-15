package main

import (
	"fmt"
	"strings"
)

func main()  {
	fmt.Printf("%q\n", strings.Split("a,b,c", ","))
	fmt.Printf("%q\n", strings.Split("ABC", ""))

	s := []string{"foo", "bar", "baz"}
	fmt.Println(strings.Join(s,",")) //连接

	fmt.Println("ba" + strings.Repeat("na", 2))
}
