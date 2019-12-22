package main

import "fmt"

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
	fmt.Println("keys: ",keys)
}
