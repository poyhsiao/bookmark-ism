package screenshot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// StorageService defines the interface for storage operations
type StorageService interface {
	StoreScreenshot(ctx context.Context, bookmarkID string, data []byte) (string, error)
}

// Service provides screenshot capture functionality
type Service struct {
	storageService StorageService
	httpClient     *http.Client
}

// NewService creates a new screenshot service
func NewService(storageService StorageService) *Service {
	return &Service{
		storageService: storageService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CaptureOptions defines options for screenshot capture
type CaptureOptions struct {
	Width     int
	Height    int
	Quality   int
	Format    string
	Thumbnail bool
}

// CaptureResult represents the result of a screenshot capture
type CaptureResult struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Size         int64  `json:"size"`
	Format       string `json:"format"`
}

// CaptureScreenshot captures a screenshot of a webpage
func (s *Service) CaptureScreenshot(ctx context.Context, bookmarkID, pageURL string, opts CaptureOptions) (*CaptureResult, error) {
	// Validate URL
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// For now, we'll simulate screenshot capture by creating a placeholder
	// In a real implementation, this would use a headless browser like Puppeteer or Playwright
	screenshotData, err := s.generatePlaceholderScreenshot(parsedURL.String(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate screenshot: %w", err)
	}

	// Store screenshot using storage service
	screenshotURL, err := s.storageService.StoreScreenshot(ctx, bookmarkID, screenshotData)
	if err != nil {
		return nil, fmt.Errorf("failed to store screenshot: %w", err)
	}

	result := &CaptureResult{
		URL:    screenshotURL,
		Width:  opts.Width,
		Height: opts.Height,
		Size:   int64(len(screenshotData)),
		Format: opts.Format,
	}

	// Generate thumbnail if requested
	if opts.Thumbnail {
		thumbnailData, err := s.generateThumbnail(screenshotData, 300, 200)
		if err == nil {
			thumbnailURL, err := s.storageService.StoreScreenshot(ctx, bookmarkID+"_thumb", thumbnailData)
			if err == nil {
				result.ThumbnailURL = thumbnailURL
			}
		}
	}

	return result, nil
}

// generatePlaceholderScreenshot generates a placeholder screenshot
// In production, this would be replaced with actual browser automation
func (s *Service) generatePlaceholderScreenshot(pageURL string, opts CaptureOptions) ([]byte, error) {
	// Create a simple placeholder image
	// This is a minimal PNG image (1x1 pixel)
	placeholderPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // Width: 1
		0x00, 0x00, 0x00, 0x01, // Height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // Bit depth, color type, compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT chunk length
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01,
		0xE2, 0x21, 0xBC, 0x33, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND chunk length
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // CRC
	}

	return placeholderPNG, nil
}

// generateThumbnail generates a thumbnail from screenshot data
func (s *Service) generateThumbnail(screenshotData []byte, width, height int) ([]byte, error) {
	// For now, return the same placeholder
	// In production, this would resize the actual screenshot
	return screenshotData, nil
}

// CaptureFromURL captures a screenshot from a URL using external service
func (s *Service) CaptureFromURL(ctx context.Context, pageURL string) ([]byte, error) {
	// This would integrate with a screenshot service like:
	// - Puppeteer/Playwright
	// - Screenshot API service
	// - Headless Chrome

	// For now, return placeholder
	return s.generatePlaceholderScreenshot(pageURL, CaptureOptions{
		Width:  1200,
		Height: 800,
		Format: "png",
	})
}

// GetFavicon attempts to get the favicon for a URL
func (s *Service) GetFavicon(ctx context.Context, pageURL string) ([]byte, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Try common favicon locations
	faviconURLs := []string{
		fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host),
		fmt.Sprintf("%s://%s/favicon.png", parsedURL.Scheme, parsedURL.Host),
		fmt.Sprintf("%s://%s/apple-touch-icon.png", parsedURL.Scheme, parsedURL.Host),
	}

	for _, faviconURL := range faviconURLs {
		req, err := http.NewRequestWithContext(ctx, "GET", faviconURL, nil)
		if err != nil {
			continue
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			if err == nil && len(data) > 0 {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("favicon not found for %s", pageURL)
}

// UpdateBookmarkScreenshot updates the screenshot for an existing bookmark
func (s *Service) UpdateBookmarkScreenshot(ctx context.Context, bookmarkID, pageURL string) (*CaptureResult, error) {
	opts := CaptureOptions{
		Width:     1200,
		Height:    800,
		Quality:   85,
		Format:    "jpeg",
		Thumbnail: true,
	}

	return s.CaptureScreenshot(ctx, bookmarkID, pageURL, opts)
}
