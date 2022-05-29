package main

import (
	"log"
	"net"
)

func main() {
	err := Init()
	if err != nil {
		log.Println(err)
		return
	}

	server, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Println("Listen() error: ", err)
		return
	}
	defer server.Close()

	for {
		client, err := server.Accept()
		if err != nil {
			log.Println("Accept() error: ", err)
			continue
		}
		go process(client)
	}
}
