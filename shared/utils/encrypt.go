package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/KB-FMF/platform-library/maskingdata"
)

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

func decrypt(key, iv []byte, encrypted string) ([]byte, error) {
	data, _ := base64.StdEncoding.DecodeString(encrypted)

	if len(string(data))%aes.BlockSize != 0 {
		return nil, fmt.Errorf("bad blocksize(%v), aes.BlockSize = %v\n", len(data), aes.BlockSize)
	}

	c, err := aes.NewCipher(key)

	if err != nil {
		fmt.Println("error chipper ", err.Error())
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks([]byte(data), []byte(data))
	out, err := pkcs7Unpad([]byte(data))
	if err != nil {
		return out, err
	}
	return out, nil
}

func DecryptCredential(encryptedText string) (string, error) {
	data, err := base64.RawStdEncoding.DecodeString(encryptedText)
	if err != nil {
		fmt.Println(err.Error())
	}

	s := strings.Split(string(data), ":")
	src, iv, key := s[0], s[1], s[2]

	keys, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		fmt.Println(err.Error())
	}

	decodeIv, err := hex.DecodeString(iv)

	decryptedText, err := decrypt(keys, decodeIv, src)
	return string(decryptedText), err
}

type UtilsInterface interface {
	PlatformEncryptText(myString string) (string, error)
	PlatformDecryptText(myString string) (string, error)
}

type Utils struct{}

func NewUtils() UtilsInterface {
	ut := &Utils{}
	return ut
}

func (u *Utils) PlatformEncryptText(myString string) (string, error) {
	cipher := maskingdata.NewCipher(os.Getenv("PLATFORM_LIBRARY_KEY"))
	return cipher.EncryptText(myString)
}

func (u *Utils) PlatformDecryptText(encryptedText string) (string, error) {
	cipher := maskingdata.NewCipher(os.Getenv("PLATFORM_LIBRARY_KEY"))
	return cipher.DecryptText(encryptedText)
}
