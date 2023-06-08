package maskingdata

// Cipher to use for implementing the BlockCipher interface
type Cipher struct {
	BlockCipher
}

type BlockCipher interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	KeySize() int
}

func NewCipher(privateKey string) *Cipher {
	c := &Cipher{}
	key := []byte(privateKey)
	c.BlockCipher = newCFB(key)
	return c
}

func (c *Cipher) EncryptText(text string) (string, error) {
	enc, err := c.Encrypt([]byte(text))
	if err != nil {
		return "", err
	}

	return c.EncodeString(enc), nil
}

func (c *Cipher) DecryptText(text string) (string, error) {
	decode, err := c.DecodingString(text)
	if err != nil {
		return "", err
	}

	decrypt, err := c.Decrypt(decode)
	return string(decrypt), err
}
