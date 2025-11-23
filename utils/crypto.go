package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// Get encryption key
func GetEncryptionKey() ([]byte, error) {
	key := os.Getenv("ENCRYPT_KEY")
	if key == "" {
		return nil, errors.New("ENCRYPT_KEY tidak ditemukan di environment variable")
	}

	// Decode key
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("gagal decode ENCRYPT_KEY: %v", err)
	}

	// Validate key length
	if len(decodedKey) != 32 {
		return nil, fmt.Errorf("ENCRYPT_KEY harus 32 byte setelah decode, tetapi dapat %d byte", len(decodedKey))
	}

	return decodedKey, nil
}

// Encrypt text
func EncryptWithAES(plaintext string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt text
func DecryptWithAES(cipherText string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	// Decode ciphertext
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Validate length
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext terlalu pendek")
	}

	// Extract nonce and ciphertext
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
