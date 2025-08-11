package main

import (
	"testing"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/storage"

	"github.com/stretchr/testify/assert"
)

// TestConfigStorageCompatibility tests that config.StorageConfig can be converted to storage.Config
func TestConfigStorageCompatibility(t *testing.T) {
	// Create a sample config.StorageConfig
	storageConfig := config.StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "bookmarks",
		UseSSL:          false,
	}

	// Test conversion function (we'll implement this)
	storageClientConfig := convertStorageConfig(storageConfig)

	// Verify the conversion worked
	assert.Equal(t, storageConfig.Endpoint, storageClientConfig.Endpoint)
	assert.Equal(t, storageConfig.AccessKeyID, storageClientConfig.AccessKeyID)
	assert.Equal(t, storageConfig.SecretAccessKey, storageClientConfig.SecretAccessKey)
	assert.Equal(t, storageConfig.BucketName, storageClientConfig.BucketName)
	assert.Equal(t, storageConfig.UseSSL, storageClientConfig.UseSSL)

	// Test that we can create a storage client with the converted config
	client, err := storage.NewClient(storageClientConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
