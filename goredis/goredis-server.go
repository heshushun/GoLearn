package main

import (
	"log"
	"net"
)

func main() {
	netListen, err := net.Listen("tcp", "127.0.0.1:9736")
	if err != nil {
		log.Print("listen err ")
	}

	defer netListen.Close()

	for {
		conn, err := netListen.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}
