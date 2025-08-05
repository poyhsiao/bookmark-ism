package community

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TrendingService handles trending calculations and retrieval
type TrendingService struct {
	db         Database
	redis      RedisClient
	jsonHelper *JSONHelper
	logger     *zap.Logger
}

// NewTrendingService creates a new trending service
func NewTrendingService(db Database, redis RedisClient, jsonHelper *JSONHelper, logger *zap.Logger) *TrendingService {
	return &TrendingService{
		db:         db,
		redis:      redis,
		jsonHelper: jsonHelper,
		logger:     logger,
	}
}

// GetTrendingBookmarksInternal returns trending bookmarks
func (s *TrendingService) GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error) {
	if req.TimeWindow == "" {
		req.TimeWindow = "daily"
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	// Validate time window
	configHelper := NewConfigHelper()
	if !configHelper.ValidateTimeWindow(req.TimeWindow) {
		return nil, ErrInvalidTimeWindow
	}

	// Get trending bookmarks from database
	var trendingBookmarks []TrendingBookmark
	query := s.db.Where("time_window = ?", req.TimeWindow)

	if req.MinScore > 0 {
		query = query.Where("trending_score >= ?", req.MinScore)
	}

	err := query.Order("trending_score DESC").Limit(req.Limit).Find(&trendingBookmarks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get trending bookmarks: %w", err)
	}

	// Convert to response format
	trending := make([]TrendingResponse, len(trendingBookmarks))
	for i, tb := range trendingBookmarks {
		engagementRate := s.calculateEngagementRate(tb)
		trending[i] = TrendingResponse{
			BookmarkID:     tb.BookmarkID,
			TrendingScore:  tb.TrendingScore,
			ViewCount:      tb.ViewCount,
			EngagementRate: engagementRate,
			TimeWindow:     tb.TimeWindow,
		}
	}

	return trending, nil
}

// CalculateTrendingScores calculates and updates trending scores for bookmarks
func (s *TrendingService) CalculateTrendingScores(ctx context.Context, timeWindow string) error {
	configHelper := NewConfigHelper()
	if !configHelper.ValidateTimeWindow(timeWindow) {
		return ErrInvalidTimeWindow
	}

	// Calculate time range
	startTime, err := configHelper.GetTimeRange(timeWindow)
	if err != nil {
		return err
	}

	// Get user behaviors in time window
	var behaviors []UserBehavior
	err = s.db.Where("created_at >= ?", startTime).Find(&behaviors).Error
	if err != nil {
		return fmt.Errorf("failed to get user behaviors: %w", err)
	}

	// Aggregate metrics by bookmark
	bookmarkMetrics := s.aggregateMetrics(behaviors, timeWindow)

	// Calculate trending scores and save
	return s.saveTrendingMetrics(ctx, bookmarkMetrics, timeWindow)
}

// UpdateTrendingCache updates the trending cache for a bookmark
func (s *TrendingService) UpdateTrendingCache(ctx context.Context, bookmarkID uint, actionType string) error {
	// Update Redis sorted set for real-time trending
	key := fmt.Sprintf("trending:%s", actionType)
	score := time.Now().Unix()

	return s.redis.ZAdd(ctx, key, score, strconv.Itoa(int(bookmarkID)))
}

// calculateEngagementRate calculates engagement rate for trending bookmark
func (s *TrendingService) calculateEngagementRate(tb TrendingBookmark) float64 {
	if tb.ViewCount == 0 {
		return 0.0
	}
	return float64(tb.ClickCount+tb.SaveCount+tb.ShareCount) / float64(tb.ViewCount)
}

// aggregateMetrics aggregates user behaviors into bookmark metrics
func (s *TrendingService) aggregateMetrics(behaviors []UserBehavior, timeWindow string) map[uint]*TrendingBookmark {
	now := time.Now()
	bookmarkMetrics := make(map[uint]*TrendingBookmark)

	for _, behavior := range behaviors {
		if bookmarkMetrics[behavior.BookmarkID] == nil {
			bookmarkMetrics[behavior.BookmarkID] = &TrendingBookmark{
				BookmarkID:   behavior.BookmarkID,
				TimeWindow:   timeWindow,
				CalculatedAt: now,
			}
		}

		metric := bookmarkMetrics[behavior.BookmarkID]
		s.updateMetricCounts(metric, behavior.ActionType)
	}

	return bookmarkMetrics
}

// updateMetricCounts updates metric counts based on action type
func (s *TrendingService) updateMetricCounts(metric *TrendingBookmark, actionType string) {
	switch actionType {
	case "view":
		metric.ViewCount++
	case "click":
		metric.ClickCount++
	case "save":
		metric.SaveCount++
	case "share":
		metric.ShareCount++
	case "like":
		metric.LikeCount++
	}
}

// saveTrendingMetrics calculates trending scores and saves metrics
func (s *TrendingService) saveTrendingMetrics(ctx context.Context, bookmarkMetrics map[uint]*TrendingBookmark, timeWindow string) error {
	now := time.Now()

	for _, metric := range bookmarkMetrics {
		// Calculate trending score using weighted formula
		metric.TrendingScore = s.calculateTrendingScore(metric, now, timeWindow)

		// Save or update trending bookmark
		if err := s.saveOrUpdateTrendingBookmark(metric, timeWindow); err != nil {
			s.logger.Error("Failed to save trending bookmark", zap.Error(err), zap.Uint("bookmark_id", metric.BookmarkID))
			continue
		}
	}

	return nil
}

// calculateTrendingScore calculates trending score with time decay
func (s *TrendingService) calculateTrendingScore(metric *TrendingBookmark, now time.Time, timeWindow string) float64 {
	// Calculate trending score using weighted formula
	viewWeight := 1.0
	clickWeight := 2.0
	saveWeight := 3.0
	shareWeight := 4.0
	likeWeight := 2.5

	rawScore := float64(metric.ViewCount)*viewWeight +
		float64(metric.ClickCount)*clickWeight +
		float64(metric.SaveCount)*saveWeight +
		float64(metric.ShareCount)*shareWeight +
		float64(metric.LikeCount)*likeWeight

	// Apply time decay
	hoursAgo := now.Sub(metric.CalculatedAt).Hours()
	timeDecay := math.Exp(-hoursAgo / 24.0) // Decay over 24 hours

	return rawScore * timeDecay
}

// saveOrUpdateTrendingBookmark saves or updates trending bookmark record
func (s *TrendingService) saveOrUpdateTrendingBookmark(metric *TrendingBookmark, timeWindow string) error {
	var existing TrendingBookmark
	err := s.db.First(&existing, "bookmark_id = ? AND time_window = ?", metric.BookmarkID, timeWindow).Error

	if err == gorm.ErrRecordNotFound {
		return s.db.Create(metric).Error
	} else if err == nil {
		// Update existing record
		existing.ViewCount = metric.ViewCount
		existing.ClickCount = metric.ClickCount
		existing.SaveCount = metric.SaveCount
		existing.ShareCount = metric.ShareCount
		existing.LikeCount = metric.LikeCount
		existing.TrendingScore = metric.TrendingScore
		existing.CalculatedAt = metric.CalculatedAt
		return s.db.Save(&existing).Error
	}

	return err
}
