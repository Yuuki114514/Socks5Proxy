package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func process(client net.Conn) {
	//var buf = make([]byte, 256)
	//_, err := client.Read(buf)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	buf, err := decryptRead(client)

	var addr string
	var port uint16
	var ip []byte
	switch buf[3] {
	case 1:
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
		ip = buf[4 : 4+net.IPv4len]
		port = binary.BigEndian.Uint16(buf[4+net.IPv4len : 4+net.IPv4len+2])
		//port = binary.BigEndian.Uint16(buf[n-2:])
	case 3:
		addrLen := int(buf[4])
		addr = string(buf[5 : 5+addrLen])
		fmt.Println(addr)
		ipAddr, err := net.ResolveIPAddr("ip", addr)
		if err != nil {
			log.Println(err)
			return
		}
		ip = ipAddr.IP
		port = binary.BigEndian.Uint16(buf[5+addrLen : 5+addrLen+2])
	case 4:
		ip = buf[4 : 4+net.IPv6len]
		port = binary.BigEndian.Uint16(buf[4+net.IPv6len : 4+net.IPv6len+2])
	}

	dstAddr := net.TCPAddr{IP: ip, Port: int(port)}
	src, err := net.DialTCP("tcp", nil, &dstAddr)
	if err != nil {
		log.Println(err)
		//src.Close()
		return
	}

	//_, err = client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	err = encryptWrite(client, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		log.Println(err)
		return
	}

	go forward(client, src)
	forward(src, client)
}

func forward(dest, src net.Conn) {
	defer src.Close()
	defer dest.Close()
	_, err := io.Copy(dest, src)
	if err != nil {
		//log.Println(err)
		return
	}
}

func Init() error {
	log.SetPrefix("[ERROR]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	file, err := os.Open("Socks5.conf")
	if err != nil {
		log.Println(err)
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")
		if split[0] == "key" {
			key = []byte(split[1])
		}
	}
	return nil
}
