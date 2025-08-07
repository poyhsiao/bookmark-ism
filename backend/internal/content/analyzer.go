package content

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// WebContentAnalyzer implements ContentAnalyzer for web content
type WebContentAnalyzer struct {
	httpClient *http.Client
}

// NewWebContentAnalyzer creates a new web content analyzer
func NewWebContentAnalyzer() *WebContentAnalyzer {
	return &WebContentAnalyzer{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExtractContent extracts content and metadata from a URL
func (a *WebContentAnalyzer) ExtractContent(urlStr string) (*ContentData, error) {
	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Fetch the webpage
	resp, err := a.httpClient.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract content
	content := &ContentData{
		URL:    urlStr,
		Domain: parsedURL.Host,
	}

	// Extract title
	content.Title = doc.Find("title").First().Text()
	if content.Title == "" {
		content.Title = doc.Find("h1").First().Text()
	}
	content.Title = strings.TrimSpace(content.Title)

	// Extract description
	content.Description = doc.Find("meta[name='description']").AttrOr("content", "")
	if content.Description == "" {
		content.Description = doc.Find("meta[property='og:description']").AttrOr("content", "")
	}
	content.Description = strings.TrimSpace(content.Description)

	// Extract keywords
	keywordsStr := doc.Find("meta[name='keywords']").AttrOr("content", "")
	if keywordsStr != "" {
		keywords := strings.Split(keywordsStr, ",")
		for i, keyword := range keywords {
			keywords[i] = strings.TrimSpace(keyword)
		}
		content.Keywords = keywords
	}

	// Extract main content
	content.Content = a.extractMainContent(doc)
	content.WordCount = len(strings.Fields(content.Content))

	// Extract language
	content.Language = doc.Find("html").AttrOr("lang", "en")
	if content.Language == "" {
		content.Language = doc.Find("meta[http-equiv='content-language']").AttrOr("content", "en")
	}

	// Extract author
	content.Author = doc.Find("meta[name='author']").AttrOr("content", "")
	if content.Author == "" {
		content.Author = doc.Find("meta[property='article:author']").AttrOr("content", "")
	}

	// Extract image URL
	content.ImageURL = doc.Find("meta[property='og:image']").AttrOr("content", "")
	if content.ImageURL == "" {
		content.ImageURL = doc.Find("meta[name='twitter:image']").AttrOr("content", "")
	}

	// Extract published date
	publishedStr := doc.Find("meta[property='article:published_time']").AttrOr("content", "")
	if publishedStr != "" {
		if publishedTime, err := time.Parse(time.RFC3339, publishedStr); err == nil {
			content.PublishedAt = &publishedTime
		}
	}

	return content, nil
}

// extractMainContent extracts the main content from the document
func (a *WebContentAnalyzer) extractMainContent(doc *goquery.Document) string {
	var content strings.Builder

	// Try to find main content areas
	selectors := []string{
		"main",
		"article",
		"[role='main']",
		".content",
		".post-content",
		".entry-content",
		".article-content",
		"#content",
	}

	for _, selector := range selectors {
		if element := doc.Find(selector).First(); element.Length() > 0 {
			element.Find("p, h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if text != "" {
					content.WriteString(text)
					content.WriteString(" ")
				}
			})
			break
		}
	}

	// Fallback: extract all paragraphs
	if content.Len() == 0 {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" && len(text) > 50 { // Only include substantial paragraphs
				content.WriteString(text)
				content.WriteString(" ")
			}
		})
	}

	return strings.TrimSpace(content.String())
}

// AnalyzeContent performs comprehensive analysis on extracted content
func (a *WebContentAnalyzer) AnalyzeContent(content *ContentData) (*ContentAnalysis, error) {
	analysis := &ContentAnalysis{
		ContentData: content,
		AnalyzedAt:  time.Now(),
	}

	// Extract topics from content
	analysis.Topics = a.extractTopics(content)

	// Analyze sentiment (basic implementation)
	analysis.Sentiment = a.analyzeSentiment(content.Content)

	// Calculate readability score
	analysis.Readability = a.calculateReadability(content.Content)

	// Determine complexity
	analysis.Complexity = a.determineComplexity(analysis.Readability, content.WordCount)

	// Extract entities (basic implementation)
	analysis.Entities = a.extractEntities(content.Content)

	// Generate summary
	analysis.Summary = a.generateSummary(content.Content)

	return analysis, nil
}

// extractTopics extracts topics from content using keyword frequency
func (a *WebContentAnalyzer) extractTopics(content *ContentData) []string {
	text := strings.ToLower(content.Title + " " + content.Description + " " + content.Content)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "them": true,
	}

	// Extract words and count frequency
	words := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`).FindAllString(text, -1)
	wordCount := make(map[string]int)

	for _, word := range words {
		word = strings.ToLower(word)
		if !stopWords[word] {
			wordCount[word]++
		}
	}

	// Sort by frequency and take top topics
	type wordFreq struct {
		word  string
		count int
	}

	var frequencies []wordFreq
	for word, count := range wordCount {
		if count >= 2 { // Only include words that appear at least twice
			frequencies = append(frequencies, wordFreq{word, count})
		}
	}

	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].count > frequencies[j].count
	})

	// Return top 10 topics
	var topics []string
	limit := 10
	if len(frequencies) < limit {
		limit = len(frequencies)
	}

	for i := 0; i < limit; i++ {
		topics = append(topics, frequencies[i].word)
	}

	return topics
}

// analyzeSentiment performs basic sentiment analysis
func (a *WebContentAnalyzer) analyzeSentiment(content string) string {
	content = strings.ToLower(content)

	positiveWords := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic", "awesome",
		"love", "like", "enjoy", "happy", "pleased", "satisfied", "success", "win",
		"best", "better", "improve", "positive", "benefit", "advantage", "helpful",
	}

	negativeWords := []string{
		"bad", "terrible", "awful", "horrible", "hate", "dislike", "angry", "sad",
		"disappointed", "frustrated", "problem", "issue", "fail", "failure", "worst",
		"worse", "negative", "disadvantage", "difficult", "hard", "challenging",
	}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(content, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(content, word)
	}

	if positiveCount > negativeCount {
		return "positive"
	} else if negativeCount > positiveCount {
		return "negative"
	}
	return "neutral"
}

// calculateReadability calculates a basic readability score
func (a *WebContentAnalyzer) calculateReadability(content string) float64 {
	if content == "" {
		return 0.0
	}

	sentences := strings.Split(content, ".")
	words := strings.Fields(content)

	if len(sentences) == 0 || len(words) == 0 {
		return 0.0
	}

	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))

	// Simple readability score: lower average words per sentence = higher readability
	// Scale from 0.0 (hard) to 1.0 (easy)
	if avgWordsPerSentence <= 10 {
		return 1.0
	} else if avgWordsPerSentence >= 30 {
		return 0.0
	} else {
		return 1.0 - ((avgWordsPerSentence - 10) / 20)
	}
}

// determineComplexity determines content complexity based on readability and word count
func (a *WebContentAnalyzer) determineComplexity(readability float64, wordCount int) string {
	if readability >= 0.7 && wordCount < 500 {
		return "simple"
	} else if readability >= 0.4 && wordCount < 1500 {
		return "medium"
	}
	return "complex"
}

// extractEntities extracts basic named entities
func (a *WebContentAnalyzer) extractEntities(content string) []Entity {
	var entities []Entity

	// Simple pattern matching for entities
	// This is a basic implementation - in production, you'd use NLP libraries

	// Find capitalized words (potential proper nouns)
	capitalizedWords := regexp.MustCompile(`\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b`).FindAllString(content, -1)

	entityCount := make(map[string]int)
	for _, word := range capitalizedWords {
		if len(word) > 2 { // Ignore short words
			entityCount[word]++
		}
	}

	// Convert to entities with confidence based on frequency
	for entity, count := range entityCount {
		if count >= 2 { // Only include entities mentioned multiple times
			confidence := float64(count) / 10.0 // Simple confidence calculation
			if confidence > 1.0 {
				confidence = 1.0
			}

			entities = append(entities, Entity{
				Text:       entity,
				Type:       "UNKNOWN", // Would be determined by NLP in production
				Confidence: confidence,
			})
		}
	}

	return entities
}

// generateSummary generates a basic summary of the content
func (a *WebContentAnalyzer) generateSummary(content string) string {
	if content == "" {
		return ""
	}

	sentences := strings.Split(content, ".")
	if len(sentences) <= 2 {
		return content
	}

	// Take first two sentences as summary
	summary := strings.TrimSpace(sentences[0])
	if len(sentences) > 1 {
		summary += ". " + strings.TrimSpace(sentences[1])
	}

	// Limit summary length
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	return summary
}

// SuggestTags suggests relevant tags based on content analysis
func (a *WebContentAnalyzer) SuggestTags(analysis *ContentAnalysis) ([]string, error) {
	var tags []string

	// Add domain-based tag
	if analysis.Domain != "" {
		domainParts := strings.Split(analysis.Domain, ".")
		if len(domainParts) >= 2 {
			tags = append(tags, domainParts[len(domainParts)-2])
		}
	}

	// Add top topics as tags
	for i, topic := range analysis.Topics {
		if i >= 5 { // Limit to top 5 topics
			break
		}
		tags = append(tags, topic)
	}

	// Add existing keywords
	tags = append(tags, analysis.Keywords...)

	// Add category-based tags
	category, _ := a.CategorizeContent(analysis)
	if category != "" {
		tags = append(tags, strings.ToLower(category))
	}

	// Remove duplicates and clean tags
	tagSet := make(map[string]bool)
	var uniqueTags []string

	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" && len(tag) > 2 && !tagSet[tag] {
			tagSet[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}

	return uniqueTags, nil
}

// CategorizeContent categorizes content into predefined categories
func (a *WebContentAnalyzer) CategorizeContent(analysis *ContentAnalysis) (string, error) {
	text := strings.ToLower(analysis.Title + " " + analysis.Description + " " + strings.Join(analysis.Topics, " "))

	// Define category keywords
	categories := map[string][]string{
		"Technology":    {"tech", "technology", "software", "programming", "code", "developer", "computer", "digital", "ai", "machine", "learning", "data", "algorithm"},
		"Business":      {"business", "startup", "company", "entrepreneur", "finance", "money", "investment", "market", "economy", "sales", "revenue"},
		"Science":       {"science", "research", "study", "experiment", "discovery", "scientific", "biology", "chemistry", "physics", "medicine"},
		"News":          {"news", "breaking", "report", "journalist", "media", "press", "current", "events", "politics", "government"},
		"Education":     {"education", "learning", "school", "university", "course", "tutorial", "teach", "student", "academic", "knowledge"},
		"Health":        {"health", "medical", "doctor", "hospital", "treatment", "disease", "wellness", "fitness", "nutrition", "exercise"},
		"Sports":        {"sports", "game", "team", "player", "match", "competition", "athletic", "football", "basketball", "soccer"},
		"Entertainment": {"entertainment", "movie", "film", "music", "celebrity", "show", "tv", "series", "actor", "artist", "culture"},
		"Travel":        {"travel", "trip", "vacation", "destination", "hotel", "flight", "tourism", "adventure", "explore", "journey"},
		"Food":          {"food", "recipe", "cooking", "restaurant", "chef", "cuisine", "meal", "ingredient", "dish", "nutrition"},
	}

	// Score each category
	categoryScores := make(map[string]int)
	for category, keywords := range categories {
		score := 0
		for _, keyword := range keywords {
			score += strings.Count(text, keyword)
		}
		if score > 0 {
			categoryScores[category] = score
		}
	}

	// Find category with highest score
	maxScore := 0
	bestCategory := "General"

	for category, score := range categoryScores {
		if score > maxScore {
			maxScore = score
			bestCategory = category
		}
	}

	return bestCategory, nil
}

// DetectDuplicates finds potential duplicate bookmarks for a user
func (a *WebContentAnalyzer) DetectDuplicates(content *ContentData, userID uint) ([]*DuplicateMatch, error) {
	// This is a simplified implementation
	// In production, you would query the database for user's bookmarks and compare

	var duplicates []*DuplicateMatch

	// For now, return mock duplicates based on domain similarity
	if strings.Contains(content.Domain, "github.com") {
		duplicates = append(duplicates, &DuplicateMatch{
			BookmarkID:  123,
			URL:         "https://github.com/similar-repo",
			Title:       "Similar GitHub Repository",
			Similarity:  0.75,
			MatchReason: "Same domain and similar content",
		})
	}

	return duplicates, nil
}
