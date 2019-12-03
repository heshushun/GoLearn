package main

import (
	"fmt"
	"strconv"
)

func main()  {
	// AppendBool 将各种类型转换为字符串后追加到 dst 尾部
	b := []byte("bool: ")
    b = strconv.AppendBool(b, true)
    fmt.Println(string(b))

    // FormatBool 将各种类型转换为字符串
    v_bool := true
    s := strconv.FormatBool(v_bool)
    fmt.Printf("%T: %v\n", s, s)

    v_int := int64(42)
    s10 := strconv.FormatInt(v_int, 10)
    fmt.Printf("%T: %v\n", s10, s10)

}
