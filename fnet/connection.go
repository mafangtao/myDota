package fnet

import (
	"errors"
	"net"
	"sync"
	"time"
	"myDota/iface"
	"myDota/logger"
	"myDota/utils"
	"io"
	"fmt"
)

type Connection struct {
	Conn        *net.TCPConn //socket接口
	isClosed    bool         //连接是否已经关闭
	SessionId   uint32
	dataPack    iface.IDataPack
	msgHandler	iface.IMsgHandle
	propertyBag map[string]interface{}
	buffChan    chan []byte //用于传送信息给负责写的goroutine
	exitChan    chan bool   //通知模块内的goroutine退出

	sendLock     sync.RWMutex
	propertyLock sync.RWMutex
}

func NewConnection(conn *net.TCPConn, sessionId uint32, msgHandler iface.IMsgHandle) *Connection {
	fconn := &Connection{
		Conn:        conn,
		isClosed:    false,
		SessionId:   sessionId,
		dataPack:    NewDataPack(),
		msgHandler:	 msgHandler,
		propertyBag: make(map[string]interface{}),
		buffChan:    make(chan []byte, utils.GlobalObject.MaxSendChanLen),
		exitChan:    make(chan bool, 1),
	}

	fconn.SetProperty("ctime", time.Now().UnixNano())

	//向连接管理器注册自己的信息
	utils.GlobalObject.TcpServer.GetConnectionMgr().Add(fconn)
	return fconn
}

func (c *Connection) Start() {
	//设置频率控制
	c.SetFrequencyControl()
	//按照utils中定义好的方式，进行连接处理
	if utils.GlobalObject.OnConnectioned != nil{
		utils.GlobalObject.OnConnectioned(c)
	}
	//开启用于WriterThread
	go c.StartWriteThread()
	//开启读取数据的goroutine
	go c.StartReadThread()
}

func (c *Connection) Stop() {
	if c.isClosed{
		return
	}

	c.sendLock.Lock()
	defer c.sendLock.Unlock()

	c.Conn.Close()
	c.exitChan <- true
	c.isClosed = true

	//防止用户定义的函数阻塞线程
	if utils.GlobalObject.OnClosed != nil{
		go utils.GlobalObject.OnClosed(c)
	}

	//从连接管理器中删除连接记录
	utils.GlobalObject.TcpServer.GetConnectionMgr().Remove(c)

	//关闭内部管道
	close(c.exitChan)
	close(c.buffChan)
}


func (c *Connection) GetSessionId() uint32 {
	return c.SessionId
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed{
		return errors.New("connection closed")
	}

	dpkg, err := c.dataPack.Pack(NewDataPkg(msgId, data))
	if err != nil{
		logger.Error("pack data err:", err)
		return errors.New("pack data err")
	}

	//防止多线程调用send造成的冲突
	c.sendLock.Lock()
	defer c.sendLock.Unlock()
	_, err = c.Conn.Write(dpkg)
	return err
}

func (c *Connection) SendMsgBuff(msgId uint32, data []byte) error {
	if c.isClosed{
		return errors.New("connection closed")
	}

	dpkg, err := c.dataPack.Pack(NewDataPkg(msgId, data))
	if err != nil{
		logger.Error("pack data err:", err)
		return errors.New("pack data err")
	}

	//设置timeout
	select{
	case <-time.After(2*time.Second):
		logger.Error("send error: timeout")
		return errors.New("send error: timeout")
	case c.buffChan <- dpkg:
		return nil
	}
}

func (c *Connection) SetProperty(k string, v interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.propertyBag[k] = v
}

func (c *Connection) GetProperty(k string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.propertyBag[k]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (c *Connection) RemoveProperty(k string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.propertyBag, k)
}

func (c *Connection) StartWriteThread() {
	logger.Info(c.RemoteAddr().String(),"conn writer is running")
	defer logger.Info(c.RemoteAddr().String(),"conn writer exit")
	defer c.Stop()

	for {
		select {
		case data := <-c.buffChan:
			c.sendLock.Lock()
			if _, err := c.Conn.Write(data); err != nil {
				logger.Info("send data error:", err, "conn writer exit")
				c.sendLock.Unlock()
				return
			}
			c.sendLock.Unlock()

		case <-c.exitChan:
			return
		}
	}

}

func(c *Connection) StartReadThread (){
	logger.Info(c.RemoteAddr().String(),"conn reader is running")
	defer logger.Info(c.RemoteAddr().String(),"conn reader exit")
	defer c.Stop()

	for{
		//频率控制
		err := c.DoFrequencyControl()
		if err != nil{
			logger.Error(err)
			return
		}

		//read head data
		headdata := make([]byte, c.dataPack.GetHeadLen())

		if _, err := io.ReadFull(c.GetTCPConnection(), headdata); err != nil{
			logger.Error(err)
			return
		}
		pkgHead, err := c.dataPack.Unpack(headdata)
		if err != nil{
			logger.Error(err)
			return
		}
		//读取数据
		l := pkgHead.GetLen()
		var temp []byte
		if l > 0{
			temp = make([]byte, l)
			if _, err := io.ReadFull(c.GetTCPConnection(), temp); err != nil{
				logger.Error(err)
				return
			}
		}
		pkgHead.SetData(temp)

		logger.Debug(fmt.Sprintf("msg id: %d, data len: %d, data: %v", pkgHead.GetMsgId(), pkgHead.GetLen(), pkgHead.GetData()))
		if utils.GlobalObject.PoolSize > 0{
			c.msgHandler.DeliverToMsgQueue(&Request{
				Fconn:c,
				Pdata:pkgHead,
			})
		}else{
			c.msgHandler.DoMsgFromGoRoutine(&Request{
				Fconn:c,
				Pdata:pkgHead,
			})
		}
	}
}

func (c *Connection) SetFrequencyControl(){
	fc_times, fc_unit := utils.GlobalObject.GetFrequency()
	if fc_unit == "h" {
		c.SetProperty("zinx_fc_count", 0)
		c.SetProperty("zinx_fc_times", fc_times)
		c.SetProperty("zinx_fc_ts", time.Now().UnixNano()/1e6 + int64(3600*1e3))
	}else if fc_unit == "m"{
		c.SetProperty("zinx_fc_count", 0)
		c.SetProperty("zinx_fc_times", fc_times)
		c.SetProperty("zinx_fc_ts", time.Now().UnixNano()/1e6 + int64(60*1e3))
	}else if fc_unit == "s"{
		c.SetProperty("zinx_fc_count", 0)
		c.SetProperty("zinx_fc_times", fc_times)
		c.SetProperty("zinx_fc_ts", time.Now().UnixNano()/1e6 + int64(1e3))
	}
}

func (c *Connection) DoFrequencyControl() error{
	zinx_fc_ts, err := c.GetProperty("zinx_fc_ts")
	if err != nil{
		//无频率控制
		return nil
	}

	if time.Now().UnixNano()/1e6 > zinx_fc_ts.(int64){
		//reset
		c.SetFrequencyControl()
		return nil
	}

	zinx_fc_count_temp, _ := c.GetProperty("zinx_fc_count")
	zinx_fc_count := zinx_fc_count_temp.(int) + 1
	zinx_fc_times_temp, _ := c.GetProperty("zinx_fc_times")
	zinx_fc_times := zinx_fc_times_temp.(int)
	if zinx_fc_count >= zinx_fc_times{
		//超出频率
		return errors.New(fmt.Sprintf("received package exceed limit: %s", utils.GlobalObject.FrequencyControl))
	}

	//计数自增
	c.SetProperty("zinx_fc_count", zinx_fc_count)
	return nil
}