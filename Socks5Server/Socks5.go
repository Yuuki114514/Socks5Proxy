package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func process(client net.Conn) {
	b := make([]byte, BlockSize)
	buf, _, err := decryptRead(client, b)
	if err != nil {
		client.Close()
		return
	}

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
	dst, err := net.DialTCP("tcp", nil, &dstAddr)
	if err != nil {
		log.Println(err)
		return
	}
	//defer dst.Close()

	_, err = encryptWrite(client, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		log.Println(err)
		client.Close()
		dst.Close()
		return
	}

	go func() {
		err := decryptForward(client, dst)
		if err != nil {
			client.Close()
			dst.Close()
			return
		}
	}()
	go func() {
		err := encryptForward(dst, client)
		if err != nil {
			dst.Close()
			client.Close()
			return
		}
	}()
}

func Init() error {
	log.SetPrefix("[ERROR]")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	file, err := os.Open("Socks5.conf")
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")
		if split[0] == "key" {
			key = []byte(split[1])
		} else if split[0] == "iv" {
			iv = []byte(split[1])
		}
	}
	return nil
}
