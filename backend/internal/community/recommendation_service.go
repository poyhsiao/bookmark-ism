package community

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// RecommendationService handles recommendation generation and retrieval
type RecommendationService struct {
	db         Database
	redis      RedisClient
	jsonHelper *JSONHelper
	logger     *zap.Logger
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(db Database, redis RedisClient, jsonHelper *JSONHelper, logger *zap.Logger) *RecommendationService {
	return &RecommendationService{
		db:         db,
		redis:      redis,
		jsonHelper: jsonHelper,
		logger:     logger,
	}
}

// GetRecommendations returns personalized bookmark recommendations
func (s *RecommendationService) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}

	validationHelper := NewValidationHelper()
	req.Limit = validationHelper.ValidateLimit(req.Limit)

	// Check cache first
	cacheKey := fmt.Sprintf("recommendations:%s:%s:%s", req.UserID, req.Algorithm, req.Context)
	var recommendations []RecommendationResponse

	cacheHelper := NewCacheHelper(s.redis, s.jsonHelper)
	err := cacheHelper.Get(ctx, cacheKey, &recommendations)
	if err == nil {
		return recommendations, nil
	}

	// Get recommendations from database
	dbRecommendations, err := s.getRecommendationsFromDB(req)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	recommendations = s.convertToResponseFormat(dbRecommendations)

	// Cache results
	cacheHelper.Set(ctx, cacheKey, recommendations, 15*time.Minute)

	return recommendations, nil
}

// GenerateRecommendations generates bookmark recommendations for a user
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID, algorithm string) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	configHelper := NewConfigHelper()
	if !configHelper.ValidateAlgorithm(algorithm) {
		return ErrInvalidAlgorithm
	}

	// Get user's behavior history
	behaviors, err := s.getUserBehaviors(userID)
	if err != nil {
		return err
	}

	// Check for insufficient data
	if len(behaviors) == 0 {
		return ErrInsufficientData
	}

	// Generate recommendations based on algorithm
	recommendations := s.generateRecommendationsByAlgorithm(userID, algorithm, behaviors)

	// Save recommendations
	return s.saveRecommendations(recommendations, userID)
}

// getRecommendationsFromDB retrieves recommendations from database
func (s *RecommendationService) getRecommendationsFromDB(req *RecommendationRequest) ([]BookmarkRecommendation, error) {
	var dbRecommendations []BookmarkRecommendation
	query := s.db.Where("user_id = ? AND expires_at > ?", req.UserID, time.Now())

	if req.Algorithm != "" {
		query = query.Where("reason_type = ?", req.Algorithm)
	}

	err := query.Order("score DESC").Limit(req.Limit).Find(&dbRecommendations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	return dbRecommendations, nil
}

// convertToResponseFormat converts database recommendations to response format
func (s *RecommendationService) convertToResponseFormat(dbRecommendations []BookmarkRecommendation) []RecommendationResponse {
	recommendations := make([]RecommendationResponse, len(dbRecommendations))
	for i, rec := range dbRecommendations {
		recommendations[i] = RecommendationResponse{
			BookmarkID: rec.BookmarkID,
			Score:      rec.Score,
			ReasonType: rec.ReasonType,
			ReasonText: s.generateReasonText(rec.ReasonType),
		}
	}
	return recommendations
}

// getUserBehaviors gets user behavior history
func (s *RecommendationService) getUserBehaviors(userID string) ([]UserBehavior, error) {
	var behaviors []UserBehavior
	err := s.db.Where("user_id = ?", userID).Find(&behaviors).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user behaviors: %w", err)
	}
	return behaviors, nil
}

// generateRecommendationsByAlgorithm generates recommendations based on algorithm
func (s *RecommendationService) generateRecommendationsByAlgorithm(userID, algorithm string, behaviors []UserBehavior) []BookmarkRecommendation {
	switch algorithm {
	case "collaborative":
		return s.generateCollaborativeRecommendations(userID, behaviors)
	case "content_based":
		return s.generateContentBasedRecommendations(userID, behaviors)
	case "trending":
		return s.generateTrendingRecommendations(userID)
	case "popularity":
		return s.generatePopularityRecommendations(userID)
	default:
		return s.generateHybridRecommendations(userID, behaviors)
	}
}

// saveRecommendations saves generated recommendations
func (s *RecommendationService) saveRecommendations(recommendations []BookmarkRecommendation, userID string) error {
	for _, rec := range recommendations {
		rec.UserID = userID
		expiresAt := time.Now().Add(24 * time.Hour)
		rec.ExpiresAt = &expiresAt

		if err := rec.Validate(); err == nil {
			if err := s.db.Create(&rec).Error; err != nil {
				s.logger.Error("Failed to save recommendation", zap.Error(err), zap.String("user_id", userID))
			}
		}
	}
	return nil
}

// generateCollaborativeRecommendations generates collaborative filtering recommendations
func (s *RecommendationService) generateCollaborativeRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	var recommendations []BookmarkRecommendation

	// Simplified collaborative filtering
	for _, behavior := range behaviors {
		if len(recommendations) >= 10 {
			break
		}

		rec := BookmarkRecommendation{
			BookmarkID: behavior.BookmarkID,
			Score:      0.7,
			ReasonType: "collaborative",
			ReasonData: `{"similar_users": ["user1", "user2"]}`,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateContentBasedRecommendations generates content-based recommendations
func (s *RecommendationService) generateContentBasedRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	var recommendations []BookmarkRecommendation

	for _, behavior := range behaviors {
		if len(recommendations) >= 10 {
			break
		}

		rec := BookmarkRecommendation{
			BookmarkID: behavior.BookmarkID,
			Score:      0.6,
			ReasonType: "content_based",
			ReasonData: `{"similar_tags": ["tech", "programming"]}`,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateTrendingRecommendations generates trending-based recommendations
func (s *RecommendationService) generateTrendingRecommendations(userID string) []BookmarkRecommendation {
	var trending []TrendingBookmark
	s.db.Where("time_window = ?", "daily").Order("trending_score DESC").Limit(10).Find(&trending)

	var recommendations []BookmarkRecommendation
	for _, t := range trending {
		rec := BookmarkRecommendation{
			BookmarkID: t.BookmarkID,
			Score:      math.Min(t.TrendingScore/100.0, 1.0),
			ReasonType: "trending",
			ReasonData: fmt.Sprintf(`{"trending_score": %.2f}`, t.TrendingScore),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generatePopularityRecommendations generates popularity-based recommendations
func (s *RecommendationService) generatePopularityRecommendations(userID string) []BookmarkRecommendation {
	var metrics []SocialMetrics
	s.db.Order("total_views DESC").Limit(10).Find(&metrics)

	var recommendations []BookmarkRecommendation
	for _, m := range metrics {
		score := float64(m.TotalViews) / 1000.0
		if score > 1.0 {
			score = 1.0
		}

		rec := BookmarkRecommendation{
			BookmarkID: m.BookmarkID,
			Score:      score,
			ReasonType: "popularity",
			ReasonData: fmt.Sprintf(`{"total_views": %d}`, m.TotalViews),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateHybridRecommendations generates hybrid recommendations
func (s *RecommendationService) generateHybridRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	collaborative := s.generateCollaborativeRecommendations(userID, behaviors)
	trending := s.generateTrendingRecommendations(userID)

	var recommendations []BookmarkRecommendation

	// Add collaborative with higher weight
	for i, rec := range collaborative {
		if i >= 5 {
			break
		}
		rec.Score *= 0.8
		rec.ReasonType = "hybrid"
		recommendations = append(recommendations, rec)
	}

	// Add trending with lower weight
	for i, rec := range trending {
		if i >= 5 {
			break
		}
		rec.Score *= 0.6
		rec.ReasonType = "hybrid"
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateReasonText generates human-readable reason text
func (s *RecommendationService) generateReasonText(reasonType string) string {
	switch reasonType {
	case "collaborative":
		return "Users with similar interests also liked this"
	case "content_based":
		return "Similar to bookmarks you've saved"
	case "trending":
		return "Trending in your network"
	case "popularity":
		return "Popular among all users"
	case "hybrid":
		return "Recommended based on your activity and trends"
	default:
		return "Recommended for you"
	}
}
