package main

import (
	"fmt"
	"net"
)

func main() {
	server, err := net.Listen("tcp", ":9999")
	defer server.Close()
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
		server.Close()
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Printf("Accept failed: %v", err)
			continue
		}
		//fmt.Println("accepted")
		go process(client)
	}
}
