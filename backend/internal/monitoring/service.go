package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service handles link monitoring and maintenance operations
type Service struct {
	db         *gorm.DB
	httpClient *http.Client
}

// NewService creates a new monitoring service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects - we want to detect them
				return http.ErrUseLastResponse
			},
		},
	}
}

// CheckLink performs a link check and returns the result
func (s *Service) CheckLink(ctx context.Context, userID uint, req *CreateLinkCheckRequest) (*LinkCheck, error) {
	// Verify bookmark belongs to user
	var bookmark struct {
		ID     uint
		UserID uint
		URL    string
	}

	if err := s.db.WithContext(ctx).
		Table("bookmarks").
		Select("id, user_id, url").
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", req.BookmarkID, userID).
		First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("bookmark not found")
		}
		return nil, fmt.Errorf("failed to verify bookmark: %w", err)
	}

	// Perform the actual link check
	linkCheck := &LinkCheck{
		BookmarkID: req.BookmarkID,
		URL:        req.URL,
		CheckedAt:  time.Now(),
	}

	start := time.Now()
	resp, err := s.httpClient.Get(req.URL)
	responseTime := time.Since(start).Milliseconds()
	linkCheck.ResponseTime = responseTime

	if err != nil {
		linkCheck.Status = LinkStatusTimeout
		linkCheck.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		linkCheck.StatusCode = resp.StatusCode

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			linkCheck.Status = LinkStatusActive
		case resp.StatusCode >= 300 && resp.StatusCode < 400:
			linkCheck.Status = LinkStatusRedirect
			if location := resp.Header.Get("Location"); location != "" {
				linkCheck.RedirectURL = location
			}
		case resp.StatusCode >= 400:
			linkCheck.Status = LinkStatusBroken
			linkCheck.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
		default:
			linkCheck.Status = LinkStatusUnknown
		}
	}

	// Save the link check result
	if err := s.db.WithContext(ctx).Create(linkCheck).Error; err != nil {
		return nil, fmt.Errorf("failed to save link check: %w", err)
	}

	// Create notification if link is broken or redirected
	if linkCheck.Status == LinkStatusBroken || linkCheck.Status == LinkStatusRedirect {
		notification := &LinkChangeNotification{
			UserID:     userID,
			BookmarkID: req.BookmarkID,
			ChangeType: string(linkCheck.Status),
			Message:    s.generateNotificationMessage(linkCheck),
		}
		s.db.WithContext(ctx).Create(notification)
	}

	return linkCheck, nil
}

// CreateMonitoringJob creates a new monitoring job
func (s *Service) CreateMonitoringJob(ctx context.Context, userID uint, req *CreateMonitoringJobRequest) (*LinkMonitoringJob, error) {
	// Validate cron expression (basic validation)
	if !s.isValidCronExpression(req.Frequency) {
		return nil, fmt.Errorf("invalid cron expression")
	}

	job := &LinkMonitoringJob{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Frequency:   req.Frequency,
	}

	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create monitoring job: %w", err)
	}

	return job, nil
}

// GetMonitoringJob retrieves a monitoring job by ID
func (s *Service) GetMonitoringJob(ctx context.Context, userID, jobID uint) (*LinkMonitoringJob, error) {
	var job LinkMonitoringJob
	if err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", jobID, userID).
		First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("monitoring job not found")
		}
		return nil, fmt.Errorf("failed to get monitoring job: %w", err)
	}
	return &job, nil
}

// UpdateMonitoringJob updates a monitoring job
func (s *Service) UpdateMonitoringJob(ctx context.Context, userID, jobID uint, req *UpdateMonitoringJobRequest) (*LinkMonitoringJob, error) {
	job, err := s.GetMonitoringJob(ctx, userID, jobID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Frequency != "" {
		if !s.isValidCronExpression(req.Frequency) {
			return nil, fmt.Errorf("invalid cron expression")
		}
		updates["frequency"] = req.Frequency
	}

	if err := s.db.WithContext(ctx).Model(job).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update monitoring job: %w", err)
	}

	return job, nil
}

// DeleteMonitoringJob deletes a monitoring job
func (s *Service) DeleteMonitoringJob(ctx context.Context, userID, jobID uint) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", jobID, userID).
		Delete(&LinkMonitoringJob{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete monitoring job: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("monitoring job not found")
	}

	return nil
}

// ListMonitoringJobs lists monitoring jobs for a user
func (s *Service) ListMonitoringJobs(ctx context.Context, userID uint, page, pageSize int) ([]*LinkMonitoringJob, int64, error) {
	var jobs []*LinkMonitoringJob
	var total int64

	// Count total
	if err := s.db.WithContext(ctx).Model(&LinkMonitoringJob{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count monitoring jobs: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list monitoring jobs: %w", err)
	}

	return jobs, total, nil
}

// GenerateMaintenanceReport generates a maintenance report for user's bookmarks
func (s *Service) GenerateMaintenanceReport(ctx context.Context, userID uint, collectionID *uint) (*LinkMaintenanceReport, error) {
	report := &LinkMaintenanceReport{
		UserID:       userID,
		CollectionID: collectionID,
		ReportType:   "comprehensive",
		GeneratedAt:  time.Now(),
	}

	// Build query based on collection filter
	query := s.db.WithContext(ctx).Table("bookmarks").
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if collectionID != nil {
		query = query.Joins("JOIN bookmark_collections ON bookmarks.id = bookmark_collections.bookmark_id").
			Where("bookmark_collections.collection_id = ?", *collectionID)
	}

	// Count total links
	var totalLinks int64
	if err := query.Count(&totalLinks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total links: %w", err)
	}
	report.TotalLinks = int(totalLinks)

	// Get recent link checks to analyze status
	var recentChecks []LinkCheck

	// Use a subquery to get the latest check for each bookmark (SQLite compatible)
	subQuery := s.db.WithContext(ctx).
		Table("link_checks").
		Select("bookmark_id, MAX(checked_at) as max_checked_at").
		Joins("JOIN bookmarks ON link_checks.bookmark_id = bookmarks.id").
		Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL", userID).
		Group("bookmark_id")

	checkQuery := s.db.WithContext(ctx).
		Table("link_checks").
		Select("link_checks.bookmark_id, link_checks.status").
		Joins("JOIN bookmarks ON link_checks.bookmark_id = bookmarks.id").
		Joins("JOIN (?) latest ON link_checks.bookmark_id = latest.bookmark_id AND link_checks.checked_at = latest.max_checked_at", subQuery).
		Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL", userID)

	if collectionID != nil {
		// Update subquery for collection filter
		subQuery = s.db.WithContext(ctx).
			Table("link_checks").
			Select("link_checks.bookmark_id, MAX(link_checks.checked_at) as max_checked_at").
			Joins("JOIN bookmarks ON link_checks.bookmark_id = bookmarks.id").
			Joins("JOIN bookmark_collections ON bookmarks.id = bookmark_collections.bookmark_id").
			Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL AND bookmark_collections.collection_id = ?", userID, *collectionID).
			Group("link_checks.bookmark_id")

		checkQuery = s.db.WithContext(ctx).
			Table("link_checks").
			Select("link_checks.bookmark_id, link_checks.status").
			Joins("JOIN bookmarks ON link_checks.bookmark_id = bookmarks.id").
			Joins("JOIN bookmark_collections ON bookmarks.id = bookmark_collections.bookmark_id").
			Joins("JOIN (?) latest ON link_checks.bookmark_id = latest.bookmark_id AND link_checks.checked_at = latest.max_checked_at", subQuery).
			Where("bookmarks.user_id = ? AND bookmarks.deleted_at IS NULL AND bookmark_collections.collection_id = ?", userID, *collectionID)
	}

	if err := checkQuery.Find(&recentChecks).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent checks: %w", err)
	}

	// Analyze status counts
	for _, check := range recentChecks {
		switch check.Status {
		case LinkStatusActive:
			report.ActiveLinks++
		case LinkStatusBroken:
			report.BrokenLinks++
		case LinkStatusRedirect:
			report.RedirectLinks++
		}
	}

	// Generate suggestions
	suggestions := s.generateMaintenanceSuggestions(report)
	suggestionsJSON, _ := json.Marshal(suggestions)
	report.Suggestions = string(suggestionsJSON)

	// Save the report
	if err := s.db.WithContext(ctx).Create(report).Error; err != nil {
		return nil, fmt.Errorf("failed to save maintenance report: %w", err)
	}

	return report, nil
}

// GetLinkChecks retrieves link checks for a bookmark
func (s *Service) GetLinkChecks(ctx context.Context, userID, bookmarkID uint, page, pageSize int) ([]*LinkCheck, int64, error) {
	// Verify bookmark belongs to user
	var bookmark struct {
		ID     uint
		UserID uint
	}

	if err := s.db.WithContext(ctx).
		Table("bookmarks").
		Select("id, user_id").
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", bookmarkID, userID).
		First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("bookmark not found")
		}
		return nil, 0, fmt.Errorf("failed to verify bookmark: %w", err)
	}

	var checks []*LinkCheck
	var total int64

	// Count total
	if err := s.db.WithContext(ctx).Model(&LinkCheck{}).
		Where("bookmark_id = ?", bookmarkID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count link checks: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).
		Where("bookmark_id = ?", bookmarkID).
		Order("checked_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&checks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list link checks: %w", err)
	}

	return checks, total, nil
}

// GetNotifications retrieves notifications for a user
func (s *Service) GetNotifications(ctx context.Context, userID uint, page, pageSize int, unreadOnly bool) ([]*LinkChangeNotification, int64, error) {
	var notifications []*LinkChangeNotification
	var total int64

	query := s.db.WithContext(ctx).Model(&LinkChangeNotification{}).
		Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("read = ?", false)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	return notifications, total, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *Service) MarkNotificationAsRead(ctx context.Context, userID, notificationID uint) error {
	result := s.db.WithContext(ctx).
		Model(&LinkChangeNotification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// Helper methods

func (s *Service) isValidCronExpression(expr string) bool {
	// Basic cron validation - should have 5 or 6 fields
	fields := strings.Fields(expr)
	return len(fields) == 5 || len(fields) == 6
}

func (s *Service) generateNotificationMessage(check *LinkCheck) string {
	switch check.Status {
	case LinkStatusBroken:
		return fmt.Sprintf("Link is broken (HTTP %d): %s", check.StatusCode, check.URL)
	case LinkStatusRedirect:
		return fmt.Sprintf("Link redirects to: %s", check.RedirectURL)
	case LinkStatusTimeout:
		return fmt.Sprintf("Link timed out: %s", check.URL)
	default:
		return fmt.Sprintf("Link status changed: %s", check.URL)
	}
}

func (s *Service) generateMaintenanceSuggestions(report *LinkMaintenanceReport) []string {
	var suggestions []string

	if report.BrokenLinks > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Fix %d broken links to improve collection health", report.BrokenLinks))
	}

	if report.RedirectLinks > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Update %d redirected links to their final destinations", report.RedirectLinks))
	}

	if report.TotalLinks > 0 {
		healthPercentage := float64(report.ActiveLinks) / float64(report.TotalLinks) * 100
		if healthPercentage < 80 {
			suggestions = append(suggestions, "Consider reviewing and cleaning up your bookmark collection")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Your bookmark collection is in good health!")
	}

	return suggestions
}
