package utils

import (
	"encoding/json"
	"myDota/iface"
	"myDota/logger"
	"io/ioutil"
	"strconv"
	"strings"
	"os"
	"myDota/timer"
)

type GlobalObj struct {
	TcpServer              iface.IServer

	//连接建立时的回调函数
	OnConnectioned         func(fconn iface.IConnection)
	//连接断开时的回调函数
	OnClosed               func(fconn iface.IConnection)
	//服务器停止时的回调函数
	OnServerStop           func() //服务器停服回调
	//主机IP
	Host                   string
	//服务器监听端口
	TcpPort                int
	//服务器最大连接数
	MaxConn                int

	//log
	LogPath          string
	LogName          string
	MaxLogNum        int32
	MaxFileSize      int64
	LogFileUnit      logger.UNIT
	LogLevel         logger.LEVEL
	SetToConsole     bool
	LogFileType      int32

	//处理消息的工作go程数
	PoolSize         int32
	//每个工作go程对应的任务队列容量
	MaxWorkerLen     int32
	//每个连接对应的发送消息的缓存队列容量
	MaxSendChanLen   int32
	//服务器名字
	Name             string
	//最大的底层封包大小
	MaxPacketSize    uint32
	//设置读取连接数据的频率
	FrequencyControl string //  例： 100/h, 100/m, 100/s
	//向主goroutine发送信号的通道
	ProcessSignalChan chan os.Signal
	//时间轮
	timerSchedule *timer.TimerSchedule
}

func (this *GlobalObj) GetFrequency() (int, string) {
	fc := strings.Split(this.FrequencyControl, "/")
	if len(fc) != 2 {
		return 0, ""
	}

	fc0_int, err := strconv.Atoi(fc[0])
	if err != nil {
		logger.Error("FrequencyControl params error: ", this.FrequencyControl)
		return 0, ""
	}

	return fc0_int, fc[1]
}

func (this *GlobalObj)IsThreadSafeMode()bool{
	if this.PoolSize == 1{
		return true
	}else{
		return false
	}
}

func (this *GlobalObj)GetTimerSchedule() *timer.TimerSchedule{
	return this.timerSchedule
}

func (this *GlobalObj)Reload(){
	//读取用户自定义配置
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		Name:                   "single_mode",
		Host:                   "0.0.0.0",
		TcpPort:                8109,
		MaxConn:                12000,
		LogPath:                "./log",
		LogName:                "server.log",
		MaxLogNum:              10,
		MaxFileSize:            100,
		LogFileUnit:            logger.KB,
		LogLevel:               logger.DEBUG,
		SetToConsole:           true,
		LogFileType:            1,
		PoolSize:               10,
		MaxWorkerLen:           1024 * 2,
		MaxSendChanLen:         1024,
		OnConnectioned:         func(fconn iface.IConnection) {},
		OnClosed:               func(fconn iface.IConnection) {},
		ProcessSignalChan:      make(chan os.Signal, 1),

		timerSchedule:			timer.NewTimerSchedule(),
	}
	GlobalObject.Reload()
}
