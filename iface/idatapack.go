package iface

//打包数据和解包数据（直接面向TCP连接中的数据流，
// 为传输数据添加头部信息，用于处理TCP粘包问题
// 或进行数据压缩，解压缩等）
type IDataPack interface{
	GetHeadLen() uint32
	Pack(msg IMsg)([]byte, error)
	Unpack([]byte)(IMsg, error)
}

//封装消息
type IMsg interface {
	GetLen() uint32		//获取消息数据段长度
	GetMsgId() uint32	//获取消息id
	GetData() []byte	//获取消息内容

	SetLen(uint32)
	SetMsgId(uint32)
	SetData([]byte)
}