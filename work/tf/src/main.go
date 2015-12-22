package main

import (
	"flag"
	"fmt"
	"network"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	//"runtime"

	"common"
	"lib/go-config/config"
)

const (
	VERSION string = "tf_v1.0"
)

func initialize() error {
	fmt.Println("tf initializing ....")

	//读取配置文件，缺省是./etc/tf.conf
	//flag是commandLine解析package
	confFile := flag.String("c", "./etc/tf.conf", "config file name")
	if confFile == nil {
		return fmt.Errorf("no config file")
	}
	//初始化相关参数
	flag.Parse()

	conf, err := config.ReadDefault(*confFile)
	common.CheckError(err, "load config file failed: ")
	common.Conf = conf

	max_processor := common.GetConfInt("global", "max_processor", runtime.NumCPU())
	fmt.Println("max_processor: ", max_processor)
	runtime.GOMAXPROCS(max_processor)
	fmt.Println("exe max_processor")

	//work dir 有什么用
	dir, err := common.Conf.String("global", "root_dir")
	if err == nil {
		err = os.Chdir(dir)
		if err != nil {
			return fmt.Errorf("change working directory to %s failed:%s, dir, err.Error()")
		}
	}
	common.Dir, _ = os.Getwd()
	fmt.Println("work directory:" + common.Dir)

	//
	num := common.GetConfInt("global", "work_num", 1)
	if num < 1 {
		return fmt.Errorf("work number must bigger than 1")
	}

	//生成channel
	common.WorkerNum = num
	common.PacketChans = make([]chan []byte, num)
	for i := 0; i < num; i++ {
		common.PacketChans[i] = make(chan []byte, common.GetConfInt("server", "packet_chan_size", 10000))
		if common.PacketChans[i] == nil {
			return fmt.Errorf("make packet channel failed")
		}
	}

	fmt.Println("initialize over")
	fmt.Printf("Program %s start success in %s at :%s, Max processor:%d Worker number:%d \n", VERSION, common.Dir, time.Now(), max_processor, num)
	return nil
}

func main() {

	if err := initialize(); err != nil {
		fmt.Sprintf("init failed %s \n", err.Error())
	}

	server := network.NewServer()
	if server == nil {
		panic("New tcp server failed")
	}
	fmt.Println("Server%s", server)
	go server.Start()

	backend := network.NewBackend(0, common.PacketChans[0])
	if backend == nil {
		panic("New backend failed")
	}

	sig_chan := make(chan os.Signal)
	signal.Notify(sig_chan, os.Interrupt, syscall.SIGTERM)
	<-sig_chan
}
