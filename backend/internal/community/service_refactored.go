package community

import (
	"context"

	"bookmark-sync-service/backend/pkg/worker"

	"go.uber.org/zap"
)

// RefactoredService is the main service that delegates to domain-focused services
type RefactoredService struct {
	socialMetrics    *SocialMetricsService
	trending         *TrendingService
	recommendations  *RecommendationService
	userRelationship *UserRelationshipService
	behaviorTracking *BehaviorTrackingService
	userFeed         *UserFeedService
	jsonHelper       *JSONHelper
	cacheHelper      *CacheHelper
	logger           *zap.Logger
}

// NewRefactoredService creates a new refactored service with domain-focused services
func NewRefactoredService(
	db Database,
	redis RedisClient,
	workerPool *worker.WorkerPool,
	logger *zap.Logger,
) *RefactoredService {
	// Create shared helpers
	jsonHelper := NewJSONHelper()
	cacheHelper := NewCacheHelper(redis, jsonHelper)

	// Create domain-focused services
	socialMetrics := NewSocialMetricsService(db, redis, jsonHelper, logger)
	trending := NewTrendingService(db, redis, jsonHelper, logger)
	recommendations := NewRecommendationService(db, redis, jsonHelper, logger)
	userRelationship := NewUserRelationshipService(db, redis, cacheHelper, logger)
	behaviorTracking := NewBehaviorTrackingService(db, redis, workerPool, socialMetrics, trending, logger)
	userFeed := NewUserFeedService(db, redis, jsonHelper, logger)

	return &RefactoredService{
		socialMetrics:    socialMetrics,
		trending:         trending,
		recommendations:  recommendations,
		userRelationship: userRelationship,
		behaviorTracking: behaviorTracking,
		userFeed:         userFeed,
		jsonHelper:       jsonHelper,
		cacheHelper:      cacheHelper,
		logger:           logger,
	}
}

// TrackUserBehavior delegates to BehaviorTrackingService
func (s *RefactoredService) TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error {
	return s.behaviorTracking.TrackUserBehavior(ctx, req)
}

// FollowUser delegates to UserRelationshipService
func (s *RefactoredService) FollowUser(ctx context.Context, followerID string, req *FollowRequest) error {
	return s.userRelationship.FollowUser(ctx, followerID, req)
}

// UnfollowUser delegates to UserRelationshipService
func (s *RefactoredService) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	return s.userRelationship.UnfollowUser(ctx, followerID, followingID)
}

// GetRecommendations delegates to RecommendationService
func (s *RefactoredService) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error) {
	return s.recommendations.GetRecommendations(ctx, req)
}

// GetTrendingBookmarksInternal delegates to TrendingService
func (s *RefactoredService) GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error) {
	return s.trending.GetTrendingBookmarksInternal(ctx, req)
}

// GenerateUserFeed delegates to UserFeedService
func (s *RefactoredService) GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error) {
	return s.userFeed.GenerateUserFeed(ctx, req)
}

// GetSocialMetrics delegates to SocialMetricsService
func (s *RefactoredService) GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error) {
	return s.socialMetrics.GetSocialMetrics(ctx, bookmarkID)
}

// UpdateSocialMetrics delegates to SocialMetricsService
func (s *RefactoredService) UpdateSocialMetrics(ctx context.Context, bookmarkID uint, actionType string) error {
	return s.socialMetrics.UpdateSocialMetrics(ctx, bookmarkID, actionType)
}

// GetUserStats delegates to UserRelationshipService
func (s *RefactoredService) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	return s.userRelationship.GetUserStats(ctx, userID)
}

// CalculateTrendingScores delegates to TrendingService
func (s *RefactoredService) CalculateTrendingScores(ctx context.Context, timeWindow string) error {
	return s.trending.CalculateTrendingScores(ctx, timeWindow)
}

// GenerateRecommendations delegates to RecommendationService
func (s *RefactoredService) GenerateRecommendations(ctx context.Context, userID, algorithm string) error {
	return s.recommendations.GenerateRecommendations(ctx, userID, algorithm)
}

// UpdateTrendingCache delegates to TrendingService
func (s *RefactoredService) UpdateTrendingCache(ctx context.Context, bookmarkID uint, actionType string) error {
	return s.trending.UpdateTrendingCache(ctx, bookmarkID, actionType)
}
