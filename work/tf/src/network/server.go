package network

import (
	"encoding/json"
	"fmt"
	"lib/log4go"
	"net"
	"os"
	"strconv"
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
	defer ln.Close() //listener也需要close?

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

	buff := make([]byte, this.headerLen+this.maxBodyLen)

	var (
		num  int
		data []byte
		err  error
		echo []byte = []byte("ok")
	)

	this.currentWorker = (this.currentWorker + 1) % this.workerNum
	//outChan := common.PacketChans[this.currentWorker]

	addr := conn.RemoteAddr().String()
	this.logger.Info("Accept connection from %s, total client number now :%d", addr, this.clientNum)

	for this.running {
		if this.slowread > 0 { // 慢读写
			time.Sleep(this.slowread)
		}

		//header
		data, err = this.receive(buff[0:this.headerLen], conn)
		if err != nil {
			this.logger.Error("Read Header form %s failed, error: %s", addr, err.Error())
			break
		}

		num, err = strconv.Atoi(string(data))
		if err != nil || num <= 0 || num > this.maxBodyLen {
			this.logger.Error("Read header from %s format error, header content:%s", addr, string(data))
			break
		}

		//read body
		num += this.headerLen
		data, err = this.receive(buff[this.headerLen:num], conn)
		if err != nil {
			this.logger.Error("Read body from %s failed, error%s", addr, err.Error())
			break
		}

		//read success
		this.logger.Info("Read from %s, length:%d, Content:%s", addr, num, string(data))

		//ack ok back
		err = this.send(echo, conn)
		if err != nil {
			this.logger.Error("Echo back %s to %s failed, error:%s", string(echo), addr, err.Error())
			break
		}

		//为什么复制一次
		/*data = make([]byte, num)
		copy(data, buff[0:num])
		select {
		case outChan <- data:
			this.logger.Debug("Insert into outchan success, bytes:%s", string(data))
		default:
			this.logger.Error("Insert into outChan failed, channel length:%d to more, drom packet:%s", len(outChan), string(data))
			//报警TODO
			time.Sleep(time.Second)
			break
		}*/
	}

	this.lock.Lock()
	this.clientNum -= 1
	this.lock.Unlock()

	this.logger.Info("connection %s closed, remaining total clients number:%d", addr, this.clientNum)

	var r = make(map[string]interface{})
	err = json.Unmarshal(data, &r)
	if err != nil {
		fmt.Println("Unmarshal err:%s", err.Error())
		//os.Exit(1)
	}

	fmt.Println("receive data from :" + addr)
	fmt.Println("data: ", string(buff))
}

func (this *Server) receive(buf []byte, conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(this.connTimeout))
	readNum := 0
	length := len(buf)

	var num int
	var err error
	for readNum < length {
		num, err = conn.Read(buf[readNum:])
		if err != nil {
			this.logger.Debug("read conn error:%s, close connection, already read %d bytes", err.Error(), readNum)
			return nil, err
		}
		this.logger.Debug("Read %d bytes: %s", num, string(buf))
		readNum += num
	}

	return buf, nil
}

func (this *Server) send(buf []byte, conn net.Conn) error {
	conn.SetWriteDeadline(time.Now().Add(this.connTimeout))
	num, err := conn.Write(buf)
	if err != nil {
		return err
	}

	if num <= 0 {
		return fmt.Errorf("connection write back %d bytes error", num)
	}

	return nil

}

func (this *Server) Stop() {
	this.running = false
	this.logger.Info("Server Stop")
}
