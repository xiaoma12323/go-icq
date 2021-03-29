package main

func main() {

	server := NewServer("192.168.73.100", 8888)
	server.Start()

}
