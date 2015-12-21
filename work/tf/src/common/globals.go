package common

import (
	"fmt"
	"lib/go-config/config"
	log "lib/log4go"
	"os"
	"time"
)

var Conf *config.Config
var Dir string
var WorkerNum int

var PacketChans []chan []byte

func CheckError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s :%s", msg, err.Error())
		os.Exit(1)
	}
}

func GetConfInt(section string, key string, def int) int {
	if Conf == nil {
		return def
	}

	val, err := Conf.Int(section, key)
	if err != nil {
		return def
	}
	return val
}

func GetConfString(section string, key string, def string) string {
	if Conf == nil {
		return def
	}

	val, err := Conf.String(section, key)
	if err != nil {
		return def
	}
	return val
}

func GetConfSecond(section string, key string, def time.Duration) time.Duration {
	def *= time.Second
	if Conf == nil {
		return def
	}

	val, err := Conf.Int(section, key)
	if err != nil {
		return def
	}
	return time.Duration(val) * time.Second
}

func NewLogger(name string) log.Logger {
	logFileName := Dir + "/logs/" + name + ".log"
	flw := log.NewFileLogWriter(logFileName, false)
	flw.SetRotateDaily(true)

	level := log.INFO
	if Conf != nil {
		val, err := Conf.String("global", "log_level")
		if err != nil {
			switch val {
			case "info":
				level = log.INFO
			case "debug":
				level = log.DEBUG
			case "error":
				level = log.ERROR
			}
		}
	}

	l := make(log.Logger)
	l.AddFilter("log", level, flw)

	return l
}
