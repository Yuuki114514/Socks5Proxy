package main

import (
	"fmt"
	"net"
)

func main() {
	err := Init()
	if err != nil {
		return
	}

	listen, err := net.Listen("tcp", "127.0.0.1:8888")

	if err != nil {
		fmt.Println(err)
		listen.Close()
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			conn.Close()
			fmt.Println(err)
			continue
		}
		go process(conn)
	}
}
