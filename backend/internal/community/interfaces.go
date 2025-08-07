package community

import "bookmark-sync-service/backend/pkg/worker"

// WorkerPoolInterface defines the interface for worker pool operations
type WorkerPoolInterface interface {
	Submit(job worker.Job) error
}
