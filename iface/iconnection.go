package iface

import "net"

type IConnection interface {
	Start()				//启动连接
	Stop()				//断开连接

	GetTCPConnection() *net.TCPConn	//获取TCP连接
	GetSessionId() uint32			//获取session id
	RemoteAddr() net.Addr			//获取远程机器地址

	SendMsg(msgId uint32, data []byte) error		//直接发送数据至TCP连接对方
	SendMsgBuff(msgId uint32, data []byte) error	//把数据发送至缓冲区

	SetProperty(k string, v interface{})		//设置连接属性
	GetProperty(k string)(interface{}, error)	//获取连接属性
	RemoveProperty(k string)					//移除连接属性
}
