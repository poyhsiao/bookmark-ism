package community

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"bookmark-sync-service/backend/pkg/worker"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Database interface for dependency injection
type Database interface {
	Create(value interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) Database
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Order(value interface{}) Database
	Limit(limit int) Database
	Offset(offset int) Database
}

// Redis interface for caching
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	ZAdd(ctx context.Context, key string, members ...interface{}) error
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
}

// Service handles community features
type Service struct {
	db         Database
	redis      RedisClient
	workerPool *worker.WorkerPool
	logger     *zap.Logger
}

// NewService creates a new community service
func NewService(db Database, redis RedisClient, workerPool *worker.WorkerPool, logger *zap.Logger) *Service {
	return &Service{
		db:         db,
		redis:      redis,
		workerPool: workerPool,
		logger:     logger,
	}
}

// TrackUserBehavior records user interactions for recommendation engine
func (s *Service) TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error {
	// Validate request
	if req.UserID == "" {
		return ErrInvalidUserID
	}
	if req.BookmarkID == 0 {
		return ErrInvalidBookmarkID
	}
	if req.ActionType == "" {
		return ErrInvalidActionType
	}

	// Create behavior record
	behavior := &UserBehavior{
		UserID:     req.UserID,
		BookmarkID: req.BookmarkID,
		ActionType: req.ActionType,
		Duration:   req.Duration,
		Context:    req.Context,
	}

	// Serialize metadata
	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err == nil {
			behavior.Metadata = string(metadataJSON)
		}
	}

	// Validate behavior
	if err := behavior.Validate(); err != nil {
		return err
	}

	// Save to database
	if err := s.db.Create(behavior).Error; err != nil {
		return fmt.Errorf("failed to track user behavior: %w", err)
	}

	// Update social metrics asynchronously using worker queue
	if s.workerPool != nil {
		socialJob := worker.NewSocialMetricsUpdateJob(req.BookmarkID, req.ActionType, s, s.logger)
		if err := s.workerPool.Submit(socialJob); err != nil {
			s.logger.Warn("Failed to submit social metrics update job", zap.Error(err))
		}
	}

	// Update trending scores if significant action
	if req.ActionType == "view" || req.ActionType == "click" || req.ActionType == "save" {
		if s.workerPool != nil {
			trendingJob := worker.NewTrendingCacheUpdateJob(req.BookmarkID, req.ActionType, s, s.logger)
			if err := s.workerPool.Submit(trendingJob); err != nil {
				s.logger.Warn("Failed to submit trending cache update job", zap.Error(err))
			}
		}
	}

	return nil
}

// FollowUser creates a following relationship
func (s *Service) FollowUser(ctx context.Context, followerID string, req *FollowRequest) error {
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
	var existing UserFollow
	err := s.db.First(&existing, "follower_id = ? AND following_id = ?", followerID, req.FollowingID).Error
	if err == nil {
		return ErrAlreadyFollowing
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing follow: %w", err)
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

	// Clear cache
	s.clearUserStatsCache(ctx, followerID)
	s.clearUserStatsCache(ctx, req.FollowingID)

	return nil
}

// UnfollowUser removes a following relationship
func (s *Service) UnfollowUser(ctx context.Context, followerID, followingID string) error {
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

	// Clear cache
	s.clearUserStatsCache(ctx, followerID)
	s.clearUserStatsCache(ctx, followingID)

	return nil
}

// GetRecommendations returns personalized bookmark recommendations
func (s *Service) GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	// Check cache first
	cacheKey := fmt.Sprintf("recommendations:%s:%s:%s", req.UserID, req.Algorithm, req.Context)
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var recommendations []RecommendationResponse
		if json.Unmarshal([]byte(cached), &recommendations) == nil {
			return recommendations, nil
		}
	}

	// Get recommendations from database
	var dbRecommendations []BookmarkRecommendation
	query := s.db.Where("user_id = ? AND expires_at > ?", req.UserID, time.Now())

	if req.Algorithm != "" {
		query = query.Where("reason_type = ?", req.Algorithm)
	}

	err := query.Order("score DESC").Limit(req.Limit).Find(&dbRecommendations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Convert to response format
	recommendations := make([]RecommendationResponse, len(dbRecommendations))
	for i, rec := range dbRecommendations {
		recommendations[i] = RecommendationResponse{
			BookmarkID: rec.BookmarkID,
			Score:      rec.Score,
			ReasonType: rec.ReasonType,
			ReasonText: s.generateReasonText(rec.ReasonType, rec.ReasonData),
		}
	}

	// Cache results
	if cacheData, err := json.Marshal(recommendations); err == nil {
		s.redis.Set(ctx, cacheKey, cacheData, 15*time.Minute)
	}

	return recommendations, nil
}

// GetTrendingBookmarksInternal returns trending bookmarks (internal method for testing)
func (s *Service) GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error) {
	if req.TimeWindow == "" {
		req.TimeWindow = "daily"
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	// Validate time window
	validWindows := map[string]bool{
		"hourly": true, "daily": true, "weekly": true, "monthly": true,
	}
	if !validWindows[req.TimeWindow] {
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
		engagementRate := 0.0
		if tb.ViewCount > 0 {
			engagementRate = float64(tb.ClickCount+tb.SaveCount+tb.ShareCount) / float64(tb.ViewCount)
		}

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

// GenerateUserFeed creates personalized feed for user
func (s *Service) GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error) {
	if req.UserID == "" {
		return nil, ErrInvalidUserID
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	// Get feed items from database
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

	// Convert to response format
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

	return feed, nil
}

// GetSocialMetrics returns social engagement metrics for a bookmark
func (s *Service) GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error) {
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
func (s *Service) UpdateSocialMetrics(ctx context.Context, bookmarkID uint, actionType string) error {
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

// GetUserStats returns user statistics and influence metrics
func (s *Service) GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user_stats:%s", userID)
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil && cached != "" {
		var stats UserStatsResponse
		if json.Unmarshal([]byte(cached), &stats) == nil {
			return &stats, nil
		}
	}

	stats := &UserStatsResponse{
		UserID: userID,
	}

	// Get followers count
	var followers []UserFollow
	s.db.Where("following_id = ? AND status = ?", userID, "active").Find(&followers)
	stats.FollowersCount = len(followers)

	// Get following count
	var following []UserFollow
	s.db.Where("follower_id = ? AND status = ?", userID, "active").Find(&following)
	stats.FollowingCount = len(following)

	// Calculate influence score (simplified)
	stats.InfluenceScore = float64(stats.FollowersCount)*0.7 + float64(stats.TotalEngagement)*0.3

	// Cache results
	if cacheData, err := json.Marshal(stats); err == nil {
		s.redis.Set(ctx, cacheKey, cacheData, 30*time.Minute)
	}

	return stats, nil
}

// CalculateTrendingScores calculates and updates trending scores for bookmarks
func (s *Service) CalculateTrendingScores(ctx context.Context, timeWindow string) error {
	validWindows := map[string]bool{
		"hourly": true, "daily": true, "weekly": true, "monthly": true,
	}
	if !validWindows[timeWindow] {
		return ErrInvalidTimeWindow
	}

	// Calculate time range
	now := time.Now()
	var startTime time.Time
	switch timeWindow {
	case "hourly":
		startTime = now.Add(-1 * time.Hour)
	case "daily":
		startTime = now.Add(-24 * time.Hour)
	case "weekly":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "monthly":
		startTime = now.Add(-30 * 24 * time.Hour)
	}

	// Get user behaviors in time window
	var behaviors []UserBehavior
	err := s.db.Where("created_at >= ?", startTime).Find(&behaviors).Error
	if err != nil {
		return fmt.Errorf("failed to get user behaviors: %w", err)
	}

	// Aggregate metrics by bookmark
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
		switch behavior.ActionType {
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

	// Calculate trending scores and save
	for _, metric := range bookmarkMetrics {
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
		hoursAgo := now.Sub(startTime).Hours()
		timeDecay := math.Exp(-hoursAgo / 24.0) // Decay over 24 hours

		metric.TrendingScore = rawScore * timeDecay

		// Save or update trending bookmark
		var existing TrendingBookmark
		err := s.db.First(&existing, "bookmark_id = ? AND time_window = ?", metric.BookmarkID, timeWindow).Error
		if err == gorm.ErrRecordNotFound {
			s.db.Create(metric)
		} else if err == nil {
			existing.ViewCount = metric.ViewCount
			existing.ClickCount = metric.ClickCount
			existing.SaveCount = metric.SaveCount
			existing.ShareCount = metric.ShareCount
			existing.LikeCount = metric.LikeCount
			existing.TrendingScore = metric.TrendingScore
			existing.CalculatedAt = metric.CalculatedAt
			s.db.Save(&existing)
		}
	}

	return nil
}

// GenerateRecommendations generates bookmark recommendations for a user
func (s *Service) GenerateRecommendations(ctx context.Context, userID, algorithm string) error {
	if userID == "" {
		return ErrInvalidUserID
	}

	validAlgorithms := map[string]bool{
		"collaborative": true, "content_based": true, "trending": true,
		"popularity": true, "category": true, "hybrid": true,
	}
	if !validAlgorithms[algorithm] {
		return ErrInvalidAlgorithm
	}

	// Get user's behavior history
	var behaviors []UserBehavior
	err := s.db.Where("user_id = ?", userID).Find(&behaviors).Error
	if err != nil {
		return fmt.Errorf("failed to get user behaviors: %w", err)
	}

	// Check for insufficient data
	if len(behaviors) == 0 {
		return ErrInsufficientData
	}

	// Generate recommendations based on algorithm
	var recommendations []BookmarkRecommendation
	switch algorithm {
	case "collaborative":
		recommendations = s.generateCollaborativeRecommendations(userID, behaviors)
	case "content_based":
		recommendations = s.generateContentBasedRecommendations(userID, behaviors)
	case "trending":
		recommendations = s.generateTrendingRecommendations(userID)
	case "popularity":
		recommendations = s.generatePopularityRecommendations(userID)
	default:
		recommendations = s.generateHybridRecommendations(userID, behaviors)
	}

	// Save recommendations
	for _, rec := range recommendations {
		rec.UserID = userID
		rec.ExpiresAt = &time.Time{}
		*rec.ExpiresAt = time.Now().Add(24 * time.Hour) // Expire in 24 hours

		if err := rec.Validate(); err == nil {
			s.db.Create(&rec)
		}
	}

	return nil
}

// Helper methods for recommendation algorithms
func (s *Service) generateCollaborativeRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	// Simplified collaborative filtering
	// In a real implementation, this would use more sophisticated algorithms
	var recommendations []BookmarkRecommendation

	// Find similar users based on common bookmarks
	bookmarkIDs := make([]uint, len(behaviors))
	for i, b := range behaviors {
		bookmarkIDs[i] = b.BookmarkID
	}

	// This is a simplified version - in production, you'd use proper collaborative filtering
	for _, bookmarkID := range bookmarkIDs {
		if len(recommendations) >= 10 {
			break
		}

		rec := BookmarkRecommendation{
			BookmarkID: bookmarkID,
			Score:      0.7, // Simplified score
			ReasonType: "collaborative",
			ReasonData: `{"similar_users": ["user1", "user2"]}`,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (s *Service) generateContentBasedRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	// Simplified content-based filtering
	var recommendations []BookmarkRecommendation

	// Analyze user's bookmark preferences and find similar content
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

func (s *Service) generateTrendingRecommendations(userID string) []BookmarkRecommendation {
	// Get trending bookmarks
	var trending []TrendingBookmark
	s.db.Where("time_window = ?", "daily").Order("trending_score DESC").Limit(10).Find(&trending)

	var recommendations []BookmarkRecommendation
	for _, t := range trending {
		rec := BookmarkRecommendation{
			BookmarkID: t.BookmarkID,
			Score:      math.Min(t.TrendingScore/100.0, 1.0), // Normalize score
			ReasonType: "trending",
			ReasonData: fmt.Sprintf(`{"trending_score": %.2f}`, t.TrendingScore),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (s *Service) generatePopularityRecommendations(userID string) []BookmarkRecommendation {
	// Get popular bookmarks based on social metrics
	var metrics []SocialMetrics
	s.db.Order("total_views DESC").Limit(10).Find(&metrics)

	var recommendations []BookmarkRecommendation
	for _, m := range metrics {
		score := float64(m.TotalViews) / 1000.0 // Normalize
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

func (s *Service) generateHybridRecommendations(userID string, behaviors []UserBehavior) []BookmarkRecommendation {
	// Combine multiple algorithms
	collaborative := s.generateCollaborativeRecommendations(userID, behaviors)
	trending := s.generateTrendingRecommendations(userID)

	// Merge and weight recommendations
	var recommendations []BookmarkRecommendation

	// Add collaborative with higher weight
	for i, rec := range collaborative {
		if i >= 5 {
			break
		}
		rec.Score *= 0.8 // Weight collaborative recommendations
		rec.ReasonType = "hybrid"
		recommendations = append(recommendations, rec)
	}

	// Add trending with lower weight
	for i, rec := range trending {
		if i >= 5 {
			break
		}
		rec.Score *= 0.6 // Weight trending recommendations
		rec.ReasonType = "hybrid"
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// Helper methods
func (s *Service) generateReasonText(reasonType, reasonData string) string {
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

func (s *Service) updateTrendingCache(ctx context.Context, bookmarkID uint, actionType string) {
	// Update Redis sorted set for real-time trending
	key := fmt.Sprintf("trending:%s", actionType)
	score := time.Now().Unix()

	s.redis.ZAdd(ctx, key, score, strconv.Itoa(int(bookmarkID)))
}

func (s *Service) clearUserStatsCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("user_stats:%s", userID)
	s.redis.Del(ctx, cacheKey)
}

// UpdateTrendingCache updates the trending cache for a bookmark
func (s *Service) UpdateTrendingCache(ctx context.Context, bookmarkID uint, actionType string) error {
	s.updateTrendingCache(ctx, bookmarkID, actionType)
	return nil
}
