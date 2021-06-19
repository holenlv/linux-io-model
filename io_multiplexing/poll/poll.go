package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"strings"
	"syscall"
)

func main() {
	// tcp socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(fmt.Sprintf("create socket failed,err=%v", err))
	}

	// 断开连接后，会有2msl时间，加这个参数，可以立即复用端口,不然会端口占用问题
	if err = syscall.SetsockoptInt(serverFD, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		panic(fmt.Sprintf("set socket opt failed,err=%v", err))
	}

	// 绑定端口 相当于 0.0.0.0:9988
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: 9988,
	}); err != nil {
		panic(fmt.Sprintf("bind failed,err=%v", err))
	}
	// 监听
	// 第二个参数的作用 @see https://www.cnblogs.com/orgliny/p/5780796.html
	if err = syscall.Listen(serverFD, 5); err != nil {
		panic(fmt.Sprintf("listen failed,err=%v", err))
	}
	var fdSet []unix.PollFd
	fdSet = append(fdSet, unix.PollFd{
		Fd:     int32(serverFD),
		Events: unix.POLLIN,
	})
	for {
		_, err := unix.Poll(fdSet, 0)
		if err != nil {
			fmt.Printf("poll failed,err=%v", err)
			continue
		}
		// 有新的客户端连接了
		// todo
		if fdSet[0].Revents&unix.POLLIN != 0 {
			newClientFD, sa, err := syscall.Accept(serverFD)
			if err != nil {
				fmt.Printf("accept failed,err=%v\n", err)
				continue
			}
			addr := sa.(*syscall.SockaddrInet4)
			fmt.Printf("握手成功 serverFD = %v,client ip=%v,port=%v\n", newClientFD, addr.Addr, addr.Port)
		}
	}
}

func readData(fd int) string {
	var data strings.Builder
	buffer := make([]byte, 2)
	for {
		// 如果客户端不发数据，Read就会一直阻塞在这里,那就会导致无法处理其他客户端的请求
		// C 一般可以用多进程来完成并发，golang使用goroutine
		// go readData(fd)()
		n, err := syscall.Read(fd, buffer)
		if err != nil {
			continue
		}
		if n == 0 {
			break
		}
		data.Write(buffer[:n])
	}
	syscall.Close(fd)
	return data.String()
}
