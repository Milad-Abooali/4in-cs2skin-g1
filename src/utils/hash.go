package utils

import (
	"crypto/sha1"
	"encoding/hex"
)

func Sha1Hash(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}
