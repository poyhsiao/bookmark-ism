package config

import "time"

// Constants for magic numbers and default values
const (
	// Cache TTL constants
	DefaultCacheTTL  = 24 * time.Hour
	OfflineQueueTTL  = 7 * 24 * time.Hour
	OfflineStatusTTL = time.Hour
	BookmarkCacheTTL = 24 * time.Hour

	// Connection timeouts
	DefaultConnectionTimeout = 5 * time.Second
	RedisConnectionTimeout   = 5 * time.Second
	TypesenseTimeout         = 5 * time.Second
	ConnectivityCheckTimeout = 5 * time.Second

	// WebSocket constants
	WebSocketWriteWait      = 10 * time.Second
	WebSocketPongWait       = 60 * time.Second
	WebSocketPingPeriod     = (WebSocketPongWait * 9) / 10
	WebSocketMaxMessageSize = 512

	// Storage constants
	DefaultThumbnailSize = 200
	DefaultMemoryLimit   = 32 << 20 // 32 MB

	// Pagination and limits
	DefaultPageSize    = 20
	MaxPageSize        = 100
	DefaultSearchLimit = 50

	// Social metrics defaults
	DefaultSaveCount    = 0
	DefaultLikeCount    = 0
	DefaultCommentCount = 0
	DefaultViewCount    = 0

	// String field sizes
	MaxTitleLength       = 255
	MaxDescriptionLength = 500
	MaxIPAddressLength   = 45
	MaxUserAgentLength   = 500
	MaxForkReasonLength  = 500

	// Worker queue settings
	DefaultWorkerPoolSize = 10
	DefaultQueueSize      = 1000
	WorkerShutdownTimeout = 30 * time.Second
)

// Redis key prefixes
const (
	OfflineBookmarkPrefix = "offline:bookmark"
	OfflineQueuePrefix    = "offline:queue"
	OfflineStatusPrefix   = "offline:status"
	OfflineStatsPrefix    = "offline:stats"
	CacheStatsPrefix      = "cache:stats"
)

// Error messages
const (
	ErrUserNotAuthenticated = "User not authenticated"
	ErrInvalidUserID        = "Invalid user ID"
	ErrInvalidBookmarkID    = "Invalid bookmark ID"
	ErrUserNotFound         = "User not found"
	ErrBookmarkNotFound     = "Bookmark not found"
	ErrInvalidData          = "Invalid data"
	ErrCacheError           = "Cache error"
	ErrInternalError        = "Internal server error"
)

// HTTP status codes for consistent error handling
const (
	StatusUnauthorized    = 401
	StatusBadRequest      = 400
	StatusNotFound        = 404
	StatusInternalError   = 500
	StatusValidationError = 422
)
