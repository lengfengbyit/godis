package main

import (
	"bufio"
	"context"
	"godis/lib/sync/atomic"
	"godis/lib/sync/wait"
	"io"
	"log"
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
	closing atomic.AtomicBool
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (c *Client) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	return c.Conn.Close()
}

func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		conn.Close()
	}

	client := &Client{
		Conn: conn,
	}
	h.activeConn.Store(client, 1)

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
				h.activeConn.Delete(conn)
			} else {
				log.Println(err)
			}
			return
		}

		// 发送数据前先只为waiting状态
		client.Waiting.Add(1)

		// 模拟关闭时未完成发送的情况
		log.Println("sleeping")
		time.Sleep(10 * time.Second)

		conn.Write([]byte(msg))
		client.Waiting.Done()
	}
}

func (h *EchoHandler) Close() error {
	log.Println("handler shuting down...")
	h.closing.Set(true)

	h.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*Client)
		// 这里用协程会好一点，防止一个链接阻塞所有链接
		go client.Close()
		return true
	})
	return nil
}

func main() {
	cfg := &Config{Address: ":8080"}
	handler := MakeEchoHandler()
	ListenAndServe(cfg, handler)
}