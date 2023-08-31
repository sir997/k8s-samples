package main

import (
	"os"
	"syscall"
)

type netsocket struct {
	fd int
}

func (n *netsocket) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return syscall.Read(n.fd, p)
}

func (n *netsocket) Write(p []byte) (int, error) {
	return syscall.Write(n.fd, p)
}

func (n *netsocket) Accept() (*netsocket, error) {
	nfd, _, err := syscall.Accept(n.fd)
	if err != nil {
		return nil, err
	}
	return &netsocket{nfd}, nil
}

func (n *netsocket) Close() error {
	return syscall.Close(n.fd)
}

func New() (*netsocket, error) {
	syscall.ForkLock.Lock()
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	syscall.ForkLock.Unlock()

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		_ = syscall.Close(fd)
		return nil, os.NewSyscallError("setsocket", err)
	}

	sa := &syscall.SockaddrInet4{Addr: [4]byte{0, 0, 0, 0}, Port: 80}
	if err = syscall.Bind(fd, sa); err != nil {
		return nil, os.NewSyscallError("bind", err)
	}
	// 开始监听客户端的连接请求
	if err = syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return nil, os.NewSyscallError("listen", err)
	}
	return &netsocket{fd: fd}, nil
}
