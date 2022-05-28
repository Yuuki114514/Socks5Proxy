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

	//go forward(client, server)
	//forward(server, client)
	go func() {
		err := encryptForward(client, server)
		if err != nil {
			return
		}
	}()
	err = decryptForward(server, client)
	if err != nil {
		return
	}
}

func authenticate(client net.Conn) error {
	buf := make([]byte, 256)
	_, err := io.ReadFull(client, buf[:2])
	if err != nil {
		log.Println(err)
		return err
	}

	//_, err := client.Read(buf)
	//if err != nil {
	//	return err
	//}

	if buf[0] != 0x05 {
		return errors.New("socks version wrong")
	}

	methods := buf[1]
	_, err = io.ReadFull(client, buf[:methods])
	if err != nil {
		log.Println(err)
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
	//n, err := client.Read(buf)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}
	_, err := io.ReadFull(client, buf[:4])

	if buf[0] != 0x05 {
		return errors.New("socks5 version wrong")
	}
	if buf[1] != 0x01 {
		return errors.New("command doesn't support")
	}

	addrType := buf[3]
	switch addrType {
	case 0x01:
		_, err := io.ReadFull(client, buf[4:4+net.IPv4len+2])
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = encryptWrite(server, buf[:4+net.IPv4len+2])
		if err != nil {
			log.Println(err)
			return err
		}
	case 0x03:
		_, err := io.ReadFull(client, buf[4:5])
		if err != nil {
			log.Println(err)
			return err
		}
		length := buf[4]
		_, err = io.ReadFull(client, buf[5:5+length+2])
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = encryptWrite(server, buf[:5+length+2])
		if err != nil {
			log.Println(err)
			return err
		}
	case 0x04:
		_, err := io.ReadFull(client, buf[4:4+net.IPv6len])
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = encryptWrite(server, buf[:4+net.IPv6len+2])
		if err != nil {
			log.Println(err)
			return err
		}
	}

	//_, err = encryptWrite(server, buf[:n])
	read, _, err := decryptRead(server)
	if err != nil {
		log.Println(err)
		return err
	}

	//buf := make([]byte, 10)
	//n, err = io.ReadFull(server, buf)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}

	//fmt.Println("length: ", n) //
	buf = read[:10]

	_, err = client.Write(buf)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func forward(server, client net.Conn) {
	defer server.Close()
	defer client.Close()
	_, err := io.Copy(client, server)
	if err != nil {
		//log.Println(err)
		return
	}
}

// 从dst读取数据加密传给src
func encryptForward(dst, src net.Conn) error {
	//fmt.Println("encryptForward")
	for {
		buf := make([]byte, 2048)
		read, err := dst.Read(buf)
		if err != nil {
			log.Println(err)
			return err
		}

		fmt.Println("1: 加密前/读取: ", read)
		if read > 0 {
			//write, err := encryptWrite(src, buf[0:read])
			//encrypt, err := AesEncrypt(buf[:read], key)
			write, err := src.Write(buf[:read])
			//write, err := src.Write(encrypt)
			fmt.Println("1: 加密后/发送: ", write)
			if err != nil {
				log.Println(err)
				return err
			} else if read != write {
				return io.ErrShortWrite
			}
		}
	}
}

// 从dst读取数据解密传给src
func decryptForward(dst, src net.Conn) error {
	//fmt.Println("decryptForward")
	for {
		//lenBuf := make([]byte, 2)
		//n, err := dst.Read(lenBuf)
		//if err != nil {
		//	log.Println(err)
		//	return err
		//}
		//if n != 2 {
		//	log.Println("length get wrong")
		//	return nil
		//}
		//length := binary.BigEndian.Uint16(lenBuf)
		//fmt.Println("decrypt length: ", length)

		//buf, read, err := decryptRead(dst)
		buf := make([]byte, 3072)
		read, err := dst.Read(buf)
		//decrypt, err := AesDecrypt(buf[:read], key)
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Println("4: 解密前: ", read)

		if read > 0 {
			write, err := src.Write(buf[0:read])
			//write, err := src.Write(buf2)
			//write, err := src.Write(decrypt)
			fmt.Println("4: 解密后: ", write)
			if err != nil {
				log.Println(err)
				return err
			} else if read != write {
				return io.ErrShortWrite
			}
		}
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
