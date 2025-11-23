package handlers

import (
	"backend/storage"
	"backend/utils"
	"context"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
)

// File handler
type FileHandler struct{}

// Constructor
func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// ServeFile serves file from S3
func (h *FileHandler) ServeFile(c *fiber.Ctx) error {
	// Ambil path
	path := c.Params("*")
	if path == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Path tidak valid", nil)
	}

	// Cek keamanan
	if !strings.HasPrefix(path, "uploads/") {
		log.Println("Invalid file access attempt:", path)
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", nil)
	}

	// Cegah traversal
	if strings.Contains(path, "..") {
		log.Println("Directory traversal attempt:", path)
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", nil)
	}

	// Ekstrak key
	var key string
	if strings.HasPrefix(path, "uploads/bukti_transfer/") {
		key = strings.TrimPrefix(path, "uploads/")
	} else {
		filename := strings.TrimPrefix(path, "uploads/")
		key = "bukti_transfer/" + filename
	}

	// Koneksi S3
	s3Client := storage.NewS3Client()

	// Ambil objek
	resp, err := s3Client.GetClient().GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3Client.GetBucket()),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to get object from S3: %v", err)
		return utils.ErrorResponse(c, fiber.StatusNotFound, "File tidak ditemukan", err)
	}
	defer resp.Body.Close()

	// Baca konten
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read object from S3: %v", err)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membaca file", err)
	}

	// Deteksi ekstensi
	ext := strings.ToLower(filepath.Ext(path))

	// Set tipe
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".pdf":
		contentType = "application/pdf"
	default:
		contentType = "application/octet-stream"
	}

	// Header tipe
	c.Set("Content-Type", contentType)

	// Mode download
	if c.Query("download") == "true" {
		c.Set("Content-Disposition", "attachment; filename=\""+filepath.Base(path)+"\"")
	} else {
		c.Set("Content-Disposition", "inline; filename=\""+filepath.Base(path)+"\"")
	}

	return c.Send(body)
}

// GetFileInfo returns file info from S3
func (h *FileHandler) GetFileInfo(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Path tidak valid", nil)
	}

	// Validasi path
	if !strings.HasPrefix(path, "uploads/") {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", nil)
	}

	// Cek traversal
	if strings.Contains(path, "..") {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", nil)
	}

	// Ekstrak key
	var key string
	if strings.HasPrefix(path, "uploads/bukti_transfer/") {
		key = strings.TrimPrefix(path, "uploads/")
	} else {
		filename := strings.TrimPrefix(path, "uploads/")
		key = "bukti_transfer/" + filename
	}

	// Koneksi S3
	s3Client := storage.NewS3Client()

	// Ambil metadata
	resp, err := s3Client.GetClient().HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s3Client.GetBucket()),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to get object head from S3: %v", err)
		return utils.ErrorResponse(c, fiber.StatusNotFound, "File tidak ditemukan", err)
	}

	// Deteksi tipe
	ext := strings.ToLower(filepath.Ext(path))
	var fileType string
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		fileType = "image"
	case ".pdf":
		fileType = "document"
	default:
		fileType = "unknown"
	}

	response := map[string]interface{}{
		"name":      filepath.Base(path),
		"size":      resp.ContentLength,
		"type":      fileType,
		"extension": ext,
		"path":      path,
		"url":       s3Client.GetFileURL(key),
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Info file berhasil diambil", response)
}
