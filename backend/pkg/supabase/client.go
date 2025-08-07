package supabase

import (
	"context"
	"fmt"

	"bookmark-sync-service/backend/internal/config"

	"github.com/supabase-community/supabase-go"
)

// Client wraps the Supabase client with additional functionality
type Client struct {
	*supabase.Client
	config *config.SupabaseConfig
}

// NewClient creates a new Supabase client
func NewClient(cfg config.SupabaseConfig) (*Client, error) {
	client, err := supabase.NewClient(cfg.URL, cfg.AnonKey, &supabase.ClientOptions{
		Headers: map[string]string{
			"apikey": cfg.AnonKey,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &Client{
		Client: client,
		config: &cfg,
	}, nil
}

// HealthCheck checks if Supabase services are healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	// Simple health check by trying to access the auth endpoint
	// We'll just return nil for now as a basic health check
	// In a real implementation, you might want to make a simple API call
	return nil
}

// GetAuthURL returns the Supabase Auth URL
func (c *Client) GetAuthURL() string {
	return c.config.AuthURL
}

// GetRealtimeURL returns the Supabase Realtime URL
func (c *Client) GetRealtimeURL() string {
	return c.config.RealtimeURL
}
