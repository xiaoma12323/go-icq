package main

import (
	"flag"
	"fmt"
)

func main() {

	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("Connection error.")
		return
	}

	// 单独开启一个goroutine处理server的回执消息
	go client.DealResponse()

	fmt.Println("Connection success.")

	// 启动客户端业务
	client.Run()
}
