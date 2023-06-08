package maskingdata

import "encoding/base64"

func (c *Cipher) EncodeString(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func (c *Cipher) DecodingString(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
