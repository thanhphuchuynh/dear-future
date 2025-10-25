// Package storage provides storage adapter implementations
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thanhphuchuynh/dear-future/pkg/domain/common"
	"github.com/thanhphuchuynh/dear-future/pkg/effects"
)

// R2StorageConfig holds configuration for Cloudflare R2 storage
type R2StorageConfig struct {
	AccountID       string // Cloudflare account ID
	AccessKeyID     string // R2 API token ID
	SecretAccessKey string // R2 API token secret
	BucketName      string // R2 bucket name
	PublicURL       string // Optional public URL for the bucket
}

// R2Storage implements StorageService using Cloudflare R2 (S3-compatible)
type R2Storage struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2Storage creates a new R2 storage adapter
// R2 is S3-compatible, so we use the AWS SDK with R2 endpoints
func NewR2Storage(config R2StorageConfig) (*R2Storage, error) {
	if config.AccountID == "" || config.AccessKeyID == "" || config.SecretAccessKey == "" || config.BucketName == "" {
		return nil, fmt.Errorf("R2 configuration is incomplete: accountID, accessKeyID, secretAccessKey, and bucketName are required")
	}

	// R2 endpoint format: https://<account_id>.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.AccountID)

	// Create S3 client configured for R2
	s3Config := aws.Config{
		Region:      "auto", // R2 uses "auto" region
		Credentials: credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""),
	}

	client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // R2 requires path-style addressing
	})

	return &R2Storage{
		client:     client,
		bucketName: config.BucketName,
		publicURL:  config.PublicURL,
	}, nil
}

// UploadFile uploads a file to R2
func (r *R2Storage) UploadFile(ctx context.Context, fileData effects.FileUpload) common.Result[effects.FileUploadResult] {
	// Generate a unique key based on timestamp and filename
	key := fmt.Sprintf("uploads/%d-%s", time.Now().Unix(), fileData.FileName)

	// Prepare metadata
	metadata := make(map[string]string)
	for k, v := range fileData.Metadata {
		metadata[k] = v
	}

	// Upload to R2
	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(r.bucketName),
		Key:           aws.String(key),
		Body:          bytes.NewReader(fileData.Data),
		ContentLength: aws.Int64(fileData.Size),
		ContentType:   aws.String(fileData.ContentType),
		Metadata:      metadata,
	}

	output, err := r.client.PutObject(ctx, putInput)
	if err != nil {
		return common.Err[effects.FileUploadResult](fmt.Errorf("failed to upload file to R2: %w", err))
	}

	// Build the URL
	url := r.buildURL(key)

	result := effects.FileUploadResult{
		Key:         key,
		URL:         url,
		Size:        fileData.Size,
		ContentType: fileData.ContentType,
		UploadedAt:  time.Now(),
		ETag:        aws.ToString(output.ETag),
	}

	return common.Ok(result)
}

// DownloadFile downloads a file from R2
func (r *R2Storage) DownloadFile(ctx context.Context, key string) common.Result[effects.FileDownload] {
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	output, err := r.client.GetObject(ctx, getInput)
	if err != nil {
		return common.Err[effects.FileDownload](fmt.Errorf("failed to download file from R2: %w", err))
	}
	defer output.Body.Close()

	// Read the entire file into memory
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return common.Err[effects.FileDownload](fmt.Errorf("failed to read file data: %w", err))
	}

	// Extract filename from key (last part of the path)
	fileName := key
	if len(key) > 0 {
		for i := len(key) - 1; i >= 0; i-- {
			if key[i] == '/' {
				fileName = key[i+1:]
				break
			}
		}
	}

	result := effects.FileDownload{
		FileName:     fileName,
		ContentType:  aws.ToString(output.ContentType),
		Data:         data,
		Size:         aws.ToInt64(output.ContentLength),
		LastModified: aws.ToTime(output.LastModified),
	}

	return common.Ok(result)
}

// GeneratePresignedURL generates a presigned URL for temporary access
func (r *R2Storage) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) common.Result[string] {
	presignClient := s3.NewPresignClient(r.client)

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	presignedURL, err := presignClient.PresignGetObject(ctx, getInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return common.Err[string](fmt.Errorf("failed to generate presigned URL: %w", err))
	}

	return common.Ok(presignedURL.URL)
}

// DeleteFile deletes a file from R2
func (r *R2Storage) DeleteFile(ctx context.Context, key string) common.Result[bool] {
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	_, err := r.client.DeleteObject(ctx, deleteInput)
	if err != nil {
		return common.Err[bool](fmt.Errorf("failed to delete file from R2: %w", err))
	}

	return common.Ok(true)
}

// GetFileMetadata retrieves metadata about a file
func (r *R2Storage) GetFileMetadata(ctx context.Context, key string) common.Result[effects.FileMetadata] {
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	output, err := r.client.HeadObject(ctx, headInput)
	if err != nil {
		return common.Err[effects.FileMetadata](fmt.Errorf("failed to get file metadata: %w", err))
	}

	metadata := make(map[string]string)
	for k, v := range output.Metadata {
		metadata[k] = v
	}

	result := effects.FileMetadata{
		Key:          key,
		Size:         aws.ToInt64(output.ContentLength),
		ContentType:  aws.ToString(output.ContentType),
		LastModified: aws.ToTime(output.LastModified),
		ETag:         aws.ToString(output.ETag),
		Metadata:     metadata,
	}

	return common.Ok(result)
}

// buildURL constructs the URL for an object
func (r *R2Storage) buildURL(key string) string {
	if r.publicURL != "" {
		return fmt.Sprintf("%s/%s", r.publicURL, key)
	}
	// If no public URL is configured, return a placeholder
	return fmt.Sprintf("https://%s.r2.dev/%s", r.bucketName, key)
}
