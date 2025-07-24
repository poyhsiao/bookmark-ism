// Package supabase provides integration with Supabase services
package supabase

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"bookmark-sync-service/backend/internal/config"

	"github.com/supabase-community/supabase-go"
)

// Client wraps the Supabase client with additional functionality
// 提供與 Supabase 服務的整合功能，包括認證、即時通訊和資料庫操作
type Client struct {
	client *supabase.Client
	config *config.SupabaseConfig
}

// NewClient creates a new Supabase client with the provided configuration
// 使用提供的配置創建新的 Supabase 客戶端
func NewClient(cfg config.SupabaseConfig) (*Client, error) {
	client, err := supabase.NewClient(cfg.URL, cfg.AnonKey, &supabase.ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &Client{
		client: client,
		config: &cfg,
	}, nil
}

// HealthCheck verifies that Supabase services are accessible
// 檢查 Supabase 服務是否可訪問
func (c *Client) HealthCheck(ctx context.Context) error {
	// Simple health check by attempting to access the auth endpoint
	// 通過嘗試訪問認證端點進行簡單的健康檢查
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.AuthURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// AuthenticateUser authenticates a user with email and password
// 使用電子郵件和密碼驗證用戶
func (c *Client) AuthenticateUser(ctx context.Context, email, password string) (*supabase.AuthenticatedDetails, error) {
	credentials := supabase.UserCredentials{
		Email:    email,
		Password: password,
	}

	user, err := c.client.Auth.SignIn(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user account
// 創建新的用戶帳戶
func (c *Client) CreateUser(ctx context.Context, email, password string, metadata map[string]interface{}) (*supabase.AuthenticatedDetails, error) {
	credentials := supabase.UserCredentials{
		Email:    email,
		Password: password,
		Data:     metadata,
	}

	user, err := c.client.Auth.SignUp(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	return user, nil
}

// RefreshToken refreshes an authentication token
// 刷新認證令牌
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*supabase.AuthenticatedDetails, error) {
	user, err := c.client.Auth.RefreshUser(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return user, nil
}

// SignOut signs out a user
// 用戶登出
func (c *Client) SignOut(ctx context.Context, accessToken string) error {
	err := c.client.Auth.SignOut(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("sign out failed: %w", err)
	}

	return nil
}

// GetUser retrieves user information by access token
// 通過訪問令牌獲取用戶信息
func (c *Client) GetUser(ctx context.Context, accessToken string) (*supabase.User, error) {
	user, err := c.client.Auth.User(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// ResetPassword initiates a password reset for the given email
// 為給定的電子郵件啟動密碼重置
func (c *Client) ResetPassword(ctx context.Context, email string) error {
	err := c.client.Auth.ResetPasswordForEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("password reset failed: %w", err)
	}

	return nil
}

// UpdateUser updates user metadata
// 更新用戶元數據
func (c *Client) UpdateUser(ctx context.Context, accessToken string, attributes map[string]interface{}) (*supabase.User, error) {
	user, err := c.client.Auth.UpdateUser(ctx, accessToken, attributes)
	if err != nil {
		return nil, fmt.Errorf("user update failed: %w", err)
	}

	return user, nil
}

// Close closes the Supabase client connection
// 關閉 Supabase 客戶端連接
func (c *Client) Close() error {
	// Supabase Go client doesn't require explicit closing
	// Supabase Go 客戶端不需要顯式關閉
	return nil
}
