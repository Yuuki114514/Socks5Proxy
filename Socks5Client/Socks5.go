package main

import (
	"fmt"
	"io"
	"net"
)

func process(client net.Conn) {
	err := authenticate(client)
	if err != nil {
		fmt.Println("auth error:", err)
		client.Close()
		return
	}

	server, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("connect to server error")
		server.Close()
		return
	}

	err = connect(client, server)
	if err != nil {
		fmt.Println("connect error:", err)
		client.Close()
		return
	}

	go forward(client, server)
	forward(server, client)
}

func authenticate(client net.Conn) error {
	buf := make([]byte, 256)
	_, err := client.Read(buf)
	if err != nil {
		return err
	}

	//无需认证
	_, err = client.Write([]byte{0x05, 0x00})
	if err != nil {
		return err
	}

	return nil
}

func connect(client net.Conn, server net.Conn) error {
	buf := make([]byte, 256)
	_, err := client.Read(buf)
	if err != nil {
		return err
	}

	_, err = server.Write(buf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	buf = make([]byte, 10)
	_, err = server.Read(buf)
	if err != nil {
		fmt.Println(err)
	}
	_, err = client.Write(buf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func forward(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(dest, src)
}
