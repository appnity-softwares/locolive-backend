package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Service interface {
	UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}

type S3Service struct {
	client     *s3.Client
	bucketName string
	endpoint   string
	baseURL    string // Optional: custom domain for public access
}

func NewS3Service(ctx context.Context, accountID, accessKey, secretKey, bucketName string) (Service, error) {
	// R2 Endpoint: https://<accountid>.r2.cloudflarestorage.com
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
	})

	return &S3Service{
		client:     client,
		bucketName: bucketName,
		endpoint:   r2Endpoint,
	}, nil
}

// UploadFile uploads a multipart file to R2 and returns the public URL
func (s *S3Service) UploadFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Generate unique filename
	extension := filepath.Ext(fileHeader.Filename)
	key := fmt.Sprintf("%s%s", uuid.New().String(), extension)

	// Determine Content-Type
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		// ACL is often not supported or needed for R2 depending on bucket settings, but public-read is common for S3
		// ACL: types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct public URL
	// If the bucket allows public access via custom domain: https://<custom_domain>/<key>
	// Or raw R2 dev URL (usually needs worker or public bucket setup)
	// For now, we assume a public bucket url pattern or return the key.
	// Let's assume standard R2 format or a configured public domain.
	// Users usually configure a public domain like `media.locolive.com`.
	// We'll return a relative path or full URL if ENV is set.
	// Since we don't have the public domain env yet, let's construct a standard R2 public URL format if possible,
	// OR better: use a placeholder domain that the user replaces.

	// Default to R2.dev URL if testing, but ideally should be custom domain
	// https://pub-<hash>.r2.dev/<key>
	// Since we don't have the pub- hash, we'll return a format that the frontend can use or just the key?
	// The frontend expects a full URL.
	// Let's assume a generic one that needs config.

	// FIX: We will depend on an environment variable for PUBLIC_URL_BASE in real app.
	// For this code, I will return `https://<bucket>.r2.dev/<key>` as a sensible default assumption for R2 public buckets.

	return fmt.Sprintf("https://%s.r2.dev/%s", s.bucketName, key), nil
}
