package community

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRelationshipService handles user following/unfollowing
type UserRelationshipService struct {
	db          Database
	redis       RedisClient
	cacheHelper *CacheHelper
	logger      *zap.Logger
}

// NewUserRelationshipService creates a new user relationship service
func NewUserRelationshipService(db Database, redis RedisClient, cacheHelper *CacheHelper, logger *zap.Logger) *UserRelationshipService {
	return &UserRelationshipService{
		db:          db,
		redis:       redis,
		cacheHelper: cacheHelper,
		logger:      logger,
	}
}

// FollowUser creates a following relationship
func (s *UserRelationshipService) FollowUser(ctx context.Context, followerID string, req *FollowRequest) error {
	if followerID == "" {
		return ErrInvalidFollowerID
	}
	if req.FollowingID == "" {
		return ErrInvalidFollowingID
	}
	if followerID == req.FollowingID {
		return ErrCannotFollowSelf
	}

	// Check if already following
	if exists, err := s.checkFollowExists(followerID, req.FollowingID); err != nil {
		return err
	} else if exists {
		return ErrAlreadyFollowing
	}

	// Create follow relationship
	follow := &UserFollow{
		FollowerID:  followerID,
		FollowingID: req.FollowingID,
		Status:      "active",
	}

	if err := follow.Validate(); err != nil {
		return err
	}

	if err := s.db.Create(follow).Error; err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	// Clear cache for both users
	s.clearUserStatsCache(ctx, followerID, req.FollowingID)

	return nil
}

// UnfollowUser removes a following relationship
func (s *UserRelationshipService) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	if followerID == "" {
		return ErrInvalidFollowerID
	}
	if followingID == "" {
		return ErrInvalidFollowingID
	}

	// Find existing follow relationship
	var follow UserFollow
	err := s.db.First(&follow, "follower_id = ? AND following_id = ?", followerID, followingID).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFollowing
	}
	if err != nil {
		return fmt.Errorf("failed to find follow relationship: %w", err)
	}

	// Delete follow relationship
	if err := s.db.Delete(&follow).Error; err != nil {
		return fmt.Errorf("failed to delete follow relationship: %w", err)
	}

	// Clear cache for both users
	s.clearUserStatsCache(ctx, followerID, followingID)

	return nil
}

// GetUserStats returns user statistics and influence metrics
func (s *UserRelationshipService) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_stats:%s", userID)
	var stats UserStatsResponse

	err := s.cacheHelper.GetOrSet(ctx, cacheKey, &stats, 30*time.Minute, func() (any, error) {
		return s.calculateUserStats(userID)
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// checkFollowExists checks if a follow relationship already exists
func (s *UserRelationshipService) checkFollowExists(followerID, followingID string) (bool, error) {
	var existing UserFollow
	err := s.db.First(&existing, "follower_id = ? AND following_id = ?", followerID, followingID).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check existing follow: %w", err)
	}
	return true, nil
}

// calculateUserStats calculates user statistics
func (s *UserRelationshipService) calculateUserStats(userID string) (*UserStatsResponse, error) {
	stats := &UserStatsResponse{
		UserID: userID,
	}

	// Get followers count
	followersCount, err := s.getFollowersCount(userID)
	if err != nil {
		return nil, err
	}
	stats.FollowersCount = followersCount

	// Get following count
	followingCount, err := s.getFollowingCount(userID)
	if err != nil {
		return nil, err
	}
	stats.FollowingCount = followingCount

	// Calculate influence score (simplified)
	stats.InfluenceScore = float64(stats.FollowersCount)*0.7 + float64(stats.TotalEngagement)*0.3

	return stats, nil
}

// getFollowersCount gets the number of followers for a user
func (s *UserRelationshipService) getFollowersCount(userID string) (int, error) {
	var followers []UserFollow
	err := s.db.Where("following_id = ? AND status = ?", userID, "active").Find(&followers).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get followers count: %w", err)
	}
	return len(followers), nil
}

// getFollowingCount gets the number of users being followed
func (s *UserRelationshipService) getFollowingCount(userID string) (int, error) {
	var following []UserFollow
	err := s.db.Where("follower_id = ? AND status = ?", userID, "active").Find(&following).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get following count: %w", err)
	}
	return len(following), nil
}

// clearUserStatsCache clears cache for multiple users
func (s *UserRelationshipService) clearUserStatsCache(ctx context.Context, userIDs ...string) {
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = fmt.Sprintf("user_stats:%s", userID)
	}

	if err := s.cacheHelper.Delete(ctx, keys...); err != nil {
		s.logger.Warn("Failed to clear user stats cache", zap.Error(err), zap.Strings("user_ids", userIDs))
	}
}
