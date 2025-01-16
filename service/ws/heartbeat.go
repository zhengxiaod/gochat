package ws

import (
	"fmt"
	"time"
)

// HeartbeatChecker 心跳检测
type HeartbeatChecker struct {
	interval time.Duration // 心跳检测时间间隔
	quit     chan struct{} // 退出信号

	clientManager *ClientManager // 所属服务端
}

func NewHeartbeatChecker(interval time.Duration, cm *ClientManager) *HeartbeatChecker {
	return &HeartbeatChecker{
		interval:      interval,
		quit:          make(chan struct{}, 1),
		clientManager: cm,
	}
}

// Start 启动心跳检测
func (h *HeartbeatChecker) Start() {
	fmt.Println("HeartbeatChecker Start ... ")

	ticker := time.NewTicker(h.interval)
	for {
		select {
		case <-ticker.C:
			h.check()
		case <-h.quit:
			ticker.Stop()
			return
		}
	}
}

// Stop 停止心跳检测
func (h *HeartbeatChecker) Stop() {
	h.quit <- struct{}{}
}

// check 超时检测
func (h *HeartbeatChecker) check() {
	fmt.Println("heart check start...", time.Now().Format("2006-01-02 15:04:05"))
	// 已验证的连接
	conns := h.clientManager.GetConnAll()
	for _, conn := range conns {
		if !conn.IsAlive() {
			conn.Stop()
		}
	}
}
