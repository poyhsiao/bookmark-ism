package community

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CommunityService interface for dependency injection
type CommunityService interface {
	TrackUserBehavior(ctx context.Context, req *BehaviorTrackingRequest) error
	FollowUser(ctx context.Context, followerID string, req *FollowRequest) error
	UnfollowUser(ctx context.Context, followerID, followingID string) error
	GetRecommendations(ctx context.Context, req *RecommendationRequest) ([]RecommendationResponse, error)
	GetTrendingBookmarksInternal(ctx context.Context, req *TrendingRequest) ([]TrendingResponse, error)
	GenerateUserFeed(ctx context.Context, req *FeedRequest) ([]UserFeedResponse, error)
	GetSocialMetrics(ctx context.Context, bookmarkID uint) (*SocialMetricsResponse, error)
	GetUserStats(ctx context.Context, userID string) (*UserStatsResponse, error)
	GenerateRecommendations(ctx context.Context, userID, algorithm string) error
	CalculateTrendingScores(ctx context.Context, timeWindow string) error
}

// Handler handles HTTP requests for community features
type Handler struct {
	service CommunityService
}

// NewHandler creates a new community handler
func NewHandler(service CommunityService) *Handler {
	return &Handler{
		service: service,
	}
}

// TrackBehavior handles POST /api/v1/community/behavior
func (h *Handler) TrackBehavior(c *gin.Context) {
	var req BehaviorTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}
	req.UserID = userID.(string)

	// Get client IP and user agent
	req.Metadata = make(map[string]interface{})
	req.Metadata["ip_address"] = c.ClientIP()
	req.Metadata["user_agent"] = c.GetHeader("User-Agent")

	if err := h.service.TrackUserBehavior(c.Request.Context(), &req); err != nil {
		switch err {
		case ErrInvalidUserID, ErrInvalidBookmarkID, ErrInvalidActionType:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to track behavior"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Behavior tracked successfully"})
}

// FollowUser handles POST /api/v1/community/follow
func (h *Handler) FollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, "Invalid request body"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}

	if err := h.service.FollowUser(c.Request.Context(), userID.(string), &req); err != nil {
		switch err {
		case ErrInvalidFollowerID, ErrInvalidFollowingID, ErrCannotFollowSelf:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		case ErrAlreadyFollowing:
			c.JSON(http.StatusConflict, NewErrorResponse(err, CodeAlreadyExists, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to follow user"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User followed successfully"})
}

// UnfollowUser handles DELETE /api/v1/community/follow/:user_id
func (h *Handler) UnfollowUser(c *gin.Context) {
	followingID := c.Param("user_id")
	if followingID == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidFollowingID, CodeValidationError, "User ID is required"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}

	if err := h.service.UnfollowUser(c.Request.Context(), userID.(string), followingID); err != nil {
		switch err {
		case ErrNotFollowing:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to unfollow user"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unfollowed successfully"})
}

// GetRecommendations handles GET /api/v1/community/recommendations
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}

	// Parse query parameters
	req := RecommendationRequest{
		UserID:    userID.(string),
		Limit:     20,
		Algorithm: c.Query("algorithm"),
		Context:   c.Query("context"),
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	recommendations, err := h.service.GetRecommendations(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		case ErrInsufficientData:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, "Not enough data for recommendations"))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get recommendations"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"total":           len(recommendations),
	})
}

// GetTrending handles GET /api/v1/community/trending
func (h *Handler) GetTrending(c *gin.Context) {
	req := TrendingRequest{
		TimeWindow: c.DefaultQuery("time_window", "daily"),
		Category:   c.Query("category"),
		Limit:      20,
		MinScore:   0.0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if minScoreStr := c.Query("min_score"); minScoreStr != "" {
		if minScore, err := strconv.ParseFloat(minScoreStr, 64); err == nil && minScore >= 0 {
			req.MinScore = minScore
		}
	}

	trending, err := h.service.GetTrendingBookmarksInternal(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case ErrInvalidTimeWindow:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get trending bookmarks"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trending":    trending,
		"total":       len(trending),
		"time_window": req.TimeWindow,
	})
}

// GetFeed handles GET /api/v1/community/feed
func (h *Handler) GetFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}

	req := FeedRequest{
		UserID:     userID.(string),
		Limit:      20,
		Offset:     0,
		SourceType: c.DefaultQuery("source_type", "all"),
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	feed, err := h.service.GenerateUserFeed(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get user feed"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feed":        feed,
		"total":       len(feed),
		"source_type": req.SourceType,
	})
}

// GetSocialMetrics handles GET /api/v1/community/metrics/:bookmark_id
func (h *Handler) GetSocialMetrics(c *gin.Context) {
	bookmarkIDStr := c.Param("bookmark_id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidBookmarkID, CodeValidationError, "Invalid bookmark ID"))
		return
	}

	metrics, err := h.service.GetSocialMetrics(c.Request.Context(), uint(bookmarkID))
	if err != nil {
		switch err {
		case ErrBookmarkNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, "Bookmark not found"))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get social metrics"))
		}
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUserStats handles GET /api/v1/community/users/:user_id/stats
func (h *Handler) GetUserStats(c *gin.Context) {
	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(ErrInvalidUserID, CodeValidationError, "User ID is required"))
		return
	}

	stats, err := h.service.GetUserStats(c.Request.Context(), targetUserID)
	if err != nil {
		switch err {
		case ErrInvalidUserID:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, "User not found"))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to get user stats"))
		}
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GenerateRecommendations handles POST /api/v1/community/recommendations/generate
func (h *Handler) GenerateRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse(ErrUserNotFound, CodePermissionDenied, "User not authenticated"))
		return
	}

	algorithm := c.DefaultQuery("algorithm", "hybrid")

	if err := h.service.GenerateRecommendations(c.Request.Context(), userID.(string), algorithm); err != nil {
		switch err {
		case ErrInvalidUserID, ErrInvalidAlgorithm:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		case ErrInsufficientData:
			c.JSON(http.StatusNotFound, NewErrorResponse(err, CodeNotFound, "Not enough data to generate recommendations"))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to generate recommendations"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recommendations generated successfully"})
}

// CalculateTrending handles POST /api/v1/community/trending/calculate
func (h *Handler) CalculateTrending(c *gin.Context) {
	// This endpoint might be restricted to admin users or scheduled jobs
	timeWindow := c.DefaultQuery("time_window", "daily")

	if err := h.service.CalculateTrendingScores(c.Request.Context(), timeWindow); err != nil {
		switch err {
		case ErrInvalidTimeWindow:
			c.JSON(http.StatusBadRequest, NewErrorResponse(err, CodeValidationError, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err, CodeInternalError, "Failed to calculate trending scores"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trending scores calculated successfully"})
}

// RegisterRoutes registers all community routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	community := router.Group("/community")
	{
		// Behavior tracking
		community.POST("/behavior", h.TrackBehavior)

		// User following
		community.POST("/follow", h.FollowUser)
		community.DELETE("/follow/:user_id", h.UnfollowUser)

		// Recommendations
		community.GET("/recommendations", h.GetRecommendations)
		community.POST("/recommendations/generate", h.GenerateRecommendations)

		// Trending
		community.GET("/trending", h.GetTrending)
		community.POST("/trending/calculate", h.CalculateTrending)

		// Feed
		community.GET("/feed", h.GetFeed)

		// Social metrics
		community.GET("/metrics/:bookmark_id", h.GetSocialMetrics)

		// User stats
		community.GET("/users/:user_id/stats", h.GetUserStats)
	}
}
