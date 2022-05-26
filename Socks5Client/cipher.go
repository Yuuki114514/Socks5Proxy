package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"log"
	"net"
)

var (
	key []byte
)

// Padding 对明文进行填充
func Padding(plainText []byte, blockSize int) []byte {
	//计算要填充的长度
	n := blockSize - len(plainText)%blockSize
	//对原来的明文填充n个n
	temp := bytes.Repeat([]byte{byte(n)}, n)
	plainText = append(plainText, temp...)
	return plainText
}

// UnPadding 对密文删除填充
func UnPadding(cipherText []byte) []byte {
	//取出密文最后一个字节end
	end := cipherText[len(cipherText)-1]
	//删除填充
	cipherText = cipherText[:len(cipherText)-int(end)]
	return cipherText
}

// AESEncrypt AEC加密（CBC模式）
func AESEncrypt(plainText []byte) []byte {
	//指定加密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//进行填充
	plainText = Padding(plainText, block.BlockSize())
	//指定初始向量vi,长度和block的块尺寸一致
	iv := []byte("12345678abcdefgh")
	//指定分组模式，返回一个BlockMode接口对象
	blockMode := cipher.NewCBCEncrypter(block, iv)
	//加密连续数据库
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	//返回密文
	return cipherText
}

// AESDecrypt AEC解密（CBC模式）
func AESDecrypt(cipherText []byte) []byte {
	//指定解密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	//指定初始化向量IV,和加密的一致
	iv := []byte("12345678abcdefgh")
	//指定分组模式，返回一个BlockMode接口对象
	blockMode := cipher.NewCBCDecrypter(block, iv)
	//解密
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	//删除填充
	plainText = UnPadding(plainText)
	return plainText
}

func encryptWrite(conn net.Conn, buf []byte) error {
	encrypt := AESEncrypt(buf)
	_, err := conn.Write(encrypt)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func decryptRead(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 2048)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	decrypt := AESDecrypt(buf)
	return decrypt, nil
}
