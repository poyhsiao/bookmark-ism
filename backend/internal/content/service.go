package content

import (
	"context"
	"time"
)

// Service handles content analysis operations
type Service struct {
	analyzer ContentAnalyzer
}

// NewService creates a new content analysis service
func NewService() *Service {
	return &Service{
		analyzer: NewWebContentAnalyzer(),
	}
}

// AnalyzeURL performs comprehensive analysis of a URL
func (s *Service) AnalyzeURL(ctx context.Context, url string, userID uint) (*AnalysisResult, error) {
	// Extract content from URL
	contentData, err := s.analyzer.ExtractContent(url)
	if err != nil {
		return nil, err
	}

	// Perform content analysis
	analysis, err := s.analyzer.AnalyzeContent(contentData)
	if err != nil {
		return nil, err
	}

	// Get tag suggestions
	suggestedTags, err := s.analyzer.SuggestTags(analysis)
	if err != nil {
		return nil, err
	}

	// Categorize content
	category, err := s.analyzer.CategorizeContent(analysis)
	if err != nil {
		return nil, err
	}

	// Detect duplicates
	duplicates, err := s.analyzer.DetectDuplicates(contentData, userID)
	if err != nil {
		return nil, err
	}

	// Build result
	result := &AnalysisResult{
		URL:           contentData.URL,
		Title:         contentData.Title,
		Description:   contentData.Description,
		SuggestedTags: suggestedTags,
		Category:      category,
		Topics:        analysis.Topics,
		Sentiment:     analysis.Sentiment,
		Readability:   analysis.Readability,
		Duplicates:    duplicates,
		Entities:      analysis.Entities,
		Summary:       analysis.Summary,
		Language:      contentData.Language,
		WordCount:     contentData.WordCount,
		AnalyzedAt:    time.Now(),
	}

	return result, nil
}

// SuggestTagsForBookmark suggests tags for a specific bookmark
func (s *Service) SuggestTagsForBookmark(ctx context.Context, bookmarkID uint, url string) ([]string, error) {
	// Extract content
	contentData, err := s.analyzer.ExtractContent(url)
	if err != nil {
		return nil, err
	}

	// Analyze content
	analysis, err := s.analyzer.AnalyzeContent(contentData)
	if err != nil {
		return nil, err
	}

	// Get tag suggestions
	return s.analyzer.SuggestTags(analysis)
}

// DetectDuplicateBookmarks finds potential duplicate bookmarks
func (s *Service) DetectDuplicateBookmarks(ctx context.Context, url string, userID uint) ([]*DuplicateMatch, error) {
	// Extract content
	contentData, err := s.analyzer.ExtractContent(url)
	if err != nil {
		return nil, err
	}

	// Detect duplicates
	return s.analyzer.DetectDuplicates(contentData, userID)
}

// CategorizeBookmark categorizes a bookmark based on its content
func (s *Service) CategorizeBookmark(ctx context.Context, url string) (string, error) {
	// Extract content
	contentData, err := s.analyzer.ExtractContent(url)
	if err != nil {
		return "", err
	}

	// Analyze content
	analysis, err := s.analyzer.AnalyzeContent(contentData)
	if err != nil {
		return "", err
	}

	// Categorize content
	return s.analyzer.CategorizeContent(analysis)
}
