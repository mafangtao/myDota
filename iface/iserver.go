package iface

import "time"

type IServer interface{
	Start()
	Stop()
	Serve()

	GetConnectionMgr() IConnectionMgr
	GetMsgHandler()	IMsgHandle
	//TODO
	GetConnectionQueue() chan interface{}

	AddRouter(name string, router IRouter)
	CallLater(duration time.Duration, f func(args ...interface{}), args ...interface{})
	CallWhen(ts string, f func(args ...interface{}), args ...interface{})
	CallLoop(duration time.Duration, f func(args ...interface{}), args ...interface{})
}