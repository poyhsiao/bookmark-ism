package content

import (
	"testing"

	"bookmark-sync-service/backend/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	service := NewService()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpiryHour: 24,
		},
	}
	handler := NewHandler(service, cfg)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.service)
	assert.NotNil(t, handler.cfg)
}
