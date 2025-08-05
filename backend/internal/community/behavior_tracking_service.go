package community

import (
	"context"
	"fmt"

	"bookmark-sync-service/backend/pkg/worker"

	"go.uber.org/zap"
)

// BehaviorTrackingService handles user behavior tracking
type BehaviorTrackingService struct {
	db            Database
	redis         RedisClient
	workerPool    *worker.WorkerPool
	socialMetrics *SocialMetricsService
	trending      *TrendingService
	logger        *zap.Logger
}

// NewBehaviorTrackingService creates a new behavior tracking service
func NewBehaviorTrackingService(
	db Database,
	redis RedisClient,
	workerPool *worker.WorkerPool,
	socialMetrics *SocialMetricsService,
	trending *TrendingService,
	logger *zap.Logger,
) *BehaviorTrackingService {
	return &BehaviorTrackingService{
		db:            db,
		redis:         redis,
		workerPool:    workerPool,
		socialMetrics: socialMetrics,
		trending:      trending,
		logger:        logger,
	}
}

// TrackUserBehavior records user interactions for recommendation engine
func (s *BehaviorTrackingService) TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return err
	}

	// Create and save behavior record
	behavior, err := s.createBehaviorRecord(req)
	if err != nil {
		return err
	}

	if err := s.saveBehaviorRecord(behavior); err != nil {
		return err
	}

	// Process async updates
	s.processAsyncUpdates(req)

	return nil
}

// validateRequest validates the behavior tracking request
func (s *BehaviorTrackingService) validateRequest(req *BehaviorTrackingRequest) error {
	validationHelper := NewValidationHelper()

	if err := validationHelper.ValidateUserID(req.UserID); err != nil {
		return err
	}
	if err := validationHelper.ValidateBookmarkID(req.BookmarkID); err != nil {
		return err
	}
	if err := validationHelper.ValidateActionType(req.ActionType); err != nil {
		return err
	}

	return nil
}

// createBehaviorRecord creates a behavior record from the request
func (s *BehaviorTrackingService) createBehaviorRecord(req *BehaviorTrackingRequest) (*UserBehavior, error) {
	behavior := &UserBehavior{
		UserID:     req.UserID,
		BookmarkID: req.BookmarkID,
		ActionType: req.ActionType,
		Duration:   req.Duration,
		Context:    req.Context,
	}

	// Serialize metadata if present
	if req.Metadata != nil {
		jsonHelper := NewJSONHelper()
		metadataJSON, err := jsonHelper.MarshalToString(req.Metadata)
		if err == nil {
			behavior.Metadata = metadataJSON
		}
	}

	// Validate behavior
	if err := behavior.Validate(); err != nil {
		return nil, err
	}

	return behavior, nil
}

// saveBehaviorRecord saves the behavior record to database
func (s *BehaviorTrackingService) saveBehaviorRecord(behavior *UserBehavior) error {
	if err := s.db.Create(behavior).Error; err != nil {
		return fmt.Errorf("failed to track user behavior: %w", err)
	}
	return nil
}

// processAsyncUpdates handles asynchronous updates via worker pool
func (s *BehaviorTrackingService) processAsyncUpdates(req *BehaviorTrackingRequest) {
	if s.workerPool == nil {
		return
	}

	// Update social metrics asynchronously
	socialJob := worker.NewSocialMetricsUpdateJob(
		req.BookmarkID,
		req.ActionType,
		s.socialMetrics,
		s.logger,
	)
	if err := s.workerPool.Submit(socialJob); err != nil {
		s.logger.Warn("Failed to submit social metrics update job", zap.Error(err))
	}

	// Update trending scores for significant actions
	if s.isSignificantAction(req.ActionType) {
		trendingJob := worker.NewTrendingCacheUpdateJob(
			req.BookmarkID,
			req.ActionType,
			s.trending,
			s.logger,
		)
		if err := s.workerPool.Submit(trendingJob); err != nil {
			s.logger.Warn("Failed to submit trending cache update job", zap.Error(err))
		}
	}
}

// isSignificantAction checks if an action is significant for trending calculations
func (s *BehaviorTrackingService) isSignificantAction(actionType string) bool {
	significantActions := map[string]bool{
		"view": true, "click": true, "save": true, "share": true, "like": true,
	}
	return significantActions[actionType]
}
