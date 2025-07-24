package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the MinIO client with additional functionality
type Client struct {
	client     *minio.Client
	config     *config.StorageConfig
	bucketName string
}

// NewClient creates a new MinIO client
func NewClient(cfg config.StorageConfig) (*Client, error) {
	// Initialize MinIO client (S3-compatible API)
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{
		client:     client,
		config:     &cfg,
		bucketName: cfg.BucketName,
	}, nil
}

// HealthCheck checks if MinIO is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.BucketExists(ctx, c.bucketName)
	return err
}

// EnsureBucketExists creates the bucket if it doesn't exist
func (c *Client) EnsureBucketExists(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO and returns the object name
func (c *Client) UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return a URL that can be used to access the file
	// In a production environment, you might want to return a presigned URL
	// For now, we'll return the object path
	return fmt.Sprintf("/storage/%s", objectName), nil
}

// DownloadFile downloads a file from MinIO
func (c *Client) DownloadFile(ctx context.Context, objectName string) ([]byte, error) {
	object, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

// DeleteFile deletes a file from MinIO
func (c *Client) DeleteFile(ctx context.Context, objectName string) error {
	err := c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileURL gets a presigned URL for a file
func (c *Client) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	return url.String(), nil
}

// ListFiles lists files in a directory
func (c *Client) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := c.client.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// StoreScreenshot stores a screenshot for a bookmark
func (c *Client) StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error) {
	objectName := fmt.Sprintf("screenshots/%s.png", bookmarkID)
	url, err := c.UploadFile(ctx, objectName, data, "image/png")
	if err != nil {
		return "", err
	}
	return url, nil
}

// StoreAvatar stores a user avatar
func (c *Client) StoreAvatar(ctx context.Context, userID string, data []byte, contentType string) (string, error) {
	objectName := fmt.Sprintf("avatars/%s", userID)
	url, err := c.UploadFile(ctx, objectName, data, contentType)
	if err != nil {
		return "", err
	}
	return url, nil
}

// StoreBackup stores a backup file
func (c *Client) StoreBackup(ctx context.Context, userID string, data []byte) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	objectName := fmt.Sprintf("backups/%s/%s.json", userID, timestamp)
	url, err := c.UploadFile(ctx, objectName, data, "application/json")
	if err != nil {
		return "", err
	}
	return url, nil
}
