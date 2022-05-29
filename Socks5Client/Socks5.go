package main

import (
	"bufio"
	"errors"
	"fmt"
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

	go func() {
		err := encryptForward(client, server)
		if err != nil {
			client.Close()
			server.Close()
			return
		}
	}()
	go func() {
		err := decryptForward(server, client)
		if err != nil {
			server.Close()
			client.Close()
			return
		}
	}()
}

func authenticate(client net.Conn) error {
	buf := make([]byte, 256)
	_, err := client.Read(buf)
	if err != nil {
		log.Println(err)
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
	buf := make([]byte, BlockSize)
	read, err := client.Read(buf)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = encryptWrite(server, buf[:read])
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

	b := make([]byte, BlockSize)
	bytes, _, err := decryptRead(server, b)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = client.Write(bytes)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, "=")
		if split[0] == "addr" {
			addr = split[1]
		} else if split[0] == "port" {
			port = split[1]
		} else if split[0] == "key" {
			key = []byte(split[1])
		} else if split[0] == "iv" {
			iv = []byte(split[1])
		}
	}
	serverAddr = fmt.Sprintf("%s:%s", addr, port)
	return nil
}
