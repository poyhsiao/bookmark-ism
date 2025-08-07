package worker

import (
	"context"
	"fmt"
	"time"

	"bookmark-sync-service/backend/pkg/database"

	"go.uber.org/zap"
)

// SocialMetricsUpdateJob handles updating social metrics asynchronously
type SocialMetricsUpdateJob struct {
	BaseJob
	BookmarkID uint
	ActionType string
	Service    SocialMetricsService
	Logger     *zap.Logger
}

// SocialMetricsService defines the interface for social metrics operations
type SocialMetricsService interface {
	UpdateSocialMetrics(ctx context.Context, bookmarkID uint, actionType string) error
}

// NewSocialMetricsUpdateJob creates a new social metrics update job
func NewSocialMetricsUpdateJob(bookmarkID uint, actionType string, service SocialMetricsService, logger *zap.Logger) *SocialMetricsUpdateJob {
	return &SocialMetricsUpdateJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("social-metrics-%d-%s-%d", bookmarkID, actionType, time.Now().UnixNano()),
			Type:       "social_metrics_update",
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		BookmarkID: bookmarkID,
		ActionType: actionType,
		Service:    service,
		Logger:     logger,
	}
}

func (j *SocialMetricsUpdateJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing social metrics update job",
		zap.String("job_id", j.ID),
		zap.Uint("bookmark_id", j.BookmarkID),
		zap.String("action_type", j.ActionType))

	return j.Service.UpdateSocialMetrics(ctx, j.BookmarkID, j.ActionType)
}

// TrendingCacheUpdateJob handles updating trending cache asynchronously
type TrendingCacheUpdateJob struct {
	BaseJob
	BookmarkID uint
	ActionType string
	Service    TrendingCacheService
	Logger     *zap.Logger
}

// TrendingCacheService defines the interface for trending cache operations
type TrendingCacheService interface {
	UpdateTrendingCache(ctx context.Context, bookmarkID uint, actionType string) error
}

// NewTrendingCacheUpdateJob creates a new trending cache update job
func NewTrendingCacheUpdateJob(bookmarkID uint, actionType string, service TrendingCacheService, logger *zap.Logger) *TrendingCacheUpdateJob {
	return &TrendingCacheUpdateJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("trending-cache-%d-%s-%d", bookmarkID, actionType, time.Now().UnixNano()),
			Type:       "trending_cache_update",
			MaxRetries: 2,
			CreatedAt:  time.Now(),
		},
		BookmarkID: bookmarkID,
		ActionType: actionType,
		Service:    service,
		Logger:     logger,
	}
}

func (j *TrendingCacheUpdateJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing trending cache update job",
		zap.String("job_id", j.ID),
		zap.Uint("bookmark_id", j.BookmarkID),
		zap.String("action_type", j.ActionType))

	return j.Service.UpdateTrendingCache(ctx, j.BookmarkID, j.ActionType)
}

// ThemeRatingUpdateJob handles updating theme ratings asynchronously
type ThemeRatingUpdateJob struct {
	BaseJob
	ThemeID uint
	Service ThemeRatingService
	Logger  *zap.Logger
}

// ThemeRatingService defines the interface for theme rating operations
type ThemeRatingService interface {
	UpdateThemeRating(ctx context.Context, themeID uint) error
}

// NewThemeRatingUpdateJob creates a new theme rating update job
func NewThemeRatingUpdateJob(themeID uint, service ThemeRatingService, logger *zap.Logger) *ThemeRatingUpdateJob {
	return &ThemeRatingUpdateJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("theme-rating-%d-%d", themeID, time.Now().UnixNano()),
			Type:       "theme_rating_update",
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		ThemeID: themeID,
		Service: service,
		Logger:  logger,
	}
}

func (j *ThemeRatingUpdateJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing theme rating update job",
		zap.String("job_id", j.ID),
		zap.Uint("theme_id", j.ThemeID))

	return j.Service.UpdateThemeRating(ctx, j.ThemeID)
}

// LinkCheckerJob handles checking bookmark links for validity
type LinkCheckerJob struct {
	BaseJob
	Bookmark *database.Bookmark
	Service  LinkCheckerService
	Logger   *zap.Logger
}

// LinkCheckerService defines the interface for link checking operations
type LinkCheckerService interface {
	CheckLink(ctx context.Context, bookmark *database.Bookmark) error
}

// NewLinkCheckerJob creates a new link checker job
func NewLinkCheckerJob(bookmark *database.Bookmark, service LinkCheckerService, logger *zap.Logger) *LinkCheckerJob {
	return &LinkCheckerJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("link-check-%d-%d", bookmark.ID, time.Now().UnixNano()),
			Type:       "link_checker",
			MaxRetries: 2,
			CreatedAt:  time.Now(),
		},
		Bookmark: bookmark,
		Service:  service,
		Logger:   logger,
	}
}

func (j *LinkCheckerJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing link checker job",
		zap.String("job_id", j.ID),
		zap.Uint("bookmark_id", j.Bookmark.ID),
		zap.String("url", j.Bookmark.URL))

	return j.Service.CheckLink(ctx, j.Bookmark)
}

// CleanupJob handles cleanup operations
type CleanupJob struct {
	BaseJob
	Service CleanupService
	Logger  *zap.Logger
}

// CleanupService defines the interface for cleanup operations
type CleanupService interface {
	RunCleanup(ctx context.Context) error
}

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(service CleanupService, logger *zap.Logger) *CleanupJob {
	return &CleanupJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("cleanup-%d", time.Now().UnixNano()),
			Type:       "cleanup",
			MaxRetries: 1,
			CreatedAt:  time.Now(),
		},
		Service: service,
		Logger:  logger,
	}
}

func (j *CleanupJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing cleanup job", zap.String("job_id", j.ID))
	return j.Service.RunCleanup(ctx)
}

// EmailNotificationJob handles sending email notifications
type EmailNotificationJob struct {
	BaseJob
	To      string
	Subject string
	Body    string
	Service EmailService
	Logger  *zap.Logger
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// NewEmailNotificationJob creates a new email notification job
func NewEmailNotificationJob(to, subject, body string, service EmailService, logger *zap.Logger) *EmailNotificationJob {
	return &EmailNotificationJob{
		BaseJob: BaseJob{
			ID:         fmt.Sprintf("email-%s-%d", to, time.Now().UnixNano()),
			Type:       "email_notification",
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		To:      to,
		Subject: subject,
		Body:    body,
		Service: service,
		Logger:  logger,
	}
}

func (j *EmailNotificationJob) Execute(ctx context.Context) error {
	j.Logger.Debug("Executing email notification job",
		zap.String("job_id", j.ID),
		zap.String("to", j.To),
		zap.String("subject", j.Subject))

	return j.Service.SendEmail(ctx, j.To, j.Subject, j.Body)
}
