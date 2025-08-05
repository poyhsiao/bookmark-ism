package community

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestMockDB provides a properly configured mock database for testing
type TestMockDB struct {
	mock.Mock
}

func (m *TestMockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	if err := args.Error(0); err != nil {
		return &gorm.DB{Error: err}
	}
	return &gorm.DB{Error: nil}
}

func (m *TestMockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if err := args.Error(0); err != nil {
		return &gorm.DB{Error: err}
	}
	return &gorm.DB{Error: nil}
}

func (m *TestMockDB) Where(query interface{}, args ...interface{}) Database {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(Database)
}

func (m *TestMockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if err := args.Error(0); err != nil {
		return &gorm.DB{Error: err}
	}
	return &gorm.DB{Error: nil}
}

func (m *TestMockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	if err := args.Error(0); err != nil {
		return &gorm.DB{Error: err}
	}
	return &gorm.DB{Error: nil}
}

func (m *TestMockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	if err := args.Error(0); err != nil {
		return &gorm.DB{Error: err}
	}
	return &gorm.DB{Error: nil}
}

func (m *TestMockDB) Order(value interface{}) Database {
	args := m.Called(value)
	return args.Get(0).(Database)
}

func (m *TestMockDB) Limit(limit int) Database {
	args := m.Called(limit)
	return args.Get(0).(Database)
}

func (m *TestMockDB) Offset(offset int) Database {
	args := m.Called(offset)
	return args.Get(0).(Database)
}

// TestMockRedisClient provides a properly configured mock Redis client for testing
type TestMockRedisClient struct {
	mock.Mock
}

func (m *TestMockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *TestMockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *TestMockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *TestMockRedisClient) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *TestMockRedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

// TestMockWorkerPool provides a properly configured mock worker pool for testing
type TestMockWorkerPool struct {
	mock.Mock
}

func (m *TestMockWorkerPool) Submit(job interface{}) error {
	args := m.Called(job)
	return args.Error(0)
}

// Helper functions for setting up common mock expectations

// SetupMockDBForFind sets up a mock DB to return specific data for Find operations
func SetupMockDBForFind(mockDB *TestMockDB, data interface{}) {
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Order", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Offset", mock.Anything).Return(mockDB)
	mockDB.On("Find", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// This would populate the destination with test data
		// Implementation depends on the specific test case
	}).Return(nil)
}

// SetupMockDBForFirst sets up a mock DB to return specific data for First operations
func SetupMockDBForFirst(mockDB *TestMockDB, data interface{}, err error) {
	mockDB.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if err == nil && data != nil {
			// Populate destination with test data
			// Implementation depends on the specific test case
		}
	}).Return(err)
}
