package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)


func ListenAndServe(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("listen err:", err)
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bing: %s, start listening...", address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("accept err:", err)
		}

		// 开启新的 goroutine 处理连接
		go Handle(conn)
	}
}

func Handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
			} else {
				log.Println(err)
			}
			conn.Close()
			return
		}

		b := []byte(msg)
		conn.Write(b)
	}
}

func Handle2(conn net.Conn) {
	for {
		// 将用户发送的内容，回发给用户
		n, err := io.Copy(conn, conn)
		if n == 0 {
			conn.Close()
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
	}
}



func main() {

	ListenAndServe(":8080")
}
