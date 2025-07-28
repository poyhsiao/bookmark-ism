package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/disintegration/imaging"
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

// ImageOptimizationOptions defines options for image optimization
type ImageOptimizationOptions struct {
	MaxWidth      int
	MaxHeight     int
	Quality       int
	Format        string // "jpeg", "png", "webp"
	Thumbnail     bool
	ThumbnailSize int
}

// OptimizeAndStoreImage optimizes an image and stores it with optional thumbnail
func (c *Client) OptimizeAndStoreImage(ctx context.Context, objectName string, imageData []byte, opts ImageOptimizationOptions) (string, string, error) {
	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if needed
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		img = imaging.Fit(img, opts.MaxWidth, opts.MaxHeight, imaging.Lanczos)
	}

	// Encode optimized image
	var optimizedData bytes.Buffer
	var contentType string

	switch strings.ToLower(opts.Format) {
	case "jpeg", "jpg":
		quality := opts.Quality
		if quality == 0 {
			quality = 85 // Default quality
		}
		err = jpeg.Encode(&optimizedData, img, &jpeg.Options{Quality: quality})
		contentType = "image/jpeg"
	case "png":
		err = png.Encode(&optimizedData, img)
		contentType = "image/png"
	default:
		// Use original format
		switch format {
		case "jpeg":
			quality := opts.Quality
			if quality == 0 {
				quality = 85
			}
			err = jpeg.Encode(&optimizedData, img, &jpeg.Options{Quality: quality})
			contentType = "image/jpeg"
		case "png":
			err = png.Encode(&optimizedData, img)
			contentType = "image/png"
		default:
			return "", "", fmt.Errorf("unsupported image format: %s", format)
		}
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to encode optimized image: %w", err)
	}

	// Store optimized image
	mainURL, err := c.UploadFile(ctx, objectName, optimizedData.Bytes(), contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to store optimized image: %w", err)
	}

	var thumbnailURL string
	// Create and store thumbnail if requested
	if opts.Thumbnail {
		thumbnailSize := opts.ThumbnailSize
		if thumbnailSize == 0 {
			thumbnailSize = 200 // Default thumbnail size
		}

		thumbnail := imaging.Fit(img, thumbnailSize, thumbnailSize, imaging.Lanczos)

		var thumbnailData bytes.Buffer
		err = jpeg.Encode(&thumbnailData, thumbnail, &jpeg.Options{Quality: 80})
		if err != nil {
			return mainURL, "", fmt.Errorf("failed to encode thumbnail: %w", err)
		}

		// Generate thumbnail object name
		ext := filepath.Ext(objectName)
		nameWithoutExt := strings.TrimSuffix(objectName, ext)
		thumbnailObjectName := fmt.Sprintf("%s_thumb%s", nameWithoutExt, ext)

		thumbnailURL, err = c.UploadFile(ctx, thumbnailObjectName, thumbnailData.Bytes(), "image/jpeg")
		if err != nil {
			return mainURL, "", fmt.Errorf("failed to store thumbnail: %w", err)
		}
	}

	return mainURL, thumbnailURL, nil
}

// StoreOptimizedScreenshot stores an optimized screenshot with thumbnail
func (c *Client) StoreOptimizedScreenshot(ctx context.Context, bookmarkID string, imageData []byte) (string, string, error) {
	objectName := fmt.Sprintf("screenshots/%s.jpg", bookmarkID)

	opts := ImageOptimizationOptions{
		MaxWidth:      1200,
		MaxHeight:     800,
		Quality:       85,
		Format:        "jpeg",
		Thumbnail:     true,
		ThumbnailSize: 300,
	}

	return c.OptimizeAndStoreImage(ctx, objectName, imageData, opts)
}

// StoreBucketFile stores a file in a specific bucket directory
func (c *Client) StoreBucketFile(ctx context.Context, bucketType, fileName string, data []byte, contentType string) (string, error) {
	objectName := fmt.Sprintf("%s/%s", bucketType, fileName)
	return c.UploadFile(ctx, objectName, data, contentType)
}

// GetBucketFiles lists files in a specific bucket directory
func (c *Client) GetBucketFiles(ctx context.Context, bucketType string) ([]string, error) {
	return c.ListFiles(ctx, bucketType+"/")
}
