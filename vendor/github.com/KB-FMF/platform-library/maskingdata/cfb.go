package maskingdata

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

const (
	// KeySize for CFB uses the generic key size
	KeySize = 32
)

type CipherCFB struct {
	key []byte
}

func newCFB(key []byte) *CipherCFB {
	return &CipherCFB{key: key}
}

// Encrypt implements the BlockCipher interface
func (c *CipherCFB) Encrypt(plaintext []byte) ([]byte, error) {
	return encrypt(c.key, plaintext)
}

// Decrypt implements the BlockCipher interface
func (c *CipherCFB) Decrypt(ciphertext []byte) ([]byte, error) {
	return decrypt(c.key, ciphertext)
}

// KeySize implements the BlockCipher interface
func (c *CipherCFB) KeySize() int {
	return KeySize
}

// encrypt plaintext using the given key with CTR encryption
func encrypt(key, plaintext []byte) ([]byte, error) {

	if (len(key) < KeySize || len(key) > KeySize) || len(string(plaintext)) <= 0 {
		return nil, ErrInvalidKeyLength
	}

	c, _ := aes.NewCipher(key)
	ct := make([]byte, aes.BlockSize+len(string(plaintext)))
	iv := ct[:aes.BlockSize]

	_, err := io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(c, iv)
	cfb.XORKeyStream(ct[aes.BlockSize:], plaintext)

	return ct, nil
}

// decrypt ciphertext using the given key
func decrypt(key, ciphertext []byte) ([]byte, error) {
	if (len(key) < KeySize || len(key) > KeySize) || len(string(ciphertext)) <= 0 {
		return nil, ErrInvalidKeyLength
	}

	if len(string(ciphertext)) < aes.BlockSize {
		return nil, ErrInvalidMessageShort
	}

	c, _ := aes.NewCipher(key)
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(c, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
