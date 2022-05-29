package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"net"
)

var iv = []byte("3010201735544643")

const (
	BlockSize = 128
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AES加密,CBC
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
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
