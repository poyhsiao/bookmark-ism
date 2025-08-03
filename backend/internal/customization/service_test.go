package customization

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock database interface
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Create(value any) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Find(dest any, conds ...any) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Where(query any, args ...any) Database {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(Database)
}

func (m *MockDB) First(dest any, conds ...any) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Save(value any) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Delete(value any, conds ...any) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

func (m *MockDB) Order(value any) Database {
	args := m.Called(value)
	return args.Get(0).(Database)
}

func (m *MockDB) Limit(limit int) Database {
	args := m.Called(limit)
	return args.Get(0).(Database)
}

func (m *MockDB) Offset(offset int) Database {
	args := m.Called(offset)
	return args.Get(0).(Database)
}

func (m *MockDB) Preload(query string, args ...any) Database {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(Database)
}

// Mock Redis client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

// Test service creation
func TestNewService(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}

	service := NewService(mockDB, mockRedis)

	assert.NotNil(t, service)
	assert.Equal(t, mockDB, service.db)
	assert.Equal(t, mockRedis, service.redis)
}

// Test theme creation
func TestCreateTheme(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"

	config := map[string]any{
		"primaryColor":   "#007bff",
		"secondaryColor": "#6c757d",
		"darkMode":       false,
	}

	req := &CreateThemeRequest{
		Name:        "test-theme",
		DisplayName: "Test Theme",
		Description: "A test theme",
		IsPublic:    true,
		Config:      config,
		PreviewURL:  "https://example.com/preview.png",
	}

	// Mock database operations
	mockDB.On("Create", mock.AnythingOfType("*customization.Theme")).Return(&gorm.DB{Error: nil})

	theme, err := service.CreateTheme(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, theme)
	assert.Equal(t, req.Name, theme.Name)
	assert.Equal(t, req.DisplayName, theme.DisplayName)
	assert.Equal(t, req.Description, theme.Description)
	assert.Equal(t, userID, theme.CreatorID)
	assert.Equal(t, req.IsPublic, theme.IsPublic)

	mockDB.AssertExpectations(t)
}

// Test theme creation with invalid data
func TestCreateThemeInvalidData(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"

	tests := []struct {
		name    string
		request *CreateThemeRequest
		wantErr error
	}{
		{
			name: "Empty name",
			request: &CreateThemeRequest{
				Name:        "",
				DisplayName: "Test Theme",
				Config:      map[string]any{"color": "blue"},
			},
			wantErr: ErrInvalidThemeName,
		},
		{
			name: "Empty display name",
			request: &CreateThemeRequest{
				Name:        "test-theme",
				DisplayName: "",
				Config:      map[string]any{"color": "blue"},
			},
			wantErr: ErrInvalidDisplayName,
		},
		{
			name: "Name too short",
			request: &CreateThemeRequest{
				Name:        "ab",
				DisplayName: "Test Theme",
				Config:      map[string]any{"color": "blue"},
			},
			wantErr: ErrInvalidThemeName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateTheme(ctx, userID, tt.request)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// Test getting user preferences
func TestGetUserPreferences(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"

	// Mock cache miss
	mockRedis.On("Get", ctx, "user_preferences:user-123").Return("", gorm.ErrRecordNotFound)

	// Mock database query
	mockDB.On("First", mock.AnythingOfType("*customization.UserPreferences"), "user_id = ?", userID).Return(&gorm.DB{Error: nil})

	// Mock cache set
	mockRedis.On("Set", ctx, "user_preferences:user-123", mock.AnythingOfType("string"), 30*time.Minute).Return(nil)

	prefs, err := service.GetUserPreferences(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// Test updating user preferences
func TestUpdateUserPreferences(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"

	req := &UpdateUserPreferencesRequest{
		Language:  "zh-CN",
		GridSize:  "large",
		ViewMode:  "list",
		SortBy:    "title",
		SortOrder: "asc",
	}

	// Mock database operations
	mockDB.On("First", mock.AnythingOfType("*customization.UserPreferences"), "user_id = ?", userID).Return(&gorm.DB{Error: nil})
	mockDB.On("Save", mock.AnythingOfType("*customization.UserPreferences")).Return(&gorm.DB{Error: nil})

	// Mock cache operations
	mockRedis.On("Del", ctx, []string{"user_preferences:user-123"}).Return(nil)

	prefs, err := service.UpdateUserPreferences(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, req.Language, prefs.Language)
	assert.Equal(t, req.GridSize, prefs.GridSize)
	assert.Equal(t, req.ViewMode, prefs.ViewMode)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// Test setting user theme
func TestSetUserTheme(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"

	req := &SetUserThemeRequest{
		ThemeID: 1,
		Config:  map[string]any{"customColor": "#ff0000"},
	}

	// Mock theme exists check
	mockDB.On("First", mock.AnythingOfType("*customization.Theme"), uint(1)).Return(&gorm.DB{Error: nil})

	// Mock user theme operations
	mockDB.On("First", mock.AnythingOfType("*customization.UserTheme"), "user_id = ?", userID).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
	mockDB.On("Create", mock.AnythingOfType("*customization.UserTheme")).Return(&gorm.DB{Error: nil})

	// Mock cache operations
	mockRedis.On("Del", ctx, []string{"user_theme:user-123"}).Return(nil)

	userTheme, err := service.SetUserTheme(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, userTheme)
	assert.Equal(t, userID, userTheme.UserID)
	assert.Equal(t, req.ThemeID, userTheme.ThemeID)
	assert.True(t, userTheme.IsActive)

	mockDB.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// Test listing themes
func TestListThemes(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()

	req := &ThemeListRequest{
		Page:       1,
		Limit:      10,
		Search:     "test",
		SortBy:     "name",
		SortOrder:  "asc",
		PublicOnly: true,
	}

	// Mock database operations
	mockDB.On("Where", "is_public = ?", true).Return(mockDB)
	mockDB.On("Where", "name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
		"%test%", "%test%", "%test%").Return(mockDB)
	mockDB.On("Order", "name asc").Return(mockDB)
	mockDB.On("Limit", 10).Return(mockDB)
	mockDB.On("Offset", 0).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]customization.Theme")).Return(&gorm.DB{Error: nil})

	themes, total, err := service.ListThemes(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, themes)
	assert.GreaterOrEqual(t, total, int64(0))

	mockDB.AssertExpectations(t)
}

// Test rating a theme
func TestRateTheme(t *testing.T) {
	mockDB := &MockDB{}
	mockRedis := &MockRedisClient{}
	service := NewService(mockDB, mockRedis)

	ctx := context.Background()
	userID := "user-123"
	themeID := uint(1)

	req := &RateThemeRequest{
		Rating:  5,
		Comment: "Great theme!",
	}

	// Mock theme exists check
	mockDB.On("First", mock.AnythingOfType("*customization.Theme"), themeID).Return(&gorm.DB{Error: nil})

	// Mock existing rating check
	mockDB.On("First", mock.AnythingOfType("*customization.ThemeRating"),
		"user_id = ? AND theme_id = ?", userID, themeID).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	// Mock rating creation
	mockDB.On("Create", mock.AnythingOfType("*customization.ThemeRating")).Return(&gorm.DB{Error: nil})

	// Mock theme rating update
	mockDB.On("Save", mock.AnythingOfType("*customization.Theme")).Return(&gorm.DB{Error: nil})

	rating, err := service.RateTheme(ctx, userID, themeID, req)

	assert.NoError(t, err)
	assert.NotNil(t, rating)
	assert.Equal(t, userID, rating.UserID)
	assert.Equal(t, themeID, rating.ThemeID)
	assert.Equal(t, req.Rating, rating.Rating)
	assert.Equal(t, req.Comment, rating.Comment)

	mockDB.AssertExpectations(t)
}

// Test validation methods
func TestThemeValidation(t *testing.T) {
	tests := []struct {
		name    string
		theme   Theme
		wantErr error
	}{
		{
			name: "Valid theme",
			theme: Theme{
				Name:        "test-theme",
				DisplayName: "Test Theme",
				Description: "A test theme",
			},
			wantErr: nil,
		},
		{
			name: "Empty name",
			theme: Theme{
				Name:        "",
				DisplayName: "Test Theme",
			},
			wantErr: ErrInvalidThemeName,
		},
		{
			name: "Empty display name",
			theme: Theme{
				Name:        "test-theme",
				DisplayName: "",
			},
			wantErr: ErrInvalidDisplayName,
		},
		{
			name: "Name too short",
			theme: Theme{
				Name:        "ab",
				DisplayName: "Test Theme",
			},
			wantErr: ErrInvalidThemeName,
		},
		{
			name: "Name too long",
			theme: Theme{
				Name:        "this-is-a-very-long-theme-name-that-exceeds-the-maximum-allowed-length",
				DisplayName: "Test Theme",
			},
			wantErr: ErrInvalidThemeName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.theme.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test user preferences validation
func TestUserPreferencesValidation(t *testing.T) {
	tests := []struct {
		name    string
		prefs   UserPreferences
		wantErr error
	}{
		{
			name: "Valid preferences",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 300,
				SidebarWidth: 250,
			},
			wantErr: nil,
		},
		{
			name: "Empty user ID",
			prefs: UserPreferences{
				UserID: "",
			},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "Invalid language",
			prefs: UserPreferences{
				UserID:   "user-123",
				Language: "invalid",
			},
			wantErr: ErrInvalidLanguage,
		},
		{
			name: "Invalid grid size",
			prefs: UserPreferences{
				UserID:   "user-123",
				Language: "en",
				GridSize: "invalid",
			},
			wantErr: ErrInvalidGridSize,
		},
		{
			name: "Invalid sync interval",
			prefs: UserPreferences{
				UserID:       "user-123",
				Language:     "en",
				GridSize:     "medium",
				ViewMode:     "grid",
				SortBy:       "created_at",
				SortOrder:    "desc",
				SyncInterval: 30, // Too low
				SidebarWidth: 250,
			},
			wantErr: ErrInvalidSyncInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prefs.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test theme rating validation
func TestThemeRatingValidation(t *testing.T) {
	tests := []struct {
		name    string
		rating  ThemeRating
		wantErr error
	}{
		{
			name: "Valid rating",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  5,
				Comment: "Great theme!",
			},
			wantErr: nil,
		},
		{
			name: "Empty user ID",
			rating: ThemeRating{
				UserID: "",
			},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "Invalid theme ID",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 0,
			},
			wantErr: ErrInvalidThemeID,
		},
		{
			name: "Rating too low",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  0,
			},
			wantErr: ErrInvalidRating,
		},
		{
			name: "Rating too high",
			rating: ThemeRating{
				UserID:  "user-123",
				ThemeID: 1,
				Rating:  6,
			},
			wantErr: ErrInvalidRating,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rating.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
