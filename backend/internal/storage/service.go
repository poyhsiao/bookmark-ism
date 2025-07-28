package storage

import (
	"context"
	"time"

	"bookmark-sync-service/backend/pkg/storage"
)

// StorageClient defines the interface for storage operations
type StorageClient interface {
	StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error)
	StoreAvatar(ctx context.Context, userID string, data []byte, contentType string) (string, error)
	StoreBackup(ctx context.Context, userID string, data []byte) (string, error)
	GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
	HealthCheck(ctx context.Context) error
}

// Service provides storage operations for the application
type Service struct {
	client StorageClient
}

// NewService creates a new storage service
func NewService(client *storage.Client) *Service {
	return &Service{
		client: client,
	}
}

// StoreScreenshot stores a screenshot for a bookmark
func (s *Service) StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error) {
	return s.client.StoreScreenshot(ctx, bookmarkID, data)
}

// StoreAvatar stores a user avatar
func (s *Service) StoreAvatar(ctx context.Context, userID string, data []byte, contentType string) (string, error) {
	return s.client.StoreAvatar(ctx, userID, data, contentType)
}

// StoreBackup stores a backup file
func (s *Service) StoreBackup(ctx context.Context, userID string, data []byte) (string, error) {
	return s.client.StoreBackup(ctx, userID, data)
}

// GetFileURL gets a presigned URL for a file
func (s *Service) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	return s.client.GetFileURL(ctx, objectName, expiry)
}

// DeleteFile deletes a file from storage
func (s *Service) DeleteFile(ctx context.Context, objectName string) error {
	return s.client.DeleteFile(ctx, objectName)
}

// HealthCheck checks if storage is healthy
func (s *Service) HealthCheck(ctx context.Context) error {
	return s.client.HealthCheck(ctx)
}
