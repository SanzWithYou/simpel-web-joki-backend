package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateAESKey() (string, error) {
	// 32 bytes = AES-256
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}

	// Base64
	return base64.StdEncoding.EncodeToString(key), nil
}
