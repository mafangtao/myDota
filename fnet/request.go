package fnet

import "myDota/iface"

type Request struct {
	Pdata iface.IMsg
	Fconn iface.IConnection
}
func (r *Request) GetConnection() iface.IConnection{
	return r.Fconn
}
func (r *Request) GetData() []byte {
	return r.Pdata.GetData()
}
func (r *Request) GetMsgId() uint32{
	return r.Pdata.GetMsgId()
}