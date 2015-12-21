package network

import (
	"encoding/json"
	"fmt"
	"lib/log4go"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	//local package
	"common"
)

type Server struct {
	ip            string // ""
	port          int    //
	maxClients    int    //
	clientNum     int
	acceptTimeout time.Duration
	connTimeout   time.Duration
	slowread      time.Duration
	headerLen     int //10
	maxBodyLen    int //100k bytes
	running       bool
	logger        log4go.Logger
	heartBeat     int64
	workerNum     int
	currentWorker int
	lock          sync.Mutex
}

func NewServer() *Server {
	s := &Server{
		port:          common.GetConfInt("server", "port", 8888),
		maxClients:    common.GetConfInt("server", "max_clients", 10000),
		clientNum:     0,
		acceptTimeout: common.GetConfSecond("server", "accept_timeout", 60*5),
		connTimeout:   common.GetConfSecond("server", "connection_timeout", 60*3),
		slowread:      common.GetConfSecond("server", "slow_read", 0),
		headerLen:     common.GetConfInt("server", "header_length", 10),
		maxBodyLen:    common.GetConfInt("server", "max_body_length", 102400),
		running:       false,
		logger:        common.NewLogger("server"),
		heartBeat:     time.Now().Unix(),
		workerNum:     common.WorkerNum,
		currentWorker: -1,
	}

	if s == nil {
		fmt.Println("new server failed")
		return nil
	}

	if s.logger == nil {
		fmt.Println("New Server logger failed")
		return nil
	}

	s.ip = common.GetConfString("server", "bind", "")

	// rlimit 是什么限制
	max_clients := uint64(s.maxClients)
	if max_clients > 1024 {
		var rlim syscall.Rlimit
		err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
		if err != nil {
			fmt.Println("Server get rlimit error" + err.Error())
			return nil
		}

		rlim.Cur = max_clients
		rlim.Max = max_clients
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
		if err != nil {
			fmt.Println("server set rlimit error: " + err.Error())
			return nil
		}
		s.logger.Info("set fd limit to %d", s.maxClients)
	}

	return s
}

func (this *Server) Start() {
	addr := fmt.Sprintf("%s:%d", this.ip, this.port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		//
		fmt.Println("Server Listen error:" + err.Error())
		return
	}
	defer ln.Close() //listener也需要close

	this.running = true
	this.currentWorker = time.Now().Nanosecond() % this.workerNum
	this.logger.Info("Tcp server start to listen on %s", addr)

	for this.running {
		conn, err := ln.Accept()
		fmt.Println("has a connection")
		if err != nil {
			fmt.Println("Accept " + err.Error())
			os.Exit(1)
		}
		go this.process(conn)
	}
}

func (this *Server) process(conn net.Conn) {
	if conn == nil {
		this.logger.Error("Input conn is nil")
		return
	}
	defer conn.Close()

	fmt.Printf("收到数据%s", time.Now())
	buff := make([]byte, 1000)
	conn.Read(buff)
	var r = make(map[string]interface{})
	err := json.Unmarshal(buff, &r)
	if err != nil {
		fmt.Println("Unmarshal err:%s", err.Error())
		//os.Exit(1)
	}

	addr := conn.RemoteAddr().String()

	fmt.Println("receive data from :" + addr)
	fmt.Println("data: ", string(buff))
}

func (this *Server) receive(buf []byte, conn net.Conn) ([]byte, error) {
	return nil, nil
}

func (this *Server) send(buf []byte, conn net.Conn) error {
	return nil
}

func (this *Server) Stop() {
	this.running = false
	this.logger.Info("Server Stop")
}
