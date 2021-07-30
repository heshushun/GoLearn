package calc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const MaxWorkers = 100000

const MaxQueue = 3000

const DataLen = 4

//协程池的最小工作单元，即具体业务处理结构体
type Job struct {
	Connection net.Conn //客户端的连接
}

//队列，用来接收、发送请求
var JobQueue chan Job

//用于执行job，可以理解为job的管理者
type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

//初始化Worker
func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

//运行Worker
func (w Worker) Start() {
	go func() {
		for {
			//将可用的worker放进队列中
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				//接收到具体请求时进行处理
				HandleConnection(job.Connection)
			case <-w.quit:
				//接收停止请求
				return
			}
		}
	}()
}

//发送停止请求
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

//解包
func Unpack(buffer []byte, readerChannel chan []byte) []byte {
	length := len(buffer)

	var i int
	for i = 0; i < length; i++ {
		if length < i+DataLen {
			break
		}
		//根据长度来获取数据
		messageLen := BytesToInt(buffer[i : i+DataLen])
		if length < i+DataLen+messageLen {
			break
		}
		data := buffer[i+DataLen : i+DataLen+messageLen]
		readerChannel <- data

		i += DataLen + messageLen - 1
	}

	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

//处理客户端请求
func HandleConnection(conn net.Conn) {
	defer func() {
		fmt.Println(conn.RemoteAddr())
		conn.Close()
	}()
	tempBuffer := make([]byte, 0)
	readerChannel := make(chan []byte, 16)
	//fmt.Println(conn.RemoteAddr())
	go reader(readerChannel)

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return
		}
		tempBuffer = Unpack(append(tempBuffer, buffer[:n]...), readerChannel)
	}
}

func reader(readerChannel chan []byte) {
	for {
		select {
		case data := <-readerChannel:
			//fmt.Println(string(data))
			data = data
		}
	}
}
