package calc

import (
	"fmt"
	"net"
	"testing"
)

func TestWork_Pool(t *testing.T) {

	l, e := net.Listen("tcp", ":3207")
	if e != nil {
		fmt.Println(e)
		return
	}

	//创建dispatcher
	dispatcher := NewDispatcher(MaxWorkers)
	dispatcher.Run()
	//初始化工作队列
	JobQueue = make(chan Job, MaxQueue)

	defer l.Close()
	defer dispatcher.Stop()

	for {
		//接受客户端的连接
		conn, err := l.Accept()
		if err != nil {
			return
		}

		job := Job{
			Connection: conn,
		}
		//客户端连接放入工作队列
		JobQueue <- job
	}

}
