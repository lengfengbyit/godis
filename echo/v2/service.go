package main

import (
	"context"
	"godis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// 客户端连接
type Client struct {
	Conn    net.Conn
	Waiting wait.Wait
}

type EchoHandler struct {
	// 保存所有工作状态 client 的集合
	// 需要使用并发安全的容器
	activeConn sync.Map

	// 和 tcp server 中作用相同的关闭状态表示位
	closing int
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (c *Client) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	return c.Conn.Close()
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {

}
