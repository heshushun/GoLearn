package main

import (
	"fmt"
	"net/http"
)

func main()  {
	// 注册路由
	http.HandleFunc("/", func(write http.ResponseWriter, request *http.Request) {
		_, _ = write.Write([]byte("Hello World!"))
	})
	// 服务监听
	if err := http.ListenAndServe(":8080", nil); err != nil{
		fmt.Println("start http server fail:", err)
	}
}