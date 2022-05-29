package main

import (
	"log"
	"net"
)

func main() {
	err := Init()
	if err != nil {
		return
	}

	listen, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Println(err)
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go process(conn)
	}
}
