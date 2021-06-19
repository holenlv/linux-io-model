package main

import (
	"fmt"
	"net"
)

func main() {
	// 直接阅读源代码
	// golang的网络IO模型底层是基于epoll(linux)
	ln, err := net.Listen("tcp", ":9988")
	if err != nil {
		fmt.Printf("listen failed,err=%v", err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(conn net.Conn) {
		}(conn)
	}
}
