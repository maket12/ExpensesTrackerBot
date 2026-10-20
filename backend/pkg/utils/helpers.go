package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func VPtr[T any](v T) *T {
	return &v
}

func ComparePtr(s1, s2 *string) bool {
	var v1, v2 string
	if s1 != nil {
		v1 = *s1
	}
	if s2 != nil {
		v2 = *s2
	}

	return v1 == v2
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
