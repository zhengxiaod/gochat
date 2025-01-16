package ws

import (
	"fmt"
	"sync"
)

type ClientManager struct {
	connMap  map[uint64]*Conn
	connLock sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		connMap: make(map[uint64]*Conn),
	}
}

// AddConn 添加连接
func (cm *ClientManager) AddConn(userId uint64, conn *Conn) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	cm.connMap[userId] = conn
	fmt.Printf("connection UserId=%d add to Server\n", userId)
}

// GetConn 根据userid获取相应的连接
func (cm *ClientManager) GetConn(userId uint64) *Conn {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	conn, ok := cm.connMap[userId]
	if ok {
		return conn
	}
	return nil
}

// GetConnAll 获取全部连接
func (cm *ClientManager) GetConnAll() []*Conn {
	conns := make([]*Conn, 0)
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	for _, conn := range cm.connMap {
		conns = append(conns, conn)
	}
	return conns
}

// RemoveConn 删除连接
func (cm *ClientManager) RemoveConn(userId uint64) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	delete(cm.connMap, userId)
	fmt.Printf("connection UserId=%d remove from Server\n", userId)
}
