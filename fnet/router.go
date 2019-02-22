package fnet

import "myDota/iface"

//实现router时，先嵌入这个基类，然后根据需要对这个基类的方法进行重写
type BaseRouter struct {}

func (br *BaseRouter)PreHandle(req iface.IRequest){}
func (br *BaseRouter)Handle(req iface.IRequest){}
func (br *BaseRouter)PostHandle(req iface.IRequest){}
