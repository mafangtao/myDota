package fnet

import (
	"testing"
	"net"
	"fmt"
	"io"
)

func TestDataPack(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:8999")
	if err != nil{
		fmt.Println("server listen err:", err)
		return
	}

	//创建服务器gotoutine，负责从客户端goroutine读取粘包的数据，然后进行解析
	go func (){
		for{
			conn, err := listener.Accept()
			if err != nil{
				fmt.Println("server accept err:", err)
			}

			go func(conn net.Conn){
				dataPack := NewDataPack()
				for{
					headdata := make([]byte, dataPack.GetHeadLen())
					io.ReadFull(conn, headdata)
					head,err := dataPack.Unpack(headdata)
					if err != nil{
						fmt.Println("server unpack err:", err)
					}
					
					h := head.(*DataPkg)
					h.Data = make([]byte, h.GetLen())
					io.ReadFull(conn, h.Data)
					
					fmt.Println( h.Length, "+++++++", h.MsgId, "++++++++++", string(h.Data))
				}
			}(conn)

		}
	}()

	//客户端goroutine，负责模拟粘包的数据，然后进行发送
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil{
		fmt.Println("client dial err:", err)
		return
	}

	data := &DataPkg{
		Length:5,
		MsgId:0,
		Data:[]byte{'h', 'e', 'l', 'l', 'o'},
	}
	
	datapack := NewDataPack()
	packed, err := datapack.Pack(data)
	if err!= nil{
		fmt.Println("client pack err:", err)
		return 
	}

	data2 := &DataPkg{
		Length:5,
		MsgId:1,
		Data:[]byte{'w', 'o', 'r', 'l', 'd'},
	}
	temp, err := datapack.Pack(data2)
	if err!= nil{
		fmt.Println("client temp pack err:", err)
		return
	}
	
	packed = append(packed, temp...)
	conn.Write(packed)
	
	select{}
}