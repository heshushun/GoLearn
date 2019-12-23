package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 数据生产者
func producer(header string, channel chan<- string)  {
	// 不停地生产数据
	for {
		// 将随机数和字符串格式化为字符串发送给通道
		channel <- fmt.Sprintf("%s: %v", header, rand.Int31())
		time.Sleep(time.Second)
	}

}

// 数据消费者
func customer(channel <-chan string)  {
	// 不停地消费数据
	for {
		// 从通道中取数据，此处会阻塞直到信道中返回数据
		message := <- channel
		fmt.Println(message)
	}
}

func main()  {
	channel := make(chan string)

	go producer("cat", channel)
	go producer("dog", channel)

	customer(channel)
}