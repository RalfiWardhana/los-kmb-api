package maskingdata

import "errors"

var (
	// ErrInvalidKeyLength occurs when a key has been used with an invalid length
	ErrInvalidKeyLength = errors.New("cipher: invalid key length")
	// ErrInvalidMessageShort occurs when a text less than BlockSize
	ErrInvalidMessageShort = errors.New("cipher: ciphertext too short")
)
