package maskingdata

import "crypto/sha256"

// hashTo32Bytes will compute a cryptographically useful hash of the input string.
func hashTo32Bytes(input string) []byte {
	data := sha256.Sum256([]byte(input))
	return data[0:]
}
