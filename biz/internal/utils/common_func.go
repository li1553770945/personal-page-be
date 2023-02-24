package utils

import "crypto/sha256"

func Sha256(str string) string {
	b := sha256.Sum256([]byte(str))
	return string(b[:])
}
