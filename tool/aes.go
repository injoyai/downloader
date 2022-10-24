package tool

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// EncryptCBC 字符串加密 CBC ,16位长度
func EncryptCBC(bytes, key []byte, ivs ...[]byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	var iv []byte
	if len(ivs) == 0 {
		iv = key
	} else {
		iv = ivs[0]
	}
	bytes = PKCS7Padding(bytes, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:blockSize])
	result := make([]byte, len(bytes))
	blockMode.CryptBlocks(result, bytes)
	return result, nil
}

// DecryptCBC 字符串解密 ,16位长度
func DecryptCBC(bytes, key []byte, ivs ...[]byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	var iv []byte
	if len(ivs) == 0 {
		iv = key
	} else {
		iv = ivs[0]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(bytes))
	blockMode.CryptBlocks(origData, bytes)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	text := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, text...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	padding := int(origData[length-1])
	return origData[:(length - padding)]
}
