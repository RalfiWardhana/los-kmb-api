package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHmac(secretKey, content string) string {
	secret := []byte(secretKey)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
