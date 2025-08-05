package community

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SocialMetricsService handles social engagement metrics
type SocialMetricsService struct {
	db         Database
	redis      RedisClient
	jsonHelper *JSONHelper
	logger     *zap.Logger
}

// NewSocialMetricsService creates a new social metrics service
func NewSocialMetricsService(db Database, redis RedisClient, jsonHelper *JSONHelper, logger *zap.Logger) *SocialMetricsService {
	return &SocialMetricsService{
		db:         db,
		redis:      redis,
		jsonHelper: jsonHelper,
		logger:     logger,
	}
}

// GetSocialMetrics returns social engagement metrics for a bookmark
func (s *SocialMetricsService) GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error) {
	if bookmarkID == 0 {
		return nil, ErrInvalidBookmarkID
	}

	var metrics SocialMetrics
	err := s.db.First(&metrics, "bookmark_id = ?", bookmarkID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrBookmarkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get social metrics: %w", err)
	}

	return &SocialMetricsResponse{
		BookmarkID:     metrics.BookmarkID,
		TotalViews:     metrics.TotalViews,
		TotalClicks:    metrics.TotalClicks,
		TotalSaves:     metrics.TotalSaves,
		TotalShares:    metrics.TotalShares,
		TotalLikes:     metrics.TotalLikes,
		UniqueViewers:  metrics.UniqueViewers,
		EngagementRate: metrics.EngagementRate,
		ViralityScore:  metrics.ViralityScore,
		QualityScore:   metrics.QualityScore,
	}, nil
}

// UpdateSocialMetrics updates social engagement metrics
func (s *SocialMetricsService) UpdateSocialMetrics(ctx context.Context, bookmarkID uint, actionType string) error {
	if bookmarkID == 0 {
		return ErrInvalidBookmarkID
	}

	var metrics SocialMetrics
	err := s.db.First(&metrics, "bookmark_id = ?", bookmarkID).Error
	if err == gorm.ErrRecordNotFound {
		// Create new metrics record
		metrics = SocialMetrics{
			BookmarkID:     bookmarkID,
			LastCalculated: time.Now(),
		}
	} else if err != nil {
		return fmt.Errorf("failed to get social metrics: %w", err)
	}

	// Update metrics based on action type
	s.updateMetricsByAction(&metrics, actionType)

	// Calculate derived metrics
	s.calculateDerivedMetrics(&metrics)

	metrics.LastCalculated = time.Now()

	// Save or create metrics
	if metrics.ID == 0 {
		err = s.db.Create(&metrics).Error
	} else {
		err = s.db.Save(&metrics).Error
	}

	if err != nil {
		return fmt.Errorf("failed to update social metrics: %w", err)
	}

	return nil
}

// updateMetricsByAction updates metrics based on action type
func (s *SocialMetricsService) updateMetricsByAction(metrics *SocialMetrics, actionType string) {
	switch actionType {
	case "view":
		metrics.TotalViews++
	case "click":
		metrics.TotalClicks++
	case "save":
		metrics.TotalSaves++
	case "share":
		metrics.TotalShares++
	case "like":
		metrics.TotalLikes++
	}
}

// calculateDerivedMetrics calculates engagement rate, virality, and quality scores
func (s *SocialMetricsService) calculateDerivedMetrics(metrics *SocialMetrics) {
	// Calculate engagement rate
	if metrics.TotalViews > 0 {
		totalEngagement := metrics.TotalClicks + metrics.TotalSaves + metrics.TotalShares + metrics.TotalLikes
		metrics.EngagementRate = float64(totalEngagement) / float64(metrics.TotalViews)
	}

	// Calculate virality score (simplified)
	metrics.ViralityScore = float64(metrics.TotalShares)*2.0 + float64(metrics.TotalSaves)*1.5

	// Calculate quality score (simplified)
	if metrics.TotalViews > 0 {
		metrics.QualityScore = (metrics.EngagementRate * 0.6) + (float64(metrics.TotalLikes) / float64(metrics.TotalViews) * 0.4)
	}
}
