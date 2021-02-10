package main

import (
	"context"
	"fmt"
	atomic2 "godis/lib/sync/atomic"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

type Config struct {
	Address string
}

func ListenAndServe(cfg *Config, handler Handler) {
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatal("listen err:", err)
	}

	var closing atomic2.AtomicBool
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("shuting down...")

			// 表示正在关闭连接
			closing.Set(true)

			// 先关闭 listener 阻止新连接进入
			_ = listener.Close()
			// 逐个关闭已建立的连接
			_ = handler.Close()
		}
	}()

	log.Println(fmt.Sprintf("bind %s, start listening...", cfg.Address))
	defer func() {
		if closing.Get() {
			return
		}

		// 在出现位置错误或 panic 后保证正常关闭
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx, _ := context.WithCancel(context.Background())
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if closing.Get() {
				// 已经收到关闭的信号
				log.Println("waiting disconnect")
				// 主协程等待应用服务器关闭连接
				waitDone.Wait()
				return
			}
			log.Println(fmt.Sprintf("accept err: %v", err))
			continue
		}

		// 创建一个新协程处理链接
		waitDone.Add(1)
		go func() {
			defer waitDone.Done()
			handler.Handle(ctx, conn)
		}()
	}
}
