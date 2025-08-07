package sync

import (
	"bookmark-sync-service/backend/pkg/websocket"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RedisClient interface for sync operations
type RedisClient interface {
	PublishSyncEvent(ctx context.Context, userID string, event interface{}) error
	SubscribeToSyncEvents(ctx context.Context, userID string) interface{}
}

// Service handles synchronization operations
type Service struct {
	db          *gorm.DB
	redisClient RedisClient
	logger      *zap.Logger
}

// SyncEventType represents the type of sync event
type SyncEventType string

const (
	SyncEventBookmarkCreated   SyncEventType = "bookmark_created"
	SyncEventBookmarkUpdated   SyncEventType = "bookmark_updated"
	SyncEventBookmarkDeleted   SyncEventType = "bookmark_deleted"
	SyncEventCollectionCreated SyncEventType = "collection_created"
	SyncEventCollectionUpdated SyncEventType = "collection_updated"
	SyncEventCollectionDeleted SyncEventType = "collection_deleted"
)

// SyncStatus represents the status of a sync event
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusFailed  SyncStatus = "failed"
)

// SyncMessageType represents the type of sync message
type SyncMessageType string

const (
	SyncMessagePing         SyncMessageType = "ping"
	SyncMessagePong         SyncMessageType = "pong"
	SyncMessageSyncRequest  SyncMessageType = "sync_request"
	SyncMessageSyncResponse SyncMessageType = "sync_response"
	SyncMessageEvent        SyncMessageType = "event"
)

// SyncEvent represents a synchronization event
type SyncEvent struct {
	ID         uint          `json:"id" gorm:"primaryKey"`
	Type       SyncEventType `json:"type" gorm:"not null"`
	UserID     string        `json:"user_id" gorm:"not null;index"`
	ResourceID string        `json:"resource_id" gorm:"not null;index"`
	Action     string        `json:"action" gorm:"not null"`
	Data       string        `json:"data" gorm:"type:jsonb"`
	DeviceID   string        `json:"device_id" gorm:"not null;index"`
	Status     SyncStatus    `json:"status" gorm:"default:'pending'"`
	Timestamp  time.Time     `json:"timestamp" gorm:"not null;index"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// SyncState represents the synchronization state for a device
type SyncState struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"not null;index"`
	DeviceID     string    `json:"device_id" gorm:"not null;index"`
	LastSyncTime time.Time `json:"last_sync_time" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DeltaSync represents a delta synchronization response
type DeltaSync struct {
	Events    []*SyncEvent `json:"events"`
	Timestamp time.Time    `json:"timestamp"`
}

// NewService creates a new sync service
func NewService(db *gorm.DB, redisClient RedisClient, logger *zap.Logger) *Service {
	return &Service{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateSyncEvent creates and publishes a sync event
func (s *Service) CreateSyncEvent(ctx context.Context, event *SyncEvent) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set default status
	if event.Status == "" {
		event.Status = SyncStatusPending
	}

	// Store event in database
	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		s.logger.Error("Failed to create sync event", zap.Error(err))
		return fmt.Errorf("failed to create sync event: %w", err)
	}

	// Publish event to Redis for real-time sync
	if err := s.redisClient.PublishSyncEvent(ctx, event.UserID, event); err != nil {
		s.logger.Error("Failed to publish sync event", zap.Error(err))
		// Don't return error here as the event is already stored
	}

	s.logger.Info("Sync event created",
		zap.String("type", string(event.Type)),
		zap.String("user_id", event.UserID),
		zap.String("resource_id", event.ResourceID),
		zap.String("device_id", event.DeviceID),
	)

	return nil
}

// ResolveConflict resolves conflicts between sync events using timestamp-based resolution
func (s *Service) ResolveConflict(events []*SyncEvent) *SyncEvent {
	if len(events) == 0 {
		return nil
	}

	if len(events) == 1 {
		return events[0]
	}

	// Sort events by timestamp (newest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Return the newest event (timestamp-based conflict resolution)
	winner := events[0]

	s.logger.Info("Conflict resolved using timestamp",
		zap.String("winner_device", winner.DeviceID),
		zap.Time("winner_timestamp", winner.Timestamp),
		zap.Int("total_conflicts", len(events)),
	)

	return winner
}

// GetSyncState retrieves the sync state for a device
func (s *Service) GetSyncState(ctx context.Context, userID, deviceID string) (*SyncState, error) {
	var state SyncState

	err := s.db.WithContext(ctx).Where("user_id = ? AND device_id = ?", userID, deviceID).First(&state).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sync state
			state = SyncState{
				UserID:       userID,
				DeviceID:     deviceID,
				LastSyncTime: time.Now(),
			}

			if err := s.db.WithContext(ctx).Create(&state).Error; err != nil {
				return nil, fmt.Errorf("failed to create sync state: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get sync state: %w", err)
		}
	}

	return &state, nil
}

// UpdateSyncState updates the last sync time for a device
func (s *Service) UpdateSyncState(ctx context.Context, userID, deviceID string, lastSyncTime time.Time) error {
	result := s.db.WithContext(ctx).Model(&SyncState{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Update("last_sync_time", lastSyncTime)

	if result.Error != nil {
		return fmt.Errorf("failed to update sync state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		// Create new sync state if it doesn't exist
		state := SyncState{
			UserID:       userID,
			DeviceID:     deviceID,
			LastSyncTime: lastSyncTime,
		}

		if err := s.db.WithContext(ctx).Create(&state).Error; err != nil {
			return fmt.Errorf("failed to create sync state: %w", err)
		}
	}

	return nil
}

// GetDeltaSync retrieves events that occurred after the last sync time
func (s *Service) GetDeltaSync(ctx context.Context, userID, deviceID string, lastSyncTime time.Time) (*DeltaSync, error) {
	var events []*SyncEvent

	err := s.db.WithContext(ctx).
		Where("user_id = ? AND device_id != ? AND timestamp > ?", userID, deviceID, lastSyncTime).
		Order("timestamp ASC").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get delta sync: %w", err)
	}

	// Optimize events to reduce bandwidth
	optimizedEvents := s.OptimizeEvents(events)

	return &DeltaSync{
		Events:    optimizedEvents,
		Timestamp: time.Now(),
	}, nil
}

// QueueOfflineEvent queues an event for offline processing
func (s *Service) QueueOfflineEvent(ctx context.Context, event *SyncEvent) error {
	event.Status = SyncStatusPending
	event.Timestamp = time.Now()

	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to queue offline event: %w", err)
	}

	s.logger.Info("Event queued for offline processing",
		zap.String("type", string(event.Type)),
		zap.String("user_id", event.UserID),
		zap.String("resource_id", event.ResourceID),
		zap.String("device_id", event.DeviceID),
	)

	return nil
}

// GetOfflineQueue retrieves pending events for a device
func (s *Service) GetOfflineQueue(ctx context.Context, userID, deviceID string) ([]*SyncEvent, error) {
	var events []*SyncEvent

	err := s.db.WithContext(ctx).
		Where("user_id = ? AND device_id = ? AND status = ?", userID, deviceID, SyncStatusPending).
		Order("timestamp ASC").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get offline queue: %w", err)
	}

	return events, nil
}

// ProcessOfflineQueue processes pending events when connectivity is restored
func (s *Service) ProcessOfflineQueue(ctx context.Context, userID, deviceID string) error {
	events, err := s.GetOfflineQueue(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	for _, event := range events {
		// Publish event to Redis
		if err := s.redisClient.PublishSyncEvent(ctx, userID, event); err != nil {
			s.logger.Error("Failed to publish offline event", zap.Error(err))
			continue
		}

		// Mark event as synced
		event.Status = SyncStatusSynced
		if err := s.db.WithContext(ctx).Save(event).Error; err != nil {
			s.logger.Error("Failed to update event status", zap.Error(err))
			continue
		}
	}

	s.logger.Info("Processed offline queue",
		zap.String("user_id", userID),
		zap.String("device_id", deviceID),
		zap.Int("events_processed", len(events)),
	)

	return nil
}

// HandleSyncMessage handles incoming WebSocket sync messages
func (s *Service) HandleSyncMessage(ctx context.Context, msg *websocket.SyncMessage) (*websocket.SyncMessage, error) {
	switch msg.Type {
	case "ping":
		return &websocket.SyncMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}, nil

	case "sync_request":
		// Extract last sync time from message data
		var lastSyncTime time.Time
		if lastSyncUnix, ok := msg.Data["last_sync_time"].(float64); ok {
			lastSyncTime = time.Unix(int64(lastSyncUnix), 0)
		} else {
			lastSyncTime = time.Now().Add(-24 * time.Hour) // Default to 24 hours ago
		}

		// Get delta sync
		delta, err := s.GetDeltaSync(ctx, msg.UserID, msg.DeviceID, lastSyncTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get delta sync: %w", err)
		}

		// Update sync state
		if err := s.UpdateSyncState(ctx, msg.UserID, msg.DeviceID, time.Now()); err != nil {
			s.logger.Error("Failed to update sync state", zap.Error(err))
		}

		// Convert delta to response data
		deltaData, err := json.Marshal(delta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal delta sync: %w", err)
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(deltaData, &responseData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal delta sync: %w", err)
		}

		return &websocket.SyncMessage{
			Type:      "sync_response",
			Data:      responseData,
			Timestamp: time.Now(),
		}, nil

	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// OptimizeEvents optimizes events for bandwidth efficiency
func (s *Service) OptimizeEvents(events []*SyncEvent) []*SyncEvent {
	if len(events) <= 1 {
		return events
	}

	// Group events by resource ID
	resourceEvents := make(map[string][]*SyncEvent)
	for _, event := range events {
		resourceEvents[event.ResourceID] = append(resourceEvents[event.ResourceID], event)
	}

	var optimized []*SyncEvent

	// For each resource, keep only the latest event
	for _, resourceEventList := range resourceEvents {
		if len(resourceEventList) == 1 {
			optimized = append(optimized, resourceEventList[0])
			continue
		}

		// Sort by timestamp and keep the latest
		sort.Slice(resourceEventList, func(i, j int) bool {
			return resourceEventList[i].Timestamp.After(resourceEventList[j].Timestamp)
		})

		optimized = append(optimized, resourceEventList[0])
	}

	// Sort final result by timestamp
	sort.Slice(optimized, func(i, j int) bool {
		return optimized[i].Timestamp.Before(optimized[j].Timestamp)
	})

	s.logger.Debug("Events optimized for bandwidth",
		zap.Int("original_count", len(events)),
		zap.Int("optimized_count", len(optimized)),
	)

	return optimized
}
