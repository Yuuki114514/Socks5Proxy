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
	buf, _, err := decryptRead(client)

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
		//src.Close()
		return
	}

	_, err = encryptWrite(client, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		log.Println(err)
		return
	}

	//go forward(client, src)
	//forward(src, client)
	//go func() {
	//	err := decryptForward(client, src)
	//	if err != nil {
	//		return
	//	}
	//}()
	//err = encryptForward(src, client)
	//if err != nil {
	//	return
	//}
	go func() {
		err := EncodeCopy(dst, client)
		if err != nil {
			return
		}
	}()
	go func() {
		err := DecodeCopy(client, dst)
		if err != nil {
			return
		}
	}()
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

// 从dst读取数据加密传给src
func encryptForward(dst, src net.Conn) error {
	//fmt.Println("encryptForward")
	for {
		buf := make([]byte, 2048)
		read, err := dst.Read(buf)

		if err != nil {
			//if err == io.EOF {
			//	return err
			//}
			log.Println(err)
			return err
		}

		fmt.Println("3: 加密前/读取: ", read)
		if read > 0 {
			//write, err := encryptWrite(src, buf[:read])
			write, err := src.Write(buf[:read])
			//encrypt, err := AesEncrypt(buf[:read], key)
			//write, err := src.Write(encrypt)

			fmt.Println("3: 加密后/发送: ", write)
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
		//buf := make([]byte, 3072)
		buf := make([]byte, 2048)
		read, err := dst.Read(buf)
		if err != nil {
			log.Println(err)
			return err
		}
		//buf, read, err := decryptRead(dst)
		//decrypt, err := AesDecrypt(buf[:read], key)

		fmt.Println("2: 解密前/读取: ", read)
		if read > 0 {
			//write, err := src.Write(buf2)
			write, err := src.Write(buf[:read])
			//write, err := src.Write(decrypt)

			fmt.Println("2: 解密后//发送: ", write)
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
		} else if split[0] == "iv" {
			iv = []byte(split[1])
		}
	}
	return nil
}
