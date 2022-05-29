package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"net"
	"time"
)

const (
	BlockSize = 128
)

var (
	iv  []byte
	key []byte
)

func encrypt(plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	text := pad(plainText, BlockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	encryptedText := make([]byte, len(text))
	blockMode.CryptBlocks(encryptedText, text)
	return encryptedText, nil
}

func pad(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)
	repeat := bytes.Repeat([]byte{0x0}, padding)
	return append(plainText, repeat...)
}

func decrypt(encryptedText []byte, padding int) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(encryptedText))
	blockMode.CryptBlocks(plainText, encryptedText)
	text := unPad(plainText, padding)
	return text, nil
}

func unPad(ciphertext []byte, padding int) []byte {
	return ciphertext[:BlockSize-padding]
}

func encryptWrite(conn net.Conn, plainText []byte) (int, error) {
	padding := BlockSize - len(plainText)
	_, err := conn.Write([]byte{byte(padding)})
	if err != nil {
		log.Println(err)
		return 0, err
	}
	encryptedText, err := encrypt(plainText)
	if err != nil {
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
	_, err := conn.Read(encryptedText[:1])
	if err != nil {
		return nil, 0, err
	}
	padding := int(encryptedText[0])
	read, err := conn.Read(encryptedText)
	if err != nil {
		return nil, 0, err
	}
	plainText, err := decrypt(encryptedText[:read], padding)
	if err != nil {
		return nil, 0, err
	}
	return plainText, len(plainText), nil
}

//从src读取数据加密后传给dst
func encryptForward(src, dst net.Conn) error {
	buffer := make([]byte, BlockSize)
	for {
		err := src.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err != nil {
			log.Println(err)
			return err
		}
		read, err := src.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return nil
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return netErr
			} else {
				return err
			}
		}
		if read > 0 {
			_, err := encryptWrite(dst, buffer[:read])
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
}

//从src读取数据解密后传给dst
func decryptForward(src, dst net.Conn) error {
	buffer := make([]byte, BlockSize)
	for {
		plainText, read, err := decryptRead(src, buffer)
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				return nil
			}
		}
		if read > 0 {
			_, err := dst.Write(plainText)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
}
