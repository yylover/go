package main

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

type Book struct {
	id   string
	name string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage :%s host:port", os.Args[0])
		os.Exit(1)
	}

	service := os.Args[1]

	i := 0
	for {
		conn, err := net.Dial("tcp", service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Dial error :%s", err.Error())
			os.Exit(1)
		}

		i++
		h := fmt.Sprintf("hello world %d", i)
		hello := []byte(h)
		_, err = conn.Write(hello)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Write Error, %s", err.Error())
			os.Exit(1)
		}
		fmt.Println(h)
		var rlim syscall.Rlimit
		err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
		fmt.Println(fmt.Sprintf("%d %d", rlim.Cur, rlim.Max))
		time.Sleep(time.Millisecond)
		conn.Close()
	}
}
