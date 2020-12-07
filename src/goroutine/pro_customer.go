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

func producer2(c chan<- int){
	defer close(c) // 生产结束关闭channel

	for i := 0; i < 10; i++{
		c <- i // 阻塞，直到消费用完数据c
		fmt.Println("生产 ",i) // 阻塞，直到生产加入数据c
	}
}

func customer2(c <-chan int, f chan <- int)  {
	for {
		if v, ok := <-c ;ok {
			fmt.Println("消费 ",v) // 阻塞，直到生产加入数据c
			//time.Sleep(1 * time.Second)
		}else {
			break
		}
	}
	f <- 1 // 标识消费完了
}

func main()  {
	//channel := make(chan string)
	//
	//go producer("cat", channel)
	//go producer("dog", channel)
	//
	//customer(channel)

	buf := make(chan int)
	// buf := make(chan int, 10)
	flg := make(chan int)
	go producer2(buf)
	go customer2(buf, flg)

	<- flg // 表面消费结束

}