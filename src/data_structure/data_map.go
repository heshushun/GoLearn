package main

import (
	"fmt"
	"sort"
)

func main()  {
	ids := map[string]int{}
	ids["steve"] = 10
	ids["mark"] = 20
	ids["adan"] = 30
	fmt.Println(len(ids))

	delete(ids, "steve")
	fmt.Println(len(ids))

	animals := map[string]string{}
	animals["cat"] = "Mittens"
	animals["dog"] = "Spot"

	keys := make([]string, 10)
	for key, value := range animals{
		fmt.Println(key, "=", value)
		keys = append(keys, key)
	}
	// 对切片进行排序
	sort.Strings(keys)
	fmt.Println("keys: ",keys)

	// map没有提供清空的函数、方法，清空map的方式就是重新make一个新的map
	fmt.Println(animals)
	animals = make(map[string]string)
	fmt.Println(animals)

}
