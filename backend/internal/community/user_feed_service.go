package community

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// UserFeedService handles user feed generation
type UserFeedService struct {
	db         Database
	redis      RedisClient
	jsonHelper *JSONHelper
	logger     *zap.Logger
}

// NewUserFeedService creates a new user feed service
func NewUserFeedService(db Database, redis RedisClient, jsonHelper *JSONHelper, logger *zap.Logger) *UserFeedService {
	return &UserFeedService{
		db:         db,
		redis:      redis,
		jsonHelper: jsonHelper,
		logger:     logger,
	}
}

// GenerateUserFeed creates personalized feed for user
func (s *UserFeedService) GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}

	validationHelper := NewValidationHelper()
	req.Limit = validationHelper.ValidateLimit(req.Limit)

	// Get feed items from database
	feedItems, err := s.getFeedItemsFromDB(req)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	feed := s.convertToFeedResponse(feedItems)

	return feed, nil
}

// getFeedItemsFromDB retrieves feed items from database
func (s *UserFeedService) getFeedItemsFromDB(req *FeedRequest) ([]UserFeed, error) {
	var feedItems []UserFeed
	query := s.db.Where("user_id = ?", req.UserID)

	if req.SourceType != "" && req.SourceType != "all" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	err := query.Order("score DESC, created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&feedItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user feed: %w", err)
	}

	return feedItems, nil
}

// convertToFeedResponse converts database feed items to response format
func (s *UserFeedService) convertToFeedResponse(feedItems []UserFeed) []UserFeedResponse {
	feed := make([]UserFeedResponse, len(feedItems))
	for i, item := range feedItems {
		feed[i] = UserFeedResponse{
			BookmarkID: item.BookmarkID,
			SourceType: item.SourceType,
			SourceID:   item.SourceID,
			Score:      item.Score,
			Position:   item.Position,
		}
	}
	return feed
}
