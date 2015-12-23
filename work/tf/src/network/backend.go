package network

import (
	"common"
	"fmt"
	"io"
	"lib/log4go"
	"net"
	"strings"
	"time"
)

type BackPoint struct {
	name           string
	backNum        int
	backAddrs      []string
	currentBackIdx int
	retryInterval  time.Duration
	conn           net.Conn
	connTimeout    time.Duration
	sendTimeout    time.Duration
	recvBuf        []byte
	retryTimes     int
	sendingChan    chan []byte
}

type Backend struct {
	index       int           //index
	backendNum  int           //backend number
	backendList []*BackPoint  //bakcend lsit
	heartbeat   int64         //
	logger      log4go.Logger //
	inChan      chan []byte   //
	running     bool          //running
}

func NewBackend(idx int, inputChan chan []byte) *Backend {
	b := &Backend{
		index:      idx,
		backendNum: 0,
		heartbeat:  time.Now().Unix(),
		logger:     common.NewLogger("backend"),
		inChan:     inputChan,
		running:    false,
	}

	if b == nil {
		fmt.Println("New Backend failed")
		return nil
	}

	if b.logger == nil {
		fmt.Println("backend new logger failed")
		return nil
	}

	if b.inChan == nil {
		fmt.Println("InputChan init failed")
		return nil
	}

	options, err := common.Conf.Options("backend")
	if err != nil {
		fmt.Printf("backend get all options failed:5s \n", err.Error())
		return nil
	}
	fmt.Println(options)

	b.backendList = make([]*BackPoint, 0, 4)
	for _, option := range options {
		fmt.Println("option: " + option)
		if !strings.HasPrefix(option, "backend_list_") {
			continue
		}

		back := common.GetConfString("backend", option, "")
		if back == "" {
			fmt.Printf("Read conf %s failed, error:%s, getBackend total number:%d \n", option, err.Error(), b.backendNum)
			continue
		}

		backend_name := strings.TrimPrefix(option, "backend_list_")
		if backend_name == "" || backend_name == option {
			fmt.Printf("Get backend name failed")
			continue
		}

		addrs := strings.Split(back, ";")
		num := len(addrs)

		if num < 1 {
			fmt.Printf("one backend: %s must have at least one address", backend_name)
			continue
		}

		point := &BackPoint{
			backAddrs:      addrs,
			name:           backend_name,
			conn:           nil,
			connTimeout:    common.GetConfSecond("backend", "connection_timeout", 180),
			currentBackIdx: -1,
			backNum:        num,
			sendTimeout:    common.GetConfSecond("backend", "send_timeout", 180),
			retryTimes:     common.GetConfInt("backend", "retry_times", 5),
			retryInterval:  common.GetConfSecond("backend", "retry_interval", 50),
			recvBuf:        make([]byte, common.GetConfInt("backend", "receive_buffer_size", 10)),
			sendingChan:    make(chan []byte, common.GetConfInt("backend", "sending_buffer_size", 1000)),
		}

		if point == nil {
			fmt.Println("New backPoint failed")
			return nil //不需要退出?
		}

		b.backendList = append(b.backendList, point)
		b.backendNum += 1
		b.logger.Debug("Backend %d get a backend: %s, %d points", idx, backend_name, back, num)
	}

	if b.backendNum < 1 {
		fmt.Println("no backend")
		return nil
	}

	return b
}

func (this *Backend) Start() {
	this.running = true
	this.logger.Info("Backend %d start to work, wait for dispatching...", this.index)

	var (
		bytes []byte
		ok    bool
		point *BackPoint
	)

	for _, point = range this.backendList {
		go this.sending(point)
	}

	for this.running {
		this.heartbeat = time.Now().Unix()

		//读取数据
		bytes, ok = <-this.inChan
		if !ok {
			this.logger.Error("Backend %d read from filter channel failed", this.index)
			break
		}
		this.logger.Debug("Backend %d get data from input channel success length: %d Content:%s", this.index, len(bytes), string(bytes))

		//向后端goroutine 转发
		for _, point = range this.backendList {
			select {
			case point.sendingChan <- bytes:
			default:
				this.logger.Error("Backend %d insert into backend point %s sending channel failed, length:%d too more, data droped")
				//TODO 报警
			}
		}
		this.logger.Debug("Backend %d dispatch out data to all back %d points success, remainning %d apckets in packet channel", this.index, this.backendNum, len(this.inChan))
	}

	this.logger.Info("Backend %d quit working", this.index)
}

func (this *Backend) Stop() {
	this.running = false
	this.logger.Info("Backend%d stop", this.index)
}

func (this *Backend) sending(point *BackPoint) {
	var (
		bytes []byte
		ok    bool
		err   error
	)

	for this.running {
		bytes, ok = <-point.sendingChan
		if !ok {
			this.logger.Error("backend %d backpoint %s read from sending channel failed", this.index, point.name)
			break
		}
		this.logger.Debug("Backend %d backpoint %s read from sending channel success%s", this.index, point.name, string(bytes))

		err = this.send(bytes, point)
		if err != nil {
			this.logger.Error("Backend %d send data out to %s failed, error: %s, data droped, remaining %d packets in sending channel:%s", this.index, point.name, err.Error(), len(point.sendingChan), string(bytes))
			if point.conn != nil {
				point.conn.Close() //为什么不用defer
				point.conn = nil   //为什么不关掉
			}
			time.Sleep(point.retryInterval)
			continue

		}
		this.logger.Debug("Backend %d send data out to %s success, remainning %d packets in sending channel", this.index, point.name, len(point.sendingChan))
	}
	this.logger.Info("Backend %d backpoint %s quit working", this.index, point.name)
}

func (this *Backend) send(bytes []byte, point *BackPoint) error {
	var err error
	var num int
	if point.conn == nil {
		err = point.connect()
		if err != nil {
			this.logger.Error("Backend %d connect to %s failed", this.index, point.name)
			return err
		}
		this.logger.Info("Backend %d connect to %s success, IP:%s", this.index, point.name, point.backAddrs[point.currentBackIdx])
	}
	length := len(bytes)
	sentNum := 0
	start := time.Now()

	//There is no need to detect where the conn is closed by peer, If it is closed, Write() will return "EOF" error
	for sentNum < length {
		num, err = point.conn.Write(bytes[sentNum:])
		if err != nil {
			this.logger.Error("Backend %d write to %s failed, duration %v, error:%s, IP: %s, yet already sent %d bytes", this.index, point.name, time.Now().Sub(start), err.Error(), point.backAddrs[point.currentBackIdx], sentNum)
			return err
		}
		sentNum += num
	}
	this.logger.Debug("Backend %s write to %s success ,duration %v, IP:%s, lenght %d, content", this.index, point.name, time.Now().Sub(start), err.Error(), point.backAddrs[point.currentBackIdx], sentNum)

	point.conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	num, err = point.conn.Read(point.recvBuf)
	if err != nil {
		this.logger.Error("Backend %d read from %s failed, IP %s, error:%s", this.index, point.name, point.backAddrs[point.currentBackIdx], err.Error())
		if err == io.EOF {
			this.logger.Error("Backend %d detect close packet from %s IP: %s", this.index, point.name, point.backAddrs[point.currentBackIdx])
			return err
		}
	}

	return nil
}

func (this *BackPoint) connect() error {
	if this.currentBackIdx < 0 || this.currentBackIdx >= this.backNum {
		this.currentBackIdx = time.Now().Nanosecond() % this.backNum
	}

	var addr string
	var conn net.Conn
	var err error

	for i := 0; i < this.retryTimes; i++ {
		this.currentBackIdx = (this.currentBackIdx + 1) % this.backNum
		addr = this.backAddrs[this.currentBackIdx]
		conn, err = net.DialTimeout("tcp", addr, this.connTimeout)
		if err == nil {
			this.conn = conn
			return nil
		}
		if conn != nil {
			conn.Close()
		}
		time.Sleep(this.retryInterval)
	}

	this.conn = nil

	//Alarm TODO
	return err
}
