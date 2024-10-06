package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

var (
	serverAddr = []string{"192.168.56.116:6443", "192.168.56.117:6443", "192.168.56.118:6443"} // 目标服务器地址
	proxyAddr  = ":6443"                                                                       // 本地代理监听地址
)

func main() {
	go ListenRemoteClient()
	shutdown := make(chan struct{})
	<-shutdown
}

func ListenRemoteClient() {
	ln, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		fmt.Println("本地监听端口时异常", err)
	}
	for {
		clientConn, err := ln.Accept()
		if err != nil {
			fmt.Println("建立与客户端连接时异常", err)
			return
		}
		go connectPipe(clientConn)
	}
}

func connectPipe(clientConn net.Conn) {

	connTimeout := 3 * time.Second
	serverConn, err := net.DialTimeout("tcp", serverAddr[0], connTimeout)
	// serverConn, err := net.Dial("tcp", serverAddr[0])
	if err != nil {
		fmt.Println("建立与服务端连接时异常", err)
		_ = clientConn.Close()
		return
	}
	pipe(clientConn, serverConn)
}

func pipe(src net.Conn, dest net.Conn) {
	errChan := make(chan error, 1)
	onClose := func(err error) {
		_ = dest.Close()
		_ = src.Close()
	}
	go func() {
		_, err := io.Copy(src, dest)
		errChan <- err
		onClose(err)
	}()
	go func() {
		_, err := io.Copy(dest, src)
		errChan <- err
		onClose(err)
	}()
	<-errChan
}
