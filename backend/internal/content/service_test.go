package content

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockContentAnalyzer is a mock implementation of ContentAnalyzer
type MockContentAnalyzer struct {
	mock.Mock
}

func (m *MockContentAnalyzer) ExtractContent(url string) (*ContentData, error) {
	args := m.Called(url)
	return args.Get(0).(*ContentData), args.Error(1)
}

func (m *MockContentAnalyzer) AnalyzeContent(content *ContentData) (*ContentAnalysis, error) {
	args := m.Called(content)
	return args.Get(0).(*ContentAnalysis), args.Error(1)
}

func (m *MockContentAnalyzer) SuggestTags(analysis *ContentAnalysis) ([]string, error) {
	args := m.Called(analysis)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockContentAnalyzer) CategorizeContent(analysis *ContentAnalysis) (string, error) {
	args := m.Called(analysis)
	return args.String(0), args.Error(1)
}

func (m *MockContentAnalyzer) DetectDuplicates(content *ContentData, userID uint) ([]*DuplicateMatch, error) {
	args := m.Called(content, userID)
	return args.Get(0).([]*DuplicateMatch), args.Error(1)
}

// ContentServiceTestSuite defines the test suite for content service
type ContentServiceTestSuite struct {
	suite.Suite
	service  *Service
	analyzer *MockContentAnalyzer
}

func (suite *ContentServiceTestSuite) SetupTest() {
	suite.analyzer = new(MockContentAnalyzer)
	suite.service = &Service{
		analyzer: suite.analyzer,
	}
}

func (suite *ContentServiceTestSuite) TestAnalyzeURL() {
	ctx := context.Background()
	url := "https://example.com/article"
	userID := uint(1)

	// Mock content data
	contentData := &ContentData{
		URL:         url,
		Title:       "Example Article",
		Description: "This is an example article about technology",
		Content:     "Technology is advancing rapidly...",
		Keywords:    []string{"technology", "innovation"},
		Language:    "en",
	}

	// Mock content analysis
	analysis := &ContentAnalysis{
		ContentData: contentData,
		Topics:      []string{"technology", "innovation"},
		Sentiment:   "positive",
		Readability: 0.8,
	}

	// Mock suggested tags
	suggestedTags := []string{"tech", "innovation", "article"}

	// Mock category
	category := "Technology"

	// Mock duplicate matches
	duplicates := []*DuplicateMatch{
		{
			BookmarkID:  123,
			URL:         "https://example.com/similar",
			Title:       "Similar Article",
			Similarity:  0.85,
			MatchReason: "Similar content and keywords",
		},
	}

	// Set up expectations
	suite.analyzer.On("ExtractContent", url).Return(contentData, nil)
	suite.analyzer.On("AnalyzeContent", contentData).Return(analysis, nil)
	suite.analyzer.On("SuggestTags", analysis).Return(suggestedTags, nil)
	suite.analyzer.On("CategorizeContent", analysis).Return(category, nil)
	suite.analyzer.On("DetectDuplicates", contentData, userID).Return(duplicates, nil)

	// Execute
	result, err := suite.service.AnalyzeURL(ctx, url, userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), url, result.URL)
	assert.Equal(suite.T(), "Example Article", result.Title)
	assert.Equal(suite.T(), suggestedTags, result.SuggestedTags)
	assert.Equal(suite.T(), category, result.Category)
	assert.Len(suite.T(), result.Duplicates, 1)
	assert.Equal(suite.T(), float64(0.85), result.Duplicates[0].Similarity)

	// Verify all expectations were met
	suite.analyzer.AssertExpectations(suite.T())
}

func (suite *ContentServiceTestSuite) TestAnalyzeURLError() {
	ctx := context.Background()
	url := "https://invalid-url.com"
	userID := uint(1)

	// Set up expectation for error
	suite.analyzer.On("ExtractContent", url).Return((*ContentData)(nil), assert.AnError)

	// Execute
	result, err := suite.service.AnalyzeURL(ctx, url, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)

	// Verify expectations
	suite.analyzer.AssertExpectations(suite.T())
}

func (suite *ContentServiceTestSuite) TestSuggestTagsForBookmark() {
	ctx := context.Background()
	bookmarkID := uint(123)
	url := "https://example.com/tech-article"

	// Mock content data
	contentData := &ContentData{
		URL:         url,
		Title:       "Tech Innovation",
		Description: "Latest technology trends",
		Content:     "Artificial intelligence and machine learning...",
		Keywords:    []string{"AI", "ML", "technology"},
		Language:    "en",
	}

	// Mock analysis
	analysis := &ContentAnalysis{
		ContentData: contentData,
		Topics:      []string{"AI", "technology", "innovation"},
		Sentiment:   "positive",
		Readability: 0.9,
	}

	// Mock suggested tags
	suggestedTags := []string{"AI", "machine-learning", "tech", "innovation"}

	// Set up expectations
	suite.analyzer.On("ExtractContent", url).Return(contentData, nil)
	suite.analyzer.On("AnalyzeContent", contentData).Return(analysis, nil)
	suite.analyzer.On("SuggestTags", analysis).Return(suggestedTags, nil)

	// Execute
	tags, err := suite.service.SuggestTagsForBookmark(ctx, bookmarkID, url)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suggestedTags, tags)

	// Verify expectations
	suite.analyzer.AssertExpectations(suite.T())
}

func (suite *ContentServiceTestSuite) TestDetectDuplicateBookmarks() {
	ctx := context.Background()
	url := "https://example.com/article"
	userID := uint(1)

	// Mock content data
	contentData := &ContentData{
		URL:         url,
		Title:       "Example Article",
		Description: "Article description",
		Content:     "Article content...",
		Keywords:    []string{"example", "article"},
		Language:    "en",
	}

	// Mock duplicates
	duplicates := []*DuplicateMatch{
		{
			BookmarkID:  456,
			URL:         "https://example.com/similar-article",
			Title:       "Similar Article",
			Similarity:  0.92,
			MatchReason: "High content similarity",
		},
		{
			BookmarkID:  789,
			URL:         "https://another.com/same-topic",
			Title:       "Same Topic Article",
			Similarity:  0.78,
			MatchReason: "Similar keywords and topics",
		},
	}

	// Set up expectations
	suite.analyzer.On("ExtractContent", url).Return(contentData, nil)
	suite.analyzer.On("DetectDuplicates", contentData, userID).Return(duplicates, nil)

	// Execute
	result, err := suite.service.DetectDuplicateBookmarks(ctx, url, userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), uint(456), result[0].BookmarkID)
	assert.Equal(suite.T(), float64(0.92), result[0].Similarity)

	// Verify expectations
	suite.analyzer.AssertExpectations(suite.T())
}

func (suite *ContentServiceTestSuite) TestCategorizeBookmark() {
	ctx := context.Background()
	url := "https://techcrunch.com/startup-news"

	// Mock content data
	contentData := &ContentData{
		URL:         url,
		Title:       "Startup Raises $10M",
		Description: "Tech startup secures funding",
		Content:     "A technology startup has raised $10 million...",
		Keywords:    []string{"startup", "funding", "technology"},
		Language:    "en",
	}

	// Mock analysis
	analysis := &ContentAnalysis{
		ContentData: contentData,
		Topics:      []string{"startup", "funding", "business"},
		Sentiment:   "positive",
		Readability: 0.7,
	}

	// Mock category
	category := "Business"

	// Set up expectations
	suite.analyzer.On("ExtractContent", url).Return(contentData, nil)
	suite.analyzer.On("AnalyzeContent", contentData).Return(analysis, nil)
	suite.analyzer.On("CategorizeContent", analysis).Return(category, nil)

	// Execute
	result, err := suite.service.CategorizeBookmark(ctx, url)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), category, result)

	// Verify expectations
	suite.analyzer.AssertExpectations(suite.T())
}

func TestContentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ContentServiceTestSuite))
}

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.analyzer)
}
