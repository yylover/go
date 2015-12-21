package main

import (
	"fmt"
	"os"
	"net"
	"io/ioutil"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	
	//解析ip地址
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err);
	fmt.Println(tcpAddr)

	//ip访问
	conn, err := net.DialTCP("tcp4", nil,  tcpAddr)
	checkError(err)

	//写入数据
	_, err = conn.Write([]byte("HEAD / HTTP/1.0\r\n\r\n"))
	checkError(err)

	//读取数据
	result, err := ioutil.ReadAll(conn)
	checkError(err)

	fmt.Println(string(result))

	os.Exit(0)

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error: %s", err.Error())
		os.Exit(1)
	}
}
