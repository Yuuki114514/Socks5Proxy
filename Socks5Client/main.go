package main

import (
	"log"
	"net"
)

func main() {
	err := getConfig()
	if err != nil {
		log.Println(err)
		return
	}

	log.SetPrefix("[ERROR]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	server, err := net.Listen("tcp", "127.0.0.1:9999")
	defer server.Close()
	if err != nil {
		log.Println("Listen() error: ", err)
		server.Close()
		return
	}

	for {
		client, err := server.Accept()
		if err != nil {
			log.Println("Accept() error: ", err)
			continue
		}
		//fmt.Println("accepted")
		go process(client)
	}
}
