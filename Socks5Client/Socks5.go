package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	serverAddr string
)

func process(client net.Conn) {
	err := authenticate(client)
	if err != nil {
		client.Close()
		return
	}

	server, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Println(err)
		return
	}

	err = connect(client, server)
	if err != nil {
		server.Close()
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

	if buf[0] != 0x05 {
		return errors.New("socks version wrong")
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
		log.Println(err)
		return err
	}

	if buf[0] != 0x05 {
		return errors.New("socks5 version wrong")
	}
	if buf[1] != 0x01 {
		return errors.New("command doesn't support")
	}

	//_, err = server.Write(buf)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}
	err = encryptWrite(server, buf)
	if err != nil {
		log.Println(err)
		return err
	}

	//buf = make([]byte, 10)
	//_, err = server.Read(buf)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}

	read, err := decryptRead(server)
	if err != nil {
		log.Println(err)
		return err
	}

	buf = read[:10]
	fmt.Printf("%d ", len(buf))
	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", buf[i])
	}
	fmt.Println()

	_, err = client.Write(buf)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func forward(src, dest net.Conn) {
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
	defer file.Close()
	if err != nil {
		return err
	}

	var addr, port string
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		text := reader.Text()
		split := strings.Split(text, "=")
		if split[0] == "addr" {
			addr = split[1]
		} else if split[0] == "port" {
			port = split[1]
		} else if split[0] == "key" {
			key = []byte(split[1])
		}
	}
	serverAddr = fmt.Sprintf("%s:%s", addr, port)
	return nil
}
