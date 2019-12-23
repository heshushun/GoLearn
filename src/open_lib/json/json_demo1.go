package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// marshall是返回Go类型的JSON编码的函数。
func main()  {
	type ColorGroup struct {
		ID int
		Name string
		Colors []string
	}

	group := ColorGroup{
		ID : 1,
		Name: "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}

	b, err := json.Marshal(group)
	if err != nil{
		fmt.Println("error", err)
	}

	os.Stdout.Write(b)

	var jsonBlob = []byte(`[
        {"Name": "Platypus", "Order": "Monotremata"},
        {"Name": "Quo",    "Order": "Dasyuromorphia"}
    ]`)

	// json 转为 对象
	type Animal struct {
		Name string
		Order string
	}
	var animals []Animal
	err = json.Unmarshal(jsonBlob, &animals)
	if err != nil{
		fmt.Println("error", err)
	}
	fmt.Println("\n ==================================================================")
	fmt.Printf("%+v", animals)
}
