package main

import (
	"errors"
	"fmt"
	"io"
	"net"
)

var count int

func main() {
	count = 0
	server, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Printf("Listen failed: %v\n", err)
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
		return
	}

	err = connect(client, server)
	if err != nil {
		fmt.Println("connect error:", err)
		client.Close()
		return
	}

	//forwardData(client, server)
	go forward(client, server)
	forward(server, client)
}

func authenticate(client net.Conn) (err error) {
	buf := make([]byte, 256)
	_, err = client.Read(buf)
	if err != nil {
		return err
	}

	//无需认证
	n, err := client.Write([]byte{0x05, 0x00})
	if n != 2 || err != nil {
		return errors.New("write rsp: " + err.Error())
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

	//addr := ""
	//var port uint16
	//switch buf[3] {
	//case 1:
	//	addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
	//	port = binary.BigEndian.Uint16(buf[8:10])
	//case 3:
	//	addrLen := int(buf[4])
	//	addr = string(buf[5 : 5+addrLen])
	//	port = binary.BigEndian.Uint16(buf[5+addrLen : 5+addrLen+2])
	//	fmt.Println(addr)
	//case 4:
	//	fmt.Println("no supported")
	//	return nil, err
	//}
	//
	//addrPort := fmt.Sprintf("%s:%d", addr, port)
	//src, err := net.Dial("tcp", addrPort)
	//if err != nil {
	//	src.Close()
	//	return nil, err
	//}

	//_, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
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

//func forwardData(client, target net.Conn) {
//	forward := func(src, dest net.Conn) {
//		defer src.Close()
//		defer dest.Close()
//		io.Copy(src, dest)
//	}
//	count++
//	//fmt.Printf("begin to forward data...  count:%d\n", count)
//	go forward(client, target)
//	go forward(target, client)
//}

func forward(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(dest, src)
}
