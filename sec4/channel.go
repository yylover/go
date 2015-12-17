package main

import (
	"fmt"
)

func Count(ch chan int) {
	ch <- 1
	fmt.Println("counting")
}

func main() {
	chs := make([]chan int, 10)
	for i := 0; i < 10; i++ {
		chs[i] = make(chan int)
		go Count(chs[i])
	}

	for _, ch := range chs {
		t := <-ch
	}
}

timeout := make(chan int, 1)
go func() { //匿名等待函数
	time.Sleep(1e9) //1秒钟
	timeout <- true
}

select {
case <-ch:
	//ch中读取数据
case <- timeout:
	//超时处理
}

func Parse(ch <-chan int) {
	for value := range ch {
		fmt.Println("Parsing value", value)
	}
}

type PipeData struct {
	value int
	handler func(int) int
	next chan int
}

func handle(queue chan *PipeData) {
	for data := range queue {
		data.next <-data.handler(data.value)
	}
}

