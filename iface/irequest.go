package iface

type IRequest interface{
	GetConnection() IConnection	//获取请求连接信息
	GetData() []byte			//获取请求消息的数据
	GetMsgId() uint32			//获取请求消息的ID
}