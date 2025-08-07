package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of Redis client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) PublishSyncEvent(ctx context.Context, userID string, event interface{}) error {
	args := m.Called(ctx, userID, event)
	return args.Error(0)
}

func (m *MockRedisClient) SubscribeToSyncEvents(ctx context.Context, userID string) interface{} {
	args := m.Called(ctx, userID)
	return args.Get(0)
}

func (m *MockRedisClient) AssertExpected(t *testing.T) {
	m.Mock.AssertExpectations(t)
}
