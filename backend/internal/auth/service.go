package auth

import (
	"context"
	"fmt"
	"time"

	"bookmark-sync-service/backend/internal/config"
	"bookmark-sync-service/backend/pkg/database"
	"bookmark-sync-service/backend/pkg/redis"
	"bookmark-sync-service/backend/pkg/supabase"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service handles authentication operations
type Service struct {
	db             *gorm.DB
	redisClient    *redis.Client
	supabaseClient *supabase.Client
	jwtConfig      *config.JWTConfig
	logger         *zap.Logger
}

// NewService creates a new authentication service
func NewService(db *gorm.DB, redisClient *redis.Client, supabaseClient *supabase.Client, jwtConfig *config.JWTConfig, logger *zap.Logger) *Service {
	return &Service{
		db:             db,
		redisClient:    redisClient,
		supabaseClient: supabaseClient,
		jwtConfig:      jwtConfig,
		logger:         logger,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         *UserInfo `json:"user"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
}

// UserInfo represents user information
type UserInfo struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar,omitempty"`
	SupabaseID  string `json:"supabase_id"`
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	s.logger.Info("User registration attempt", zap.String("email", req.Email), zap.String("username", req.Username))

	// Check if user already exists
	var existingUser database.User
	if err := s.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email or username already exists")
	}

	// For now, we'll create a simple user without Supabase integration
	// In a full implementation, you would integrate with Supabase Auth here
	supabaseID := fmt.Sprintf("user_%s_%d", req.Username, time.Now().Unix())

	// Create user in our database
	user := database.User{
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		SupabaseID:  supabaseID,
		Preferences: `{"theme": "light", "gridSize": "medium", "defaultView": "grid"}`,
	}

	if err := s.db.Create(&user).Error; err != nil {
		s.logger.Error("Failed to create user in database", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateTokens(&user)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in Redis
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	s.logger.Info("User registered successfully", zap.Uint("user_id", user.ID), zap.String("email", req.Email))

	return &AuthResponse{
		User: &UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			SupabaseID:  user.SupabaseID,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtConfig.ExpiryHour * 3600,
	}, nil
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	s.logger.Info("User login attempt", zap.String("email", req.Email))

	// For now, we'll do basic email/password validation
	// In a full implementation, you would integrate with Supabase Auth here
	var user database.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		s.logger.Error("User not found", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("invalid credentials")
	}

	// For demo purposes, we'll accept any password for existing users
	// In production, you would validate the password hash
	if req.Password == "" {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last active timestamp
	now := time.Now()
	user.LastActiveAt = &now
	if err := s.db.Save(&user).Error; err != nil {
		s.logger.Warn("Failed to update last active timestamp", zap.Error(err), zap.Uint("user_id", user.ID))
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.generateTokens(&user)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in Redis
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	s.logger.Info("User logged in successfully", zap.Uint("user_id", user.ID), zap.String("email", req.Email))

	return &AuthResponse{
		User: &UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			SupabaseID:  user.SupabaseID,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtConfig.ExpiryHour * 3600,
	}, nil
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(ctx context.Context, req *RefreshRequest) (*AuthResponse, error) {
	// Parse refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}
	userID := uint(userIDFloat)

	// Check if refresh token exists in Redis
	storedToken, err := s.redisClient.Get(ctx, fmt.Sprintf("refresh_token:%d", userID)).Result()
	if err != nil || storedToken != req.RefreshToken {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// Get user from database
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	accessToken, refreshToken, err := s.generateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store new refresh token in Redis
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	s.logger.Info("Token refreshed successfully", zap.Uint("user_id", user.ID))

	return &AuthResponse{
		User: &UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			SupabaseID:  user.SupabaseID,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtConfig.ExpiryHour * 3600,
	}, nil
}

// Logout logs out a user
func (s *Service) Logout(ctx context.Context, userID uint) error {
	// Remove refresh token from Redis
	if err := s.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%d", userID)).Err(); err != nil {
		s.logger.Warn("Failed to remove refresh token from Redis", zap.Error(err), zap.Uint("user_id", userID))
	}

	s.logger.Info("User logged out successfully", zap.Uint("user_id", userID))
	return nil
}

// ResetPassword initiates password reset
func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	s.logger.Info("Password reset requested", zap.String("email", req.Email))

	// Check if user exists
	var user database.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		s.logger.Error("User not found for password reset", zap.Error(err), zap.String("email", req.Email))
		return fmt.Errorf("user not found")
	}

	// For now, we'll just log the password reset request
	// In a full implementation, you would integrate with Supabase Auth or send an email
	s.logger.Info("Password reset email would be sent", zap.String("email", req.Email))
	return nil
}

// ValidateToken validates a JWT token and returns user information
func (s *Service) ValidateToken(tokenString string) (*UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}
	userID := uint(userIDFloat)

	// Get user from database
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Avatar:      user.Avatar,
		SupabaseID:  user.SupabaseID,
	}, nil
}

// generateTokens generates access and refresh tokens
func (s *Service) generateTokens(user *database.User) (string, string, error) {
	now := time.Now()

	// Access token claims
	accessClaims := jwt.MapClaims{
		"user_id":     user.ID,
		"email":       user.Email,
		"username":    user.Username,
		"supabase_id": user.SupabaseID,
		"sub":         user.SupabaseID,
		"iat":         now.Unix(),
		"exp":         now.Add(time.Duration(s.jwtConfig.ExpiryHour) * time.Hour).Unix(),
		"type":        "access",
	}

	// Refresh token claims (longer expiry)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"sub":     user.SupabaseID,
		"iat":     now.Unix(),
		"exp":     now.Add(7 * 24 * time.Hour).Unix(), // 7 days
		"type":    "refresh",
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

// storeRefreshToken stores refresh token in Redis
func (s *Service) storeRefreshToken(ctx context.Context, userID uint, token string) error {
	key := fmt.Sprintf("refresh_token:%d", userID)
	return s.redisClient.Set(ctx, key, token, 7*24*time.Hour).Err()
}
