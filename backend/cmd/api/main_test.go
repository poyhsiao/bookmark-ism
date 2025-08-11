package main

import (
	"testing"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/storage"

	"github.com/stretchr/testify/assert"
)

// TestStorageClientIntegration tests that the storage client can be created from app config
func TestStorageClientIntegration(t *testing.T) {
	// Create a sample config.StorageConfig
	storageConfig := config.StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "bookmarks",
		UseSSL:          false,
	}

	// Test that we can create a storage client using the adapter
	client, err := storage.NewClientFromConfig(storageConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
