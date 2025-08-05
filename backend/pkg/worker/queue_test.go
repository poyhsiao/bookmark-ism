package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestJob implements the Job interface for testing
type TestJob struct {
	BaseJob
	ExecuteFunc func(ctx context.Context) error
	executed    bool
	mu          sync.Mutex
}

func NewTestJob(id, jobType string, maxRetries int, executeFunc func(ctx context.Context) error) *TestJob {
	return &TestJob{
		BaseJob: BaseJob{
			ID:         id,
			Type:       jobType,
			MaxRetries: maxRetries,
			CreatedAt:  time.Now(),
		},
		ExecuteFunc: executeFunc,
	}
}

func (j *TestJob) Execute(ctx context.Context) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.executed = true

	if j.ExecuteFunc != nil {
		return j.ExecuteFunc(ctx)
	}
	return nil
}

func (j *TestJob) IsExecuted() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.executed
}

type WorkerPoolTestSuite struct {
	suite.Suite
	pool   *WorkerPool
	logger *zap.Logger
}

func (suite *WorkerPoolTestSuite) SetupTest() {
	suite.logger = zaptest.NewLogger(suite.T())
	suite.pool = NewWorkerPool(2, 10, suite.logger)
}

func (suite *WorkerPoolTestSuite) TearDownTest() {
	if suite.pool.IsStarted() {
		suite.pool.Stop()
	}
}

func (suite *WorkerPoolTestSuite) TestNewWorkerPool() {
	pool := NewWorkerPool(5, 100, suite.logger)

	assert.Equal(suite.T(), 5, pool.workers)
	assert.Equal(suite.T(), 100, cap(pool.jobQueue))
	assert.False(suite.T(), pool.IsStarted())
}

func (suite *WorkerPoolTestSuite) TestNewWorkerPool_DefaultValues() {
	pool := NewWorkerPool(0, 0, suite.logger)

	assert.Equal(suite.T(), 10, pool.workers)         // Default worker pool size
	assert.Equal(suite.T(), 1000, cap(pool.jobQueue)) // Default queue size
}

func (suite *WorkerPoolTestSuite) TestStartStop() {
	assert.False(suite.T(), suite.pool.IsStarted())

	suite.pool.Start()
	assert.True(suite.T(), suite.pool.IsStarted())

	suite.pool.Stop()
	assert.False(suite.T(), suite.pool.IsStarted())
}

func (suite *WorkerPoolTestSuite) TestStartAlreadyStarted() {
	suite.pool.Start()
	assert.True(suite.T(), suite.pool.IsStarted())

	// Starting again should not cause issues
	suite.pool.Start()
	assert.True(suite.T(), suite.pool.IsStarted())
}

func (suite *WorkerPoolTestSuite) TestStopNotStarted() {
	assert.False(suite.T(), suite.pool.IsStarted())

	// Stopping when not started should not cause issues
	suite.pool.Stop()
	assert.False(suite.T(), suite.pool.IsStarted())
}

func (suite *WorkerPoolTestSuite) TestSubmitJob_Success() {
	suite.pool.Start()

	job := NewTestJob("test-1", "test", 3, nil)
	err := suite.pool.Submit(job)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, suite.pool.GetQueueSize())
}

func (suite *WorkerPoolTestSuite) TestSubmitJob_NotStarted() {
	job := NewTestJob("test-1", "test", 3, nil)
	err := suite.pool.Submit(job)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "worker pool not started")
}

func (suite *WorkerPoolTestSuite) TestJobExecution() {
	suite.pool.Start()

	executed := false
	job := NewTestJob("test-1", "test", 3, func(ctx context.Context) error {
		executed = true
		return nil
	})

	err := suite.pool.Submit(job)
	assert.NoError(suite.T(), err)

	// Wait for job to be processed
	time.Sleep(100 * time.Millisecond)

	assert.True(suite.T(), executed)
	assert.True(suite.T(), job.IsExecuted())
}

func (suite *WorkerPoolTestSuite) TestJobRetry() {
	suite.pool.Start()

	executionCount := 0
	job := NewTestJob("test-1", "test", 2, func(ctx context.Context) error {
		executionCount++
		if executionCount <= 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	err := suite.pool.Submit(job)
	assert.NoError(suite.T(), err)

	// Wait for job to be processed with retries
	time.Sleep(3 * time.Second)

	// Job should be executed at least twice (initial + 1 retry minimum)
	assert.GreaterOrEqual(suite.T(), executionCount, 2)
	assert.True(suite.T(), job.IsExecuted())
}

func (suite *WorkerPoolTestSuite) TestJobMaxRetriesExceeded() {
	suite.pool.Start()

	executionCount := 0
	job := NewTestJob("test-1", "test", 1, func(ctx context.Context) error {
		executionCount++
		return errors.New("persistent error")
	})

	err := suite.pool.Submit(job)
	assert.NoError(suite.T(), err)

	// Wait for job to be processed with retries
	time.Sleep(2 * time.Second)

	assert.Equal(suite.T(), 2, executionCount) // Initial + 1 retry
	assert.True(suite.T(), job.IsExecuted())
	assert.Equal(suite.T(), 1, job.GetRetryCount())
}

func (suite *WorkerPoolTestSuite) TestMultipleJobs() {
	suite.pool.Start()

	jobCount := 5
	executedJobs := make([]bool, jobCount)
	var mu sync.Mutex

	for i := 0; i < jobCount; i++ {
		index := i
		job := NewTestJob(
			fmt.Sprintf("test-%d", i),
			"test",
			3,
			func(ctx context.Context) error {
				mu.Lock()
				executedJobs[index] = true
				mu.Unlock()
				return nil
			},
		)

		err := suite.pool.Submit(job)
		assert.NoError(suite.T(), err)
	}

	// Wait for all jobs to be processed
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	for i, executed := range executedJobs {
		assert.True(suite.T(), executed, "Job %d was not executed", i)
	}
	mu.Unlock()
}

func (suite *WorkerPoolTestSuite) TestQueueFull() {
	// Create a small queue
	pool := NewWorkerPool(1, 2, suite.logger)
	pool.Start()
	defer pool.Stop()

	// Fill the queue
	for i := 0; i < 2; i++ {
		job := NewTestJob(fmt.Sprintf("test-%d", i), "test", 3, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond) // Slow job to fill queue
			return nil
		})
		err := pool.Submit(job)
		assert.NoError(suite.T(), err)
	}

	// Try to submit one more job - should fail
	job := NewTestJob("test-overflow", "test", 3, nil)
	err := pool.Submit(job)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "job queue is full")
}

func (suite *WorkerPoolTestSuite) TestGracefulShutdown() {
	suite.pool.Start()

	// Submit a quick job
	jobExecuted := false
	job := NewTestJob("quick-job", "test", 3, func(ctx context.Context) error {
		jobExecuted = true
		return nil
	})

	err := suite.pool.Submit(job)
	assert.NoError(suite.T(), err)

	// Give job time to execute
	time.Sleep(50 * time.Millisecond)

	// Stop the pool
	suite.pool.Stop()

	// Job should have been executed
	assert.True(suite.T(), jobExecuted)
	assert.False(suite.T(), suite.pool.IsStarted())
}

func TestWorkerPoolTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerPoolTestSuite))
}
