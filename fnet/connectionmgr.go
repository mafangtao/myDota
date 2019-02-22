package fnet

import (
	"myDota/iface"
	"sync"
	"myDota/logger"
	"fmt"
	"errors"
)

type ConnectionMgr struct {
	connections map[uint32]iface.IConnection
	connMgrLock	sync.RWMutex
}

func NewConnectonMgr() *ConnectionMgr{
	return &ConnectionMgr{
		connections:make(map[uint32] iface.IConnection),
	}
}

func (cm *ConnectionMgr) Add(conn iface.IConnection){
	cm.connMgrLock.Lock()
	defer cm.connMgrLock.Unlock()

	cm.connections[conn.GetSessionId()] = conn
	logger.Debug(fmt.Sprintf("connection add to ConnectionMgr successfully: total connections:%d", len(cm.connections)))
}

func (cm *ConnectionMgr) Remove(connection iface.IConnection)	{
	cm.connMgrLock.Lock()
	defer cm.connMgrLock.Unlock()

	delete(cm.connections, connection.GetSessionId())
}

func (cm *ConnectionMgr) Get(sessionId uint32) (iface.IConnection, error)	{
	cm.connMgrLock.RLock()
	defer cm.connMgrLock.RUnlock()

	if c, ok := cm.connections[sessionId]; ok{
		return c, nil
	}else{
		return nil, errors.New("connection not found")
	}
}

func (cm *ConnectionMgr) Len() int{
	return len(cm.connections)
}