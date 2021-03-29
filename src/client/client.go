package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前 client 模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	// 返回对象
	return client
}

// 处理server回应的消息，直接显示到标准输出
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {

	var flag int

	fmt.Println("1.public chat...")
	fmt.Println("2.private chat...")
	fmt.Println("3.update username...")
	fmt.Println("0.exit...")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>> 请输入相应数字 <<<")
		return false
	}

}

// 公共聊天
func (client *Client) PublicChat() {

	// 提示用户输入消息
	var chatMsg string
	fmt.Println(">>> 请输入发送内容，输入\"exit\"退出. <<<")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送服务器
		// 判断消息不为空
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err: ", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>> 请输入发送内容，输入\"exit\"退出. <<<")
		fmt.Scanln(&chatMsg)
	}

}

// 查询在线用户
func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err: ", err)
		return
	}
}

// 私聊
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>> 请输入对方用户名，输入\"exit\"退出. <<<")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>> 请输入消息内容，输入\"exit\"退出. <<<")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			// 判断消息不为空
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err: ", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>> 请输入消息内容，输入\"exit\"退出. <<<")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println(">>> 请输入对方用户名，输入\"exit\"退出. <<<")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>> 请输入用户名 <<<")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Whrte err: ", err)
		return false
	}

	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// public
			client.PublicChat()
			break
		case 2:
			// private
			client.PrivateChat()
			break
		case 3:
			// update
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认为8888)")
}
