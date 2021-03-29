package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {

	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前userChannel的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线
func (u *User) Online() {

	// 用户上线，将用户加入到onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线消息
	u.server.BroadCast(u, "Online")

}

// 用户下线
func (u *User) Offline() {

	// 用户下线，将用户从onlineMap中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户下线消息
	u.server.BroadCast(u, "Offline")

}

// 给当前User对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// 用户消息处理
func (u *User) DoMessage(msg string) {

	if msg == "who" {
		// 查询当前在线用户
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": " + "Online...\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("Username has been used: " + u.Name + "\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("Username has been updated: " + u.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式：to|张三|消息内容

		// 1.获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("format error\n Tips:\"to|zhang3|hello\"\n")
			return
		}

		// 2.根据用户名得到对方User对象
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("User not exist.\n")
			return
		}

		// 3.获取消息内容，通过对方User对象将消息内容发送
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("empty msg, try again plz.\n")
			return
		}

		remoteUser.SendMsg(u.Name + " To " + remoteUser.Name + " :" + content + "\n")

	} else {
		u.server.BroadCast(u, msg)
	}

}

// 监听当前 User channel, 一旦有消息，直接发送给对端客户端
func (u *User) ListenMessage() {

	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}

}
