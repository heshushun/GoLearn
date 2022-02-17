package main

import (
	"GoLearn/goredis/core/proto"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Hi Godis")

	IPPort := "127.0.0.1:9736"
	reader := bufio.NewReader(os.Stdin)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", IPPort)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	defer conn.Close()
	//log.Println(tcpAddr, conn.LocalAddr(), conn.RemoteAddr())

	for {
		fmt.Print(IPPort + "> ")
		res, _, _ := reader.ReadLine()
		msg := strings.TrimSpace(string(res))

		if len(strings.Fields(msg)) == 0 || helpCmd(strings.Fields(msg)) {
			continue
		}

		_, err1 := send2Server(msg, conn)

		n, resp, err2 := replyFromServer(conn)

		if err1 != nil {
			fmt.Println(IPPort+"> ", "err proto encode")
		} else if n == 0 {
			fmt.Println(IPPort+"> ", "nil")
		} else if err2 != nil {
			fmt.Println(IPPort+"> ", "err server response")
		} else {
			fmt.Println(IPPort+"> ", string(resp.Value))
		}
	}

}

func send2Server(msg string, conn net.Conn) (n int, err error) {
	encodeMsg, e := proto.EncodeCmd(msg)
	if e != nil {
		return 0, e
	}
	//fmt.Println("proto encode", encodeMsg, string(encodeMsg))
	n, err = conn.Write(encodeMsg)
	return n, err
}

func replyFromServer(conn net.Conn) (n int, resp *proto.Resp, err error) {
	buff := make([]byte, 1024)
	n, _ = conn.Read(buff)
	resp, err = proto.DecodeFromBytes(buff)
	return n, resp, err
}

func checkError(err error) {
	if err != nil {
		log.Println("err ", err.Error())
		os.Exit(1)
	}
}

func helpCmd(argv []string) (ret bool) {
	if argv[0] == "-v" || argv[0] == "--version" {
		version()
		ret = true
	}
	if argv[0] == "--help" || argv[0] == "-h" {
		usage()
		ret = true
	}
	return
}

func version() {
	fmt.Println("Goredis server v=0.0.1")
}

func usage() {
	fmt.Println("Usage: ./goredis-server [/path/to/redis.conf] [options]")
	fmt.Println("       > get hello")
	fmt.Println("       > set hello 123")
	fmt.Println("       > subscribe test")
	fmt.Println("       > publish test hello")
}
