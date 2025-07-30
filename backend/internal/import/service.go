package import_export

import (
	"bookmark-sync-service/backend/pkg/database"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Service provides import/export functionality
type Service struct {
	db *gorm.DB
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	ImportedBookmarksCount   int      `json:"imported_bookmarks_count"`
	ImportedCollectionsCount int      `json:"imported_collections_count"`
	DuplicatesSkipped        int      `json:"duplicates_skipped"`
	Errors                   []string `json:"errors"`
	ProcessingTimeMs         int64    `json:"processing_time_ms"`
}

// ImportProgress represents the progress of an import operation
type ImportProgress struct {
	JobID          string    `json:"job_id"`
	UserID         uint      `json:"user_id"`
	Status         string    `json:"status"` // pending, processing, completed, failed
	TotalItems     int       `json:"total_items"`
	ProcessedItems int       `json:"processed_items"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ChromeBookmark represents Chrome bookmark format
type ChromeBookmark struct {
	DateAdded    string           `json:"date_added"`
	GUID         string           `json:"guid"`
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Type         string           `json:"type"`
	URL          string           `json:"url,omitempty"`
	Children     []ChromeBookmark `json:"children,omitempty"`
	DateModified string           `json:"date_modified,omitempty"`
}

// ChromeBookmarkFile represents Chrome bookmark file structure
type ChromeBookmarkFile struct {
	Checksum string `json:"checksum"`
	Roots    struct {
		BookmarkBar ChromeBookmark `json:"bookmark_bar"`
		Other       ChromeBookmark `json:"other"`
		Synced      ChromeBookmark `json:"synced"`
	} `json:"roots"`
	Version int `json:"version"`
}

// NewService creates a new import/export service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// ImportBookmarksFromChrome imports bookmarks from Chrome format
func (s *Service) ImportBookmarksFromChrome(ctx context.Context, userID uint, reader io.Reader) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Parse Chrome bookmarks JSON
	var chromeFile ChromeBookmarkFile
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&chromeFile); err != nil {
		return nil, fmt.Errorf("failed to parse Chrome bookmarks: %w", err)
	}

	// Process bookmark bar
	if err := s.processChromeFolderRecursive(ctx, userID, &chromeFile.Roots.BookmarkBar, nil, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error processing bookmark bar: %v", err))
	}

	// Process other bookmarks
	if err := s.processChromeFolderRecursive(ctx, userID, &chromeFile.Roots.Other, nil, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error processing other bookmarks: %v", err))
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// processChromeFolderRecursive processes Chrome bookmark folders recursively
func (s *Service) processChromeFolderRecursive(ctx context.Context, userID uint, folder *ChromeBookmark, parentCollection *database.Collection, result *ImportResult) error {
	var currentCollection *database.Collection

	// Create collection for folder (except root folders)
	if folder.Type == "folder" && folder.Name != "Bookmarks bar" && folder.Name != "Other bookmarks" {
		collection := &database.Collection{
			UserID:      userID,
			Name:        folder.Name,
			Description: fmt.Sprintf("Imported from Chrome on %s", time.Now().Format("2006-01-02")),
			Visibility:  "private",
		}

		if parentCollection != nil {
			collection.ParentID = &parentCollection.ID
		}

		if err := s.db.Create(collection).Error; err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}

		currentCollection = collection
		result.ImportedCollectionsCount++
	}

	// Process children
	for _, child := range folder.Children {
		if child.Type == "url" {
			// Check for duplicates
			isDuplicate, err := s.DetectDuplicate(ctx, userID, child.URL)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking duplicate for %s: %v", child.URL, err))
				continue
			}

			if isDuplicate {
				result.DuplicatesSkipped++
				continue
			}

			// Create bookmark
			bookmark := &database.Bookmark{
				UserID:      userID,
				URL:         child.URL,
				Title:       child.Name,
				Description: fmt.Sprintf("Imported from Chrome on %s", time.Now().Format("2006-01-02")),
				Status:      "active",
			}

			if err := s.db.Create(bookmark).Error; err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create bookmark %s: %v", child.Name, err))
				continue
			}

			// Associate with collection if exists
			if currentCollection != nil {
				if err := s.db.Model(currentCollection).Association("Bookmarks").Append(bookmark); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to associate bookmark with collection: %v", err))
				}
			}

			result.ImportedBookmarksCount++
		} else if child.Type == "folder" {
			// Recursively process subfolder
			if err := s.processChromeFolderRecursive(ctx, userID, &child, currentCollection, result); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error processing subfolder %s: %v", child.Name, err))
			}
		}
	}

	return nil
}

// ImportBookmarksFromFirefox imports bookmarks from Firefox HTML format
func (s *Service) ImportBookmarksFromFirefox(ctx context.Context, userID uint, reader io.Reader) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Read HTML content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Firefox bookmarks: %w", err)
	}

	htmlContent := string(content)

	// Parse Firefox HTML bookmarks (simplified parser)
	if err := s.parseFirefoxHTML(ctx, userID, htmlContent, result); err != nil {
		return nil, fmt.Errorf("failed to parse Firefox HTML: %w", err)
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// parseFirefoxHTML parses Firefox bookmark HTML format
func (s *Service) parseFirefoxHTML(ctx context.Context, userID uint, htmlContent string, result *ImportResult) error {
	lines := strings.Split(htmlContent, "\n")
	var currentCollection *database.Collection

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse folder
		if strings.Contains(line, "<H3") && strings.Contains(line, ">") && strings.Contains(line, "</H3>") {
			// Find the last ">" before "</H3>" to handle HTML attributes
			startIdx := strings.LastIndex(line[:strings.Index(line, "</H3>")], ">")
			endIdx := strings.Index(line, "</H3>")
			var folderName string
			if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
				folderName = strings.TrimSpace(line[startIdx+1 : endIdx])
			}
			if folderName != "" && folderName != "Bookmarks Menu" {
				collection := &database.Collection{
					UserID:      userID,
					Name:        folderName,
					Description: fmt.Sprintf("Imported from Firefox on %s", time.Now().Format("2006-01-02")),
					Visibility:  "private",
				}

				if err := s.db.Create(collection).Error; err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to create collection %s: %v", folderName, err))
					continue
				}

				currentCollection = collection
				result.ImportedCollectionsCount++
			}
		}

		// Parse bookmark
		if strings.Contains(line, "<DT><A HREF=") {
			url := extractTextBetween(line, "HREF=\"", "\"")
			title := extractTextBetween(line, "\">", "</A>")

			if url != "" && title != "" {
				// Check for duplicates
				isDuplicate, err := s.DetectDuplicate(ctx, userID, url)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Error checking duplicate for %s: %v", url, err))
					continue
				}

				if isDuplicate {
					result.DuplicatesSkipped++
					continue
				}

				// Create bookmark
				bookmark := &database.Bookmark{
					UserID:      userID,
					URL:         url,
					Title:       title,
					Description: fmt.Sprintf("Imported from Firefox on %s", time.Now().Format("2006-01-02")),
					Status:      "active",
				}

				if err := s.db.Create(bookmark).Error; err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to create bookmark %s: %v", title, err))
					continue
				}

				// Associate with collection if exists
				if currentCollection != nil {
					if err := s.db.Model(currentCollection).Association("Bookmarks").Append(bookmark); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to associate bookmark with collection: %v", err))
					}
				}

				result.ImportedBookmarksCount++
			}
		}
	}

	return nil
}

// extractTextBetween extracts text between two delimiters
func extractTextBetween(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)

	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return ""
	}

	result := text[startIdx : startIdx+endIdx]

	// Clean up HTML content - remove any HTML tags or attributes
	result = strings.TrimSpace(result)

	return result
}

// ImportBookmarksFromSafari imports bookmarks from Safari plist format
func (s *Service) ImportBookmarksFromSafari(ctx context.Context, userID uint, reader io.Reader) (*ImportResult, error) {
	startTime := time.Now()
	result := &ImportResult{
		Errors: make([]string, 0),
	}

	// Read plist content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read Safari bookmarks: %w", err)
	}

	plistContent := string(content)

	// Parse Safari plist bookmarks (simplified parser)
	if err := s.parseSafariPlist(ctx, userID, plistContent, result); err != nil {
		return nil, fmt.Errorf("failed to parse Safari plist: %w", err)
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// parseSafariPlist parses Safari bookmark plist format (simplified)
func (s *Service) parseSafariPlist(ctx context.Context, userID uint, plistContent string, result *ImportResult) error {
	lines := strings.Split(plistContent, "\n")
	var currentCollection *database.Collection
	var currentURL, currentTitle string

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Parse folder title
		if strings.Contains(line, "<key>Title</key>") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.Contains(nextLine, "<string>") && !strings.Contains(nextLine, "BookmarksBar") {
				folderName := extractTextBetween(nextLine, "<string>", "</string>")
				if folderName != "" && folderName != "BookmarksBar" {
					collection := &database.Collection{
						UserID:      userID,
						Name:        folderName,
						Description: fmt.Sprintf("Imported from Safari on %s", time.Now().Format("2006-01-02")),
						Visibility:  "private",
					}

					if err := s.db.Create(collection).Error; err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Failed to create collection %s: %v", folderName, err))
						continue
					}

					currentCollection = collection
					result.ImportedCollectionsCount++
				}
			}
		}

		// Parse URL
		if strings.Contains(line, "<key>URLString</key>") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.Contains(nextLine, "<string>") {
				currentURL = extractTextBetween(nextLine, "<string>", "</string>")
			}
		}

		// Parse title
		if strings.Contains(line, "<key>title</key>") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.Contains(nextLine, "<string>") {
				currentTitle = extractTextBetween(nextLine, "<string>", "</string>")
			}
		}

		// Create bookmark when we have both URL and title
		if currentURL != "" && currentTitle != "" {
			// Check for duplicates
			isDuplicate, err := s.DetectDuplicate(ctx, userID, currentURL)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking duplicate for %s: %v", currentURL, err))
			} else if isDuplicate {
				result.DuplicatesSkipped++
			} else {
				// Create bookmark
				bookmark := &database.Bookmark{
					UserID:      userID,
					URL:         currentURL,
					Title:       currentTitle,
					Description: fmt.Sprintf("Imported from Safari on %s", time.Now().Format("2006-01-02")),
					Status:      "active",
				}

				if err := s.db.Create(bookmark).Error; err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Failed to create bookmark %s: %v", currentTitle, err))
				} else {
					// Associate with collection if exists
					if currentCollection != nil {
						if err := s.db.Model(currentCollection).Association("Bookmarks").Append(bookmark); err != nil {
							result.Errors = append(result.Errors, fmt.Sprintf("Failed to associate bookmark with collection: %v", err))
						}
					}
					result.ImportedBookmarksCount++
				}
			}

			// Reset for next bookmark
			currentURL = ""
			currentTitle = ""
		}
	}

	return nil
}

// ExportBookmarksToJSON exports bookmarks to JSON format
func (s *Service) ExportBookmarksToJSON(ctx context.Context, userID uint, writer io.Writer) error {
	// Get all bookmarks for user
	var bookmarks []database.Bookmark
	if err := s.db.Where("user_id = ?", userID).Preload("Collections").Find(&bookmarks).Error; err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	// Get all collections for user
	var collections []database.Collection
	if err := s.db.Where("user_id = ?", userID).Preload("Bookmarks").Find(&collections).Error; err != nil {
		return fmt.Errorf("failed to fetch collections: %w", err)
	}

	// Create export data structure
	exportData := map[string]interface{}{
		"version":     "1.0",
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"user_id":     userID,
		"bookmarks":   bookmarks,
		"collections": collections,
		"metadata": map[string]interface{}{
			"total_bookmarks":   len(bookmarks),
			"total_collections": len(collections),
			"export_format":     "json",
		},
	}

	// Encode to JSON
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ExportBookmarksToHTML exports bookmarks to HTML format (Netscape format)
func (s *Service) ExportBookmarksToHTML(ctx context.Context, userID uint, writer io.Writer) error {
	// Get all bookmarks for user
	var bookmarks []database.Bookmark
	if err := s.db.Where("user_id = ?", userID).Find(&bookmarks).Error; err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	// Get all collections for user
	var collections []database.Collection
	if err := s.db.Where("user_id = ?", userID).Preload("Bookmarks").Find(&collections).Error; err != nil {
		return fmt.Errorf("failed to fetch collections: %w", err)
	}

	// Write HTML header
	fmt.Fprintf(writer, `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks Menu</H1>
<DL><p>
`)

	// Write collections with their bookmarks
	for _, collection := range collections {
		addDate := collection.CreatedAt.Unix()
		lastModified := collection.UpdatedAt.Unix()

		fmt.Fprintf(writer, `    <DT><H3 ADD_DATE="%d" LAST_MODIFIED="%d">%s</H3>
    <DL><p>
`, addDate, lastModified, escapeHTML(collection.Name))

		for _, bookmark := range collection.Bookmarks {
			bookmarkAddDate := bookmark.CreatedAt.Unix()
			fmt.Fprintf(writer, `        <DT><A HREF="%s" ADD_DATE="%d">%s</A>
`, escapeHTML(bookmark.URL), bookmarkAddDate, escapeHTML(bookmark.Title))
		}

		fmt.Fprintf(writer, `    </DL><p>
`)
	}

	// Write uncategorized bookmarks
	uncategorizedBookmarks := make([]database.Bookmark, 0)
	for _, bookmark := range bookmarks {
		if len(bookmark.Collections) == 0 {
			uncategorizedBookmarks = append(uncategorizedBookmarks, bookmark)
		}
	}

	for _, bookmark := range uncategorizedBookmarks {
		addDate := bookmark.CreatedAt.Unix()
		fmt.Fprintf(writer, `    <DT><A HREF="%s" ADD_DATE="%d">%s</A>
`, escapeHTML(bookmark.URL), addDate, escapeHTML(bookmark.Title))
	}

	// Write HTML footer
	fmt.Fprintf(writer, `</DL><p>
`)

	return nil
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// DetectDuplicate checks if a bookmark URL already exists for the user
func (s *Service) DetectDuplicate(ctx context.Context, userID uint, bookmarkURL string) (bool, error) {
	// Normalize URL
	parsedURL, err := url.Parse(bookmarkURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}
	normalizedURL := parsedURL.String()

	var count int64
	if err := s.db.Model(&database.Bookmark{}).
		Where("user_id = ? AND url = ?", userID, normalizedURL).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return count > 0, nil
}

// CreateImportJob creates a new import job for progress tracking
func (s *Service) CreateImportJob(ctx context.Context, userID uint, jobID string, totalItems int) error {
	// For simplicity, we'll store import progress in Redis or memory
	// In a real implementation, you might want to use a dedicated table
	// For now, we'll simulate this functionality
	return nil
}

// UpdateImportProgress updates the progress of an import job
func (s *Service) UpdateImportProgress(ctx context.Context, userID uint, jobID string, processedItems int, status string) error {
	// For simplicity, we'll store import progress in Redis or memory
	// In a real implementation, you might want to use a dedicated table
	// For now, we'll simulate this functionality
	return nil
}

// GetImportProgress gets the progress of an import job
func (s *Service) GetImportProgress(ctx context.Context, userID uint, jobID string) (*ImportProgress, error) {
	// For simplicity, we'll return nil for non-existent jobs
	// In a real implementation, you would fetch from Redis or database
	// For testing, we'll simulate that jobs don't exist initially
	return nil, nil
}

// Validate validates the import result
func (r *ImportResult) Validate() error {
	if r.ImportedBookmarksCount < 0 {
		return fmt.Errorf("imported bookmarks count cannot be negative")
	}
	if r.ImportedCollectionsCount < 0 {
		return fmt.Errorf("imported collections count cannot be negative")
	}
	if r.DuplicatesSkipped < 0 {
		return fmt.Errorf("duplicates skipped count cannot be negative")
	}
	return nil
}
