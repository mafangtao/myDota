package fserver

import (
	"testing"
	"myDota/iface"
	"fmt"
	"myDota/utils"
	"myDota/fnet"
	"net"
	"time"
	"io"
)

func TestServer(t *testing.T){
	s := NewServer()

	s.AddRouter("0", &api0{})

	utils.GlobalObject.OnConnectioned = OnConnectionMade
	utils.GlobalObject.OnClosed = OnConnectionLost

	//开启客户端goroutine
	go func(){
		time.Sleep(3 * time.Second)
		conn, err := net.Dial("tcp", "127.0.0.1:8109")
		if err != nil{
			fmt.Println("client start err, exit")
			return
		}

		datapack := fnet.NewDataPack()
		data0 := fnet.NewDataPkg(0, []byte("hello"))
		data1 := fnet.NewDataPkg(0, []byte("world"))

		dp0, err := datapack.Pack(data0)
		dp1, err := datapack.Pack(data1)

		req := append(dp0, dp1...)
		conn.Write(req)
		fmt.Println("client write to server:", req)


		for{
			temp := make([]byte, datapack.GetHeadLen())
			_, err := io.ReadFull(conn, temp)
			if err != nil{
				fmt.Println("client datapack read full err")
				return
			}

			dpkg, err := datapack.Unpack(temp)
			if err != nil{
				fmt.Println("client data unpack err:", err)
				return
			}

			temp = make([]byte, dpkg.GetLen())
			_, err = io.ReadFull(conn, temp)
			if err != nil{
				fmt.Println("client datapack read full err")
				return
			}

			dpkg.SetData(temp)
			fmt.Println("client received data:", string(dpkg.GetData()))
		}
	}()



	s.Serve()
}

func OnConnectionMade(fconn iface.IConnection){
	fmt.Println("do on connection made")
}

func OnConnectionLost(fconn iface.IConnection){
	fmt.Println("do on connection lost")
}

type api0 struct{
	fnet.BaseRouter
}

func (r *api0)Handle(req iface.IRequest){
	fmt.Println(string(req.GetData()))
	req.GetConnection().SendMsg(0,[]byte("got it!!! by SendMsg"))
	req.GetConnection().SendMsgBuff(0,[]byte("got it!!! by SendMsgBuff"))
}
