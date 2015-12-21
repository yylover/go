package main

import (
	"fmt"
	"os"
	"time"
	"net"
)

type Book struct {
	id string;
	name string
}


func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage :%s host:port", os.Args[0])
		os.Exit(1)
	}

	service := os.Args[1]

	for {
		conn, err := net.Dial("tcp", service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Dial error :%s", err.Error())
			os.Exit(1)
		}
		_, err = conn.Write([]byte("Hello world TT"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write Error, %s", err.Error())
			os.Exit(1)
		}
		fmt.Println("发送数据:", string([]byte("hello")))
		time.Sleep(time.Microsecond*10);
		conn.Close()
	}
}
