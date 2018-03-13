// dec
package main

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/sha256"
	"encoding/base64"

	"io/ioutil"

	"fmt"
)

func Decrypt(file string, out string) {
	arg1 := sha256.Sum224([]byte("12345678910"))
	key := arg1[:24]
	plain, _ := ioutil.ReadFile(file)
	block, _ := des.NewTripleDESCipher(key)
	DecryptMode := cipher.NewCBCDecrypter(block, key[:8])
	plain, _ = base64.StdEncoding.DecodeString(string(plain))
	DecryptMode.CryptBlocks(plain, plain)
	plain = PKCS5remove(plain)
	err := ioutil.WriteFile(out, plain, 0600)
	if err != nil {
		fmt.Println("Decrypt Failed!")
	} else {
		fmt.Println("Decrypt Success!")
	}
}

func PKCS5remove(plaintext []byte) []byte {
	length := len(plaintext)
	num := int(plaintext[length-1])
	return plaintext[:(length - num)]
}
