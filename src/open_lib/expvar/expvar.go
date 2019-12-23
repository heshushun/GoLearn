package main

import (
	"expvar"
	"net"
	"net/http"
)
/**
监控变量 访问: http://localhost:8000/debug/vars
 */
var (
	test = expvar.NewMap("Test")
)

func init () {
	test.Add("go", 10)
	test.Add("go1", 10)
}

func main(){
    sock, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic("error")
	}
	go func() {
		http.Serve(sock, nil)
	}()

	select {}
}