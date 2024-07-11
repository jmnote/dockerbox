package util

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

func NewHash(length int) string {
	return Hash(uuid.New().String()+uuid.New().String(), length)
}

func Hash(s string, n int) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)

	base36Chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	var base36Hash bytes.Buffer
	for _, b := range hashBytes {
		base36Hash.WriteByte(base36Chars[b%36])
	}
	var result bytes.Buffer
	cnt := n / base36Hash.Len()
	remainder := n % base36Hash.Len()
	for i := 0; i < cnt; i++ {
		result.Write(base36Hash.Bytes())
	}
	result.WriteString(base36Hash.String()[:remainder])
	return result.String()
}
