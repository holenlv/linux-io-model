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
	//set := &syscall.FdSet{}
	fdSet := &unix.FdSet{}
	clientFDSet := map[int]bool{}
	maxSocket := serverFD
	for {
		fdSet.Zero()
		fdSet.Set(serverFD)
		//unix.Select()
		for clientFD := range clientFDSet {
			fdSet.Set(clientFD)
		}
		// 会阻塞到事件的发生
		// fdSet 每次需要从用户态拷贝到内核态内存 比较耗费性能
		ret, err := unix.Select(maxSocket+1, fdSet, nil, nil, nil)
		if err != nil {
			fmt.Printf("select falied,err=%v\n", err)
			continue
		}
		if ret == 0 {
			fmt.Println("ret is 0")
		}
		// select 不会返回哪些fd准备好，需要自己去遍历，比较耗性能
		for clientFD := range clientFDSet {
			// 数据是否准备好了
			if fdSet.IsSet(clientFD) {
				fmt.Printf("fd = %v data is readdy\n", clientFD)
				readData(clientFD)
				delete(clientFDSet, clientFD)
			}
		}
		// 是否有新连接
		if fdSet.IsSet(serverFD) {
			newClientFD, sa, err := syscall.Accept(serverFD)
			if err != nil {
				fmt.Printf("accept failed,err=%v\n", err)
				continue
			}
			addr := sa.(*syscall.SockaddrInet4)
			fmt.Printf("握手成功 fd = %v,client ip=%v,port=%v\n", newClientFD, addr.Addr, addr.Port)

			clientFDSet[newClientFD] = true
			if newClientFD > maxSocket {
				maxSocket = newClientFD
			}
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
