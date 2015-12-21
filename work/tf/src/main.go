package main

import (
	"fmt"
	"flag"
	"os"
	"time"
	"net"
	//"runtime"
	"encoding/json"
)

const (
	VERSION string = "tf_v1.0"
)

func initialize() error {
	fmt.Println("tf initializing ....")

	confFile := flag.String("c", "./etc/tf.conf", "config file name")
	if confFile == nil {
		return fmt.Errorf("no config file")
	}

	fmt.Println("has config file")
	return nil
}

func main() {

	if err := initialize(); err != nil {
		fmt.Sprintf("init failed %s \n", err.Error())
	
	}

	ln, err := net.Listen("tcp", ":5986")
	if err != nil {
		//
		fmt.Println("Listen error:" + err.Error())
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		fmt.Println("has a connection")
		if err != nil {
			fmt.Println("Accept " + err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Printf("收到数据%s", time.Now())
	buff := make([]byte, 1000)
	conn.Read(buff)
	var r = make(map[string] interface{})
	err := json.Unmarshal(buff, &r)
	if err != nil {
		fmt.Println("Unmarshal err:%s", err.Error())
		//os.Exit(1)
	}

	addr := conn.RemoteAddr().String()

	fmt.Println("receive data from :" + addr)
	fmt.Println("data: ",string(buff))
}
