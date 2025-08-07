package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"go.uber.org/zap"
)

// Job represents a unit of work to be processed
type Job interface {
	Execute(ctx context.Context) error
	GetID() string
	GetType() string
	GetRetryCount() int
	IncrementRetryCount()
	GetMaxRetries() int
}

// BaseJob provides common job functionality
type BaseJob struct {
	ID         string
	Type       string
	RetryCount int
	MaxRetries int
	CreatedAt  time.Time
}

func (j *BaseJob) GetID() string {
	return j.ID
}

func (j *BaseJob) GetType() string {
	return j.Type
}

func (j *BaseJob) GetRetryCount() int {
	return j.RetryCount
}

func (j *BaseJob) IncrementRetryCount() {
	j.RetryCount++
}

func (j *BaseJob) GetMaxRetries() int {
	return j.MaxRetries
}

// WorkerPool manages a pool of workers to process jobs
type WorkerPool struct {
	workers  int
	jobQueue chan Job
	quit     chan bool
	wg       sync.WaitGroup
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	started  bool
	mu       sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int, logger *zap.Logger) *WorkerPool {
	if workers <= 0 {
		workers = config.DefaultWorkerPoolSize
	}
	if queueSize <= 0 {
		queueSize = config.DefaultQueueSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:  workers,
		jobQueue: make(chan Job, queueSize),
		quit:     make(chan bool),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.started {
		wp.logger.Warn("Worker pool already started")
		return
	}

	wp.logger.Info("Starting worker pool", zap.Int("workers", wp.workers))

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	wp.started = true
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.started {
		wp.logger.Warn("Worker pool not started")
		return
	}

	wp.logger.Info("Stopping worker pool")

	// Close job queue to prevent new jobs
	close(wp.jobQueue)

	// Cancel context to signal workers to stop
	wp.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("Worker pool stopped gracefully")
	case <-time.After(config.WorkerShutdownTimeout):
		wp.logger.Warn("Worker pool shutdown timeout exceeded")
	}

	wp.started = false
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job Job) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.started {
		return fmt.Errorf("worker pool not started")
	}

	select {
	case wp.jobQueue <- job:
		wp.logger.Debug("Job submitted",
			zap.String("job_id", job.GetID()),
			zap.String("job_type", job.GetType()))
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				wp.logger.Debug("Worker stopping - job queue closed", zap.Int("worker_id", id))
				return
			}

			wp.processJob(id, job)

		case <-wp.ctx.Done():
			wp.logger.Debug("Worker stopping - context cancelled", zap.Int("worker_id", id))
			return
		}
	}
}

// processJob processes a single job with retry logic
func (wp *WorkerPool) processJob(workerID int, job Job) {
	logger := wp.logger.With(
		zap.Int("worker_id", workerID),
		zap.String("job_id", job.GetID()),
		zap.String("job_type", job.GetType()),
		zap.Int("retry_count", job.GetRetryCount()),
	)

	logger.Debug("Processing job")

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(wp.ctx, 30*time.Second)
	defer cancel()

	err := job.Execute(jobCtx)
	if err != nil {
		logger.Error("Job execution failed", zap.Error(err))

		// Retry logic
		if job.GetRetryCount() < job.GetMaxRetries() {
			job.IncrementRetryCount()
			logger.Info("Retrying job", zap.Int("retry_count", job.GetRetryCount()))

			// Exponential backoff
			backoff := time.Duration(job.GetRetryCount()) * time.Second
			time.Sleep(backoff)

			// Try to resubmit job with timeout to avoid blocking
			retryCtx, retryCancel := context.WithTimeout(wp.ctx, time.Second)
			defer retryCancel()

			select {
			case wp.jobQueue <- job:
				logger.Debug("Job resubmitted for retry")
			case <-retryCtx.Done():
				logger.Debug("Cannot retry job - worker pool shutting down or timeout")
				return
			}
		} else {
			logger.Error("Job failed after max retries",
				zap.Int("max_retries", job.GetMaxRetries()))
		}
	} else {
		logger.Debug("Job completed successfully")
	}
}

// GetQueueSize returns the current number of jobs in the queue
func (wp *WorkerPool) GetQueueSize() int {
	return len(wp.jobQueue)
}

// IsStarted returns whether the worker pool is started
func (wp *WorkerPool) IsStarted() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.started
}
