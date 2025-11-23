package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"backend/storage"
)

// SaveUploadToS3 saves uploaded file to S3
func SaveUploadToS3(file multipart.File, header *multipart.FileHeader, subDir string) (string, error) {
	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".pdf"}
	valid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			valid = true
			break
		}
	}
	if !valid {
		return "", errors.New("ekstensi tidak diizinkan")
	}

	// Check size
	maxSize := int64(2 * 1024 * 1024)
	if header.Size > maxSize {
		return "", errors.New("ukuran file terlalu besar")
	}

	// Generate filename
	randomStr, err := generateRandomString(8)
	if err != nil {
		return "", fmt.Errorf("gagal generate nama file: %v", err)
	}

	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s%s", timestamp, randomStr, ext)
	key := fmt.Sprintf("%s/%s", subDir, filename)

	// Read file content
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("gagal membaca file: %v", err)
	}

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Default content type based on extension
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".pdf":
			contentType = "application/pdf"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Upload to S3
	s3Client := storage.NewS3Client()
	url, err := s3Client.UploadFile(context.Background(), key, buf.Bytes(), contentType)
	if err != nil {
		return "", fmt.Errorf("gagal upload file: %v", err)
	}

	return url, nil
}

// DeleteFileFromS3 deletes a file from S3
func DeleteFileFromS3(url string) error {
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return errors.New("URL tidak valid")
	}

	key := strings.Join(parts[4:], "/")

	s3Client := storage.NewS3Client()
	return s3Client.DeleteFile(context.Background(), key)
}

// Generate random string
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:length], nil
}
