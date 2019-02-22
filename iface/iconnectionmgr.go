package iface

type IConnectionMgr interface {
	Add(connection IConnection)					//添加连接
	Remove(connection IConnection)				//移除连接
	Get(sessionId uint32) (IConnection, error)	//利用sessionId获取连接
	Len() int									//获取所有网络连接的个数
}
