package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个 server 的接口
func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听Message广播消息channel的goroutine, 一旦有消息，发送给全部在线User
func (s *Server) ListenMessage() {

	for {
		msg := <-s.Message
		// 将msg发送给全部在线User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}

}

// 广播消息
func (s *Server) BroadCast(user *User, msg string) {

	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg

	s.Message <- sendMsg

}

func (s *Server) Handler(conn net.Conn) {

	// ...当前链接的业务
	user := NewUser(conn, s)
	fmt.Println("Connection established:", user.Addr)

	/* 用户业务封装
	// 用户上线，将用户加入到onlineMap中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()
	// 广播当前用户上线消息
	s.BroadCast(user, "Online")
	*/

	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				fmt.Println("Connection closed:", user.Addr)
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			/* 用户业务封装
			// 将提取的消息广播
			s.BroadCast(user,msg)
			*/
			user.DoMessage(msg)

			// 用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，重置定时器
			// 不做任何操纵，仅为了激活select， 更新下面的定时器
		case <-time.After(time.Minute * 2):
			// 已经超时，强制关闭当前User
			user.SendMsg("Time out!!!\n")
			//销毁占用的资源
			close(user.C)
			// 关闭链接
			conn.Close()
			// 推出当前Handler
			return //runtime.Goexit()
		}
	}
}

// 启动服务器的接口
func (s *Server) Start() {

	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go s.ListenMessage()

	fmt.Println("Server Starting...")

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go s.Handler(conn)

	}

}
