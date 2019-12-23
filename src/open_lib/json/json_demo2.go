package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)
// Encode是编码器上的一个方法，它将JSON编码的Go类型写入输出流
func main() {
	const jsonStream = `
        {"Name": "Ed", "Text": "Knock knock."}
        {"Name": "Sam", "Text": "Who's there?"}
        {"Name": "Ed", "Text": "Go fmt."}
        {"Name": "Sam", "Text": "Go fmt who?"}
        {"Name": "Ed", "Text": "Go fmt yourself!"}
    `
	type Message struct {
		Name string
		Text string
	}

	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m Message
		err := dec.Decode(&m)
		// 错误信息处理
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.Name, m.Text)
	}
}