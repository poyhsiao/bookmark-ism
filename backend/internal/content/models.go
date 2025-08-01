package content

import "time"

// ContentData represents extracted content from a webpage
type ContentData struct {
	URL         string     `json:"url"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	Keywords    []string   `json:"keywords"`
	Language    string     `json:"language"`
	Author      string     `json:"author,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ImageURL    string     `json:"image_url,omitempty"`
	Domain      string     `json:"domain"`
	WordCount   int        `json:"word_count"`
}

// ContentAnalysis represents the analysis results of content
type ContentAnalysis struct {
	*ContentData
	Topics      []string  `json:"topics"`
	Sentiment   string    `json:"sentiment"`   // positive, negative, neutral
	Readability float64   `json:"readability"` // 0.0 to 1.0
	Complexity  string    `json:"complexity"`  // simple, medium, complex
	Entities    []Entity  `json:"entities"`
	Summary     string    `json:"summary"`
	AnalyzedAt  time.Time `json:"analyzed_at"`
}

// Entity represents a named entity found in the content
type Entity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"` // PERSON, ORGANIZATION, LOCATION, etc.
	Confidence float64 `json:"confidence"`
}

// DuplicateMatch represents a potential duplicate bookmark
type DuplicateMatch struct {
	BookmarkID  uint    `json:"bookmark_id"`
	URL         string  `json:"url"`
	Title       string  `json:"title"`
	Similarity  float64 `json:"similarity"` // 0.0 to 1.0
	MatchReason string  `json:"match_reason"`
}

// AnalysisResult represents the complete analysis result for a URL
type AnalysisResult struct {
	URL           string            `json:"url"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	SuggestedTags []string          `json:"suggested_tags"`
	Category      string            `json:"category"`
	Topics        []string          `json:"topics"`
	Sentiment     string            `json:"sentiment"`
	Readability   float64           `json:"readability"`
	Duplicates    []*DuplicateMatch `json:"duplicates"`
	Entities      []Entity          `json:"entities"`
	Summary       string            `json:"summary"`
	Language      string            `json:"language"`
	WordCount     int               `json:"word_count"`
	AnalyzedAt    time.Time         `json:"analyzed_at"`
}

// TagSuggestion represents a suggested tag with confidence score
type TagSuggestion struct {
	Tag        string  `json:"tag"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"` // content, keywords, topics, etc.
}

// CategoryPrediction represents a category prediction with confidence
type CategoryPrediction struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// ContentAnalyzer defines the interface for content analysis
type ContentAnalyzer interface {
	// ExtractContent extracts content and metadata from a URL
	ExtractContent(url string) (*ContentData, error)

	// AnalyzeContent performs comprehensive analysis on extracted content
	AnalyzeContent(content *ContentData) (*ContentAnalysis, error)

	// SuggestTags suggests relevant tags based on content analysis
	SuggestTags(analysis *ContentAnalysis) ([]string, error)

	// CategorizeContent categorizes content into predefined categories
	CategorizeContent(analysis *ContentAnalysis) (string, error)

	// DetectDuplicates finds potential duplicate bookmarks for a user
	DetectDuplicates(content *ContentData, userID uint) ([]*DuplicateMatch, error)
}

// AnalysisRequest represents a request for content analysis
type AnalysisRequest struct {
	URL    string `json:"url" binding:"required"`
	UserID uint   `json:"user_id"`
}

// TagSuggestionRequest represents a request for tag suggestions
type TagSuggestionRequest struct {
	BookmarkID uint   `json:"bookmark_id" binding:"required"`
	URL        string `json:"url" binding:"required"`
}

// DuplicateDetectionRequest represents a request for duplicate detection
type DuplicateDetectionRequest struct {
	URL    string `json:"url" binding:"required"`
	UserID uint   `json:"user_id" binding:"required"`
}

// CategoryRequest represents a request for content categorization
type CategoryRequest struct {
	URL string `json:"url" binding:"required"`
}
