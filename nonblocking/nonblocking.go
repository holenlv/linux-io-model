// 同步非阻塞
package main

import (
	"fmt"
	"strings"
	"syscall"
)

func main() {
	// tcp socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(fmt.Sprintf("create socket failed,err=%v", err))
	}

	// 断开连接后，会有2msl时间，加这个参数，可以立即复用端口,不然会端口占用问题
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		panic(fmt.Sprintf("set socket opt failed,err=%v", err))
	}

	// 绑定端口 相当于 0.0.0.0:9988
	if err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: 9988,
	}); err != nil {
		panic(fmt.Sprintf("bind failed,err=%v", err))
	}
	// 监听
	// 第二个参数的作用 @see https://www.cnblogs.com/orgliny/p/5780796.html
	if err = syscall.Listen(fd, 5); err != nil {
		panic(fmt.Sprintf("listen failed,err=%v", err))
	}
	for {
		// 握手以后完成Accept返回
		nfd, sa, err := syscall.Accept(fd)
		if err != nil {
			fmt.Printf("accept failed,err=%v\n", err)
			continue
		}
		addr := sa.(*syscall.SockaddrInet4)
		fmt.Printf("握手成功 fd = %v,client ip=%v,port=%v\n", nfd, addr.Addr, addr.Port)
		fmt.Printf("data = %v\n", readData(nfd))
	}
}

func readData(fd int) string {
	// 设置fd为非阻塞IO，会导致Read函数不会阻塞，如果没有数据可读，会返回 EAGAIN 错误
	if err := syscall.SetNonblock(fd, true); err != nil {
		panic(err)
	}
	var data strings.Builder
	buffer := make([]byte, 2)
	for {
		n, err := syscall.Read(fd, buffer)
		if err != nil {
			if err == syscall.EAGAIN {
				// for debug print
				//fmt.Println("EAGAIN error")
				continue
			}
			fmt.Printf("err = %v\n", err)
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
