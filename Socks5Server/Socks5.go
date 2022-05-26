package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func process(client net.Conn) {
	var buf = make([]byte, 256)
	_, err := client.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	var addr string
	var port uint16
	switch buf[3] {
	case 1:
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
		port = binary.BigEndian.Uint16(buf[8:10])
	case 3:
		addrLen := int(buf[4])
		addr = string(buf[5 : 5+addrLen])
		port = binary.BigEndian.Uint16(buf[5+addrLen : 5+addrLen+2])
		fmt.Println(addr)
	case 4:
		fmt.Println("no supported")
		return
	}

	addrPort := fmt.Sprintf("%s:%d", addr, port)
	src, err := net.Dial("tcp", addrPort)
	if err != nil {
		src.Close()
		return
	}

	_, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		fmt.Println(err)
		return
	}

	go forward(client, src)
	forward(src, client)
}

func forward(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(dest, src)
}
