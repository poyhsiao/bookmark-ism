package community

import (
	"time"

	"gorm.io/gorm"
)

// UserBehavior tracks user interactions for recommendation engine
type UserBehavior struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	UserID     string         `json:"user_id" gorm:"not null;index"`
	BookmarkID uint           `json:"bookmark_id" gorm:"not null;index"`
	ActionType string         `json:"action_type" gorm:"not null"` // view, click, save, share, like
	Duration   int            `json:"duration"`                    // time spent in seconds
	Context    string         `json:"context"`                     // search, recommendation, trending, etc.
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Metadata   string         `json:"metadata" gorm:"type:text"` // JSON metadata
}

// UserFollow represents user following relationships
type UserFollow struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	FollowerID  string         `json:"follower_id" gorm:"not null;index"`
	FollowingID string         `json:"following_id" gorm:"not null;index"`
	Status      string         `json:"status" gorm:"default:'active'"` // active, blocked, muted
}

// BookmarkRecommendation stores generated recommendations
type BookmarkRecommendation struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	UserID      string         `json:"user_id" gorm:"not null;index"`
	BookmarkID  uint           `json:"bookmark_id" gorm:"not null;index"`
	Score       float64        `json:"score" gorm:"not null"`        // recommendation confidence score
	ReasonType  string         `json:"reason_type" gorm:"not null"`  // similar_users, content_based, trending
	ReasonData  string         `json:"reason_data" gorm:"type:text"` // JSON explanation
	IsViewed    bool           `json:"is_viewed" gorm:"default:false"`
	IsClicked   bool           `json:"is_clicked" gorm:"default:false"`
	IsDismissed bool           `json:"is_dismissed" gorm:"default:false"`
	ExpiresAt   *time.Time     `json:"expires_at"`
}

// TrendingBookmark tracks bookmark popularity metrics
type TrendingBookmark struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	BookmarkID    uint           `json:"bookmark_id" gorm:"not null;uniqueIndex"`
	ViewCount     int            `json:"view_count" gorm:"default:0"`
	ClickCount    int            `json:"click_count" gorm:"default:0"`
	SaveCount     int            `json:"save_count" gorm:"default:0"`
	ShareCount    int            `json:"share_count" gorm:"default:0"`
	LikeCount     int            `json:"like_count" gorm:"default:0"`
	TrendingScore float64        `json:"trending_score" gorm:"default:0"`
	TimeWindow    string         `json:"time_window" gorm:"not null"` // hourly, daily, weekly
	CalculatedAt  time.Time      `json:"calculated_at"`
}

// UserFeed represents personalized feed items
type UserFeed struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	UserID     string         `json:"user_id" gorm:"not null;index"`
	BookmarkID uint           `json:"bookmark_id" gorm:"not null;index"`
	SourceType string         `json:"source_type" gorm:"not null"` // following, trending, recommended
	SourceID   string         `json:"source_id"`                   // user_id for following, algorithm for others
	Score      float64        `json:"score" gorm:"not null"`
	IsViewed   bool           `json:"is_viewed" gorm:"default:false"`
	IsClicked  bool           `json:"is_clicked" gorm:"default:false"`
	Position   int            `json:"position"`
}

// SocialMetrics aggregates social engagement data
type SocialMetrics struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	BookmarkID     uint           `json:"bookmark_id" gorm:"not null;uniqueIndex"`
	TotalViews     int            `json:"total_views" gorm:"default:0"`
	TotalClicks    int            `json:"total_clicks" gorm:"default:0"`
	TotalSaves     int            `json:"total_saves" gorm:"default:0"`
	TotalShares    int            `json:"total_shares" gorm:"default:0"`
	TotalLikes     int            `json:"total_likes" gorm:"default:0"`
	UniqueViewers  int            `json:"unique_viewers" gorm:"default:0"`
	EngagementRate float64        `json:"engagement_rate" gorm:"default:0"`
	ViralityScore  float64        `json:"virality_score" gorm:"default:0"`
	QualityScore   float64        `json:"quality_score" gorm:"default:0"`
	LastCalculated time.Time      `json:"last_calculated"`
}

// Validation methods
func (ub *UserBehavior) Validate() error {
	if ub.UserID == "" {
		return ErrInvalidUserID
	}
	if ub.BookmarkID == 0 {
		return ErrInvalidBookmarkID
	}
	if ub.ActionType == "" {
		return ErrInvalidActionType
	}
	validActions := map[string]bool{
		"view": true, "click": true, "save": true, "share": true, "like": true,
		"dismiss": true, "report": true, "follow": true, "unfollow": true,
	}
	if !validActions[ub.ActionType] {
		return ErrInvalidActionType
	}
	return nil
}

func (uf *UserFollow) Validate() error {
	if uf.FollowerID == "" {
		return ErrInvalidFollowerID
	}
	if uf.FollowingID == "" {
		return ErrInvalidFollowingID
	}
	if uf.FollowerID == uf.FollowingID {
		return ErrCannotFollowSelf
	}
	validStatuses := map[string]bool{
		"active": true, "blocked": true, "muted": true,
	}
	if !validStatuses[uf.Status] {
		return ErrInvalidFollowStatus
	}
	return nil
}

func (br *BookmarkRecommendation) Validate() error {
	if br.UserID == "" {
		return ErrInvalidUserID
	}
	if br.BookmarkID == 0 {
		return ErrInvalidBookmarkID
	}
	if br.Score < 0 || br.Score > 1 {
		return ErrInvalidScore
	}
	validReasons := map[string]bool{
		"similar_users": true, "content_based": true, "trending": true,
		"collaborative": true, "popularity": true, "category": true,
	}
	if !validReasons[br.ReasonType] {
		return ErrInvalidReasonType
	}
	return nil
}

// Request/Response models
type RecommendationRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	Limit     int    `json:"limit" binding:"min=1,max=100"`
	Algorithm string `json:"algorithm"` // collaborative, content, hybrid, trending
	Context   string `json:"context"`   // homepage, search, category
}

type TrendingRequest struct {
	TimeWindow string  `json:"time_window"` // hourly, daily, weekly, monthly
	Category   string  `json:"category"`
	Limit      int     `json:"limit" binding:"min=1,max=100"`
	MinScore   float64 `json:"min_score"`
}

type FollowRequest struct {
	FollowingID string `json:"following_id" binding:"required"`
}

type FeedRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Limit      int    `json:"limit" binding:"min=1,max=100"`
	Offset     int    `json:"offset" binding:"min=0"`
	SourceType string `json:"source_type"` // following, trending, recommended, all
}

type BehaviorTrackingRequest struct {
	UserID     string                 `json:"user_id"` // Set by handler from auth context
	BookmarkID uint                   `json:"bookmark_id" binding:"required"`
	ActionType string                 `json:"action_type" binding:"required"`
	Duration   int                    `json:"duration"`
	Context    string                 `json:"context"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Response models
type RecommendationResponse struct {
	BookmarkID   uint        `json:"bookmark_id"`
	Score        float64     `json:"score"`
	ReasonType   string      `json:"reason_type"`
	ReasonText   string      `json:"reason_text"`
	BookmarkData interface{} `json:"bookmark_data,omitempty"`
}

type TrendingResponse struct {
	BookmarkID     uint        `json:"bookmark_id"`
	TrendingScore  float64     `json:"trending_score"`
	ViewCount      int         `json:"view_count"`
	EngagementRate float64     `json:"engagement_rate"`
	TimeWindow     string      `json:"time_window"`
	BookmarkData   interface{} `json:"bookmark_data,omitempty"`
}

type UserFeedResponse struct {
	BookmarkID   uint        `json:"bookmark_id"`
	SourceType   string      `json:"source_type"`
	SourceID     string      `json:"source_id"`
	Score        float64     `json:"score"`
	Position     int         `json:"position"`
	BookmarkData interface{} `json:"bookmark_data,omitempty"`
}

type SocialMetricsResponse struct {
	BookmarkID     uint    `json:"bookmark_id"`
	TotalViews     int     `json:"total_views"`
	TotalClicks    int     `json:"total_clicks"`
	TotalSaves     int     `json:"total_saves"`
	TotalShares    int     `json:"total_shares"`
	TotalLikes     int     `json:"total_likes"`
	UniqueViewers  int     `json:"unique_viewers"`
	EngagementRate float64 `json:"engagement_rate"`
	ViralityScore  float64 `json:"virality_score"`
	QualityScore   float64 `json:"quality_score"`
}

type UserStatsResponse struct {
	UserID          string  `json:"user_id"`
	FollowersCount  int     `json:"followers_count"`
	FollowingCount  int     `json:"following_count"`
	BookmarksCount  int     `json:"bookmarks_count"`
	PublicBookmarks int     `json:"public_bookmarks"`
	TotalViews      int     `json:"total_views"`
	TotalEngagement int     `json:"total_engagement"`
	InfluenceScore  float64 `json:"influence_score"`
}
