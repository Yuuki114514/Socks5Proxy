package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"net"
)

const (
	BlockSize = 128
)

var (
	iv  = []byte
	key []byte
)

/////
func Pad(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)
	repeat := bytes.Repeat([]byte{0}, padding)
	return append(plainText, repeat...)
}

func UnPad(cipherText []byte, padding uint8) []byte {
	return cipherText[:BlockSize-padding]
}

func encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	text := Pad(plainText, BlockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	encryptedText := make([]byte, len(text))
	blockMode.CryptBlocks(encryptedText, text)
	return encryptedText, nil
}

func decrypt(encryptedText []byte, padding uint8) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(encryptedText))
	blockMode.CryptBlocks(plainText, encryptedText)
	unPad := UnPad(plainText, padding)
	return unPad, nil
}

func encryptWrite(conn net.Conn, plainText []byte) (int, error) {
	padding := BlockSize - len(plainText)
	_, err := conn.Write([]byte{byte(padding)})
	if err != nil {
		log.Println("send padding wrong", err)
		return 0, err
	}
	encryptedText, err := encrypt(plainText)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	write, err := conn.Write(encryptedText)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return write, nil
}

func decryptRead(conn net.Conn, encryptedText []byte) ([]byte, int, error) {
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	read, err := conn.Read(encryptedText)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	plainText, err := decrypt(encryptedText[:read], buf[0])
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}
	return plainText, len(plainText), nil
}

func Encrypt(text []byte) ([]byte, error) {
	//生成cipher.Block 数据块
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("错误 -" + err.Error())
		return nil, err
	}
	//填充内容，如果不足16位字符
	originData := pad(text, BlockSize)
	//加密方式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	//加密，输出到[]byte数组
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)
	// log.Println(crypted)
	return crypted, nil
}

func pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)
	padtext := bytes.Repeat([]byte{0x0}, padding)
	return append(ciphertext, padtext...)
}

func Decrypt(decode_data []byte, sign int) ([]byte, error) {
	//生成密码数据块cipher.Block
	block, _ := aes.NewCipher(key)
	//解密模式
	blockMode := cipher.NewCBCDecrypter(block, iv)
	//输出到[]byte数组
	origin_data := make([]byte, len(decode_data))
	blockMode.CryptBlocks(origin_data, decode_data)
	// log.Println(origin_data)
	// log.Println(unpad(origin_data))
	//去除填充,并返回
	return unpad(origin_data, sign), nil
}

func unpad(ciphertext []byte, sign int) []byte {
	unpadSign := BlockSize - sign
	return ciphertext[:unpadSign]
}

func EncodeWrite(conn net.Conn, bs []byte) (n int, err error) {
	sign := BlockSize - len(bs)
	conn.Write([]byte{byte(sign)})
	cipherText, err := Encrypt(bs)
	if err != nil {
		return
	}
	return conn.Write(cipherText)
}

func DecodeRead(conn net.Conn, bs []byte) (plainText []byte, n int, err error) {
	_, err = conn.Read(bs[:1])
	if err != nil {
		return
	}
	sign := int(bs[0])
	n, err = conn.Read(bs)
	if err != nil {
		return
	}
	plainText, err = Decrypt(bs[0:n], sign)
	if err != nil {
		return
	}
	return plainText, len(plainText), nil
}

func EncodeCopy(src net.Conn, dst net.Conn) error {
	buffer := make([]byte, BlockSize)
	for {
		readCount, readErr := src.Read(buffer)
		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			} else {
				return nil
			}
		}
		if readCount > 0 {
			_, writeErr := EncodeWrite(dst, buffer[0:readCount])
			if writeErr != nil {
				return writeErr
			}
		}
	}
}

func DecodeCopy(src net.Conn, dst net.Conn) error {
	buffer := make([]byte, BlockSize)
	for {
		plainText, readCount, readErr := DecodeRead(src, buffer)
		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			} else {
				return nil
			}
		}
		if readCount > 0 {
			writeCount, writeErr := dst.Write(plainText)
			if writeErr != nil {
				return writeErr
			}
			if readCount != writeCount {
				log.Printf("DecodeCopy:readCount:%d\n", readCount)
				log.Printf("DecodeCopy:writecount:%d\n", writeCount)
				return io.ErrShortWrite
			}
		}
	}
}
