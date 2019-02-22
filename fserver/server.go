package fserver

import (
	"time"
	"myDota/iface"
	"myDota/fnet"
	"net"
	"myDota/utils"
	"fmt"
	"myDota/logger"
	"os/signal"
	"syscall"
	"os"
)

type Server struct{
	Name string
	IPVersion string
	IP string
	Port int
	MaxConn int
	GenNum *utils.UUIDGenerator

	connectionMgr iface.IConnectionMgr
	msgHandler iface.IMsgHandle
}

func  NewServer() iface.IServer{
	s:=  &Server{
		Name:utils.GlobalObject.Name,
		IPVersion:"tcp4",
		IP: "0.0.0.0",
		Port:utils.GlobalObject.TcpPort,
		MaxConn:utils.GlobalObject.MaxConn,
		GenNum: utils.NewUUIDGenerator(""),
		connectionMgr: fnet.NewConnectonMgr(),
		msgHandler:fnet.NewMsgHandle(),
	}
	utils.GlobalObject.TcpServer = s
	return s
}

func (s *Server) Start(){
	fmt.Printf("server listening at IP: %s, Port: %d is starting\n", s.IP, s.Port)
	utils.GlobalObject.TcpServer = s
	go func(){
		s.msgHandler.StartWorker(int(utils.GlobalObject.PoolSize))
		addr, err := net.ResolveTCPAddr(s.IPVersion,fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil{
			logger.Fatal("resolve tcp addr err:", err)
			return
		}
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil{
			logger.Fatal("listen tcp err:", err)
			return
		}

		logger.Info(fmt.Sprintf("start myDota server %s...", s.Name))
		for{
			conn, err := listener.AcceptTCP()
			if err != nil{
				logger.Error(err)
				continue
			}

			//服务器最大连接数控制
			if s.connectionMgr.Len() >= s.MaxConn{
				conn.Close()
			}else{
				go s.handleConnection(conn)
			}
		}
	}()

}
func (s *Server) Stop(){
	logger.Info("stop myDota server ", s.Name)
	if utils.GlobalObject.OnServerStop != nil{
		utils.GlobalObject.OnServerStop()
	}
}
func (s *Server) Serve(){
	s.Start()
	s.WaitSignal()
}

func (s *Server) GetConnectionMgr() iface.IConnectionMgr{
	return s.connectionMgr
}
//TODO
func (s *Server) GetConnectionQueue() chan interface{}{
	return nil
}

func (s *Server) GetMsgHandler() iface.IMsgHandle{
	return s.msgHandler
}

func (s *Server) AddRouter(name string, router iface.IRouter){
	logger.Info("add router:", name)
	s.msgHandler.AddRouter(name, router)
}
func (s *Server) CallLater(duration time.Duration, f func(args ...interface{}), args ...interface{}){}
func (s *Server) CallWhen(ts string, f func(args ...interface{}), args ...interface{}){}
func (s *Server) CallLoop(duration time.Duration, f func(args ...interface{}), args ...interface{}){}

func (s *Server) handleConnection(conn  *net.TCPConn){
	fmt.Printf("handle conneciton %s\n", conn.RemoteAddr().String())
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)

	fconn := fnet.NewConnection(conn, s.GenNum.GetUint32(), s.msgHandler)
	fconn.Start()
}

func (s *Server) WaitSignal(){
	signal.Notify(utils.GlobalObject.ProcessSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	sig := <- utils.GlobalObject.ProcessSignalChan
	logger.Info(fmt.Sprintf("server exit. signal: [%s]",sig))
	s.Stop()
}