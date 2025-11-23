package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

// Inisialisasi S3
func NewS3Client() *S3Client {
	// Ambil env
	endpoint := os.Getenv("OS_ENDPOINT_URL")
	accessKeyID := os.Getenv("OS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("OS_SECRET_ACCESS_KEY")
	region := os.Getenv("OS_REGION")
	bucket := os.Getenv("OS_BUCKET_NAME")

	// Resolver kustom
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: region,
		}, nil
	})

	// Konfigurasi AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Buat klien
	client := s3.NewFromConfig(cfg)

	return &S3Client{
		client: client,
		bucket: bucket,
	}
}

// Ambil klien
func (s *S3Client) GetClient() *s3.Client {
	return s.client
}

// Ambil bucket
func (s *S3Client) GetBucket() string {
	return s.bucket
}

// Upload file
func (s *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Buat URL
	endpoint := os.Getenv("OS_ENDPOINT_URL")
	url := fmt.Sprintf("https://%s.%s/%s", s.bucket, strings.TrimPrefix(endpoint, "https://"), key)
	return url, nil
}

// Hapus file
func (s *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Buat URL
func (s *S3Client) GetFileURL(key string) string {
	endpoint := os.Getenv("OS_ENDPOINT_URL")
	return fmt.Sprintf("https://%s.%s/%s", s.bucket, strings.TrimPrefix(endpoint, "https://"), key)
}
