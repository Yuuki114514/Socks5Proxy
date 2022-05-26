package main

import (
	"fmt"
	"net"
)

func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:8888")
	defer listen.Close()
	if err != nil {
		fmt.Println(err)
		listen.Close()
		return
	}

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
