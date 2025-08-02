package sharing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bookmark-sync-service/backend/pkg/database"
)

// SharingServiceTestSuite defines the test suite for sharing service
type SharingServiceTestSuite struct {
	suite.Suite
	service *Service
	db      *gorm.DB
}

func (suite *SharingServiceTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Run migrations
	err = db.AutoMigrate(
		&database.User{},
		&database.Collection{},
		&database.Bookmark{},
		&CollectionShare{},
		&CollectionCollaborator{},
		&CollectionFork{},
		&ShareActivity{},
	)
	suite.Require().NoError(err)

	suite.db = db
	suite.service = NewService(db, "http://localhost:3000")
}

func (suite *SharingServiceTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.Exec("DELETE FROM collection_shares")
	suite.db.Exec("DELETE FROM collection_collaborators")
	suite.db.Exec("DELETE FROM collection_forks")
	suite.db.Exec("DELETE FROM share_activities")
	suite.db.Exec("DELETE FROM collections")
	suite.db.Exec("DELETE FROM bookmarks")
	suite.db.Exec("DELETE FROM users")
}

func (suite *SharingServiceTestSuite) TestCreateShare() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:      user.ID,
		Name:        "Test Collection",
		Description: "Test Description",
		Visibility:  "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	request := &CreateShareRequest{
		CollectionID: collection.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		Title:        "Test Share",
		Description:  "Test Description",
	}

	share, err := suite.service.CreateShare(context.Background(), user.ID, request)

	suite.NoError(err)
	suite.NotNil(share)
	suite.Equal(collection.ID, share.CollectionID)
	suite.Equal(ShareTypePublic, share.ShareType)
	suite.Equal(PermissionView, share.Permission)
	suite.Equal("Test Share", share.Title)
	suite.NotEmpty(share.ShareToken)
	suite.True(share.IsActive)
}

func (suite *SharingServiceTestSuite) TestCreateShareCollectionNotFound() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	request := &CreateShareRequest{
		CollectionID: 999, // Non-existent collection
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
	}

	share, err := suite.service.CreateShare(context.Background(), user.ID, request)

	suite.Error(err)
	suite.Equal(ErrCollectionNotFound, err)
	suite.Nil(share)
}

func (suite *SharingServiceTestSuite) TestGetShareByToken() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:      user.ID,
		Name:        "Test Collection",
		Description: "Test Description",
		Visibility:  "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	// Create test share
	shareToken := "test-token-123"
	testShare := &CollectionShare{
		CollectionID: collection.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		ShareToken:   shareToken,
		Title:        "Test Share",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(testShare).Error)

	share, err := suite.service.GetShareByToken(context.Background(), shareToken)

	suite.NoError(err)
	suite.NotNil(share)
	suite.Equal(shareToken, share.ShareToken)
	suite.Equal("Test Share", share.Title)
	suite.True(share.IsActive)
}

func (suite *SharingServiceTestSuite) TestGetShareByTokenNotFound() {
	share, err := suite.service.GetShareByToken(context.Background(), "invalid-token")

	suite.Error(err)
	suite.Equal(ErrShareNotFound, err)
	suite.Nil(share)
}

func (suite *SharingServiceTestSuite) TestGetShareByTokenExpired() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:      user.ID,
		Name:        "Test Collection",
		Description: "Test Description",
		Visibility:  "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	// Create expired share
	shareToken := "expired-token"
	expiredTime := time.Now().Add(-1 * time.Hour)
	testShare := &CollectionShare{
		CollectionID: collection.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		ShareToken:   shareToken,
		Title:        "Expired Share",
		IsActive:     true,
		ExpiresAt:    &expiredTime,
	}
	suite.Require().NoError(suite.db.Create(testShare).Error)

	share, err := suite.service.GetShareByToken(context.Background(), shareToken)

	suite.Error(err)
	suite.Equal(ErrShareExpired, err)
	suite.Nil(share)
}

func (suite *SharingServiceTestSuite) TestUpdateShare() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:      user.ID,
		Name:        "Test Collection",
		Description: "Test Description",
		Visibility:  "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	// Create test share
	testShare := &CollectionShare{
		CollectionID: collection.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePrivate,
		Permission:   PermissionView,
		ShareToken:   "test-token",
		Title:        "Old Title",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(testShare).Error)

	newTitle := "New Title"
	newShareType := ShareTypePublic
	request := &UpdateShareRequest{
		Title:     &newTitle,
		ShareType: &newShareType,
	}

	share, err := suite.service.UpdateShare(context.Background(), user.ID, testShare.ID, request)

	suite.NoError(err)
	suite.NotNil(share)
	suite.Equal("New Title", share.Title)
	suite.Equal(ShareTypePublic, share.ShareType)
}

func (suite *SharingServiceTestSuite) TestDeleteShare() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:      user.ID,
		Name:        "Test Collection",
		Description: "Test Description",
		Visibility:  "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	// Create test share
	testShare := &CollectionShare{
		CollectionID: collection.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		ShareToken:   "test-token",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(testShare).Error)

	err := suite.service.DeleteShare(context.Background(), user.ID, testShare.ID)

	suite.NoError(err)

	// Verify share is deleted
	var deletedShare CollectionShare
	err = suite.db.First(&deletedShare, testShare.ID).Error
	suite.Error(err) // Should not find the share
}

func (suite *SharingServiceTestSuite) TestGetUserShares() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collections
	collection1 := &database.Collection{
		UserID:     user.ID,
		Name:       "Collection 1",
		Visibility: "private",
		ShareLink:  "share-link-1",
	}
	suite.Require().NoError(suite.db.Create(collection1).Error)

	collection2 := &database.Collection{
		UserID:     user.ID,
		Name:       "Collection 2",
		Visibility: "private",
		ShareLink:  "share-link-2",
	}
	suite.Require().NoError(suite.db.Create(collection2).Error)

	// Create test shares
	share1 := &CollectionShare{
		CollectionID: collection1.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		ShareToken:   "token1",
		Title:        "Share 1",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(share1).Error)

	share2 := &CollectionShare{
		CollectionID: collection2.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePrivate,
		Permission:   PermissionEdit,
		ShareToken:   "token2",
		Title:        "Share 2",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(share2).Error)

	shares, err := suite.service.GetUserShares(context.Background(), user.ID)

	suite.NoError(err)
	suite.Len(shares, 2)
}

func (suite *SharingServiceTestSuite) TestRecordActivity() {
	// Create test user
	user := &database.User{
		Email:      "test@example.com",
		Username:   "testuser",
		SupabaseID: "test-supabase-id",
	}
	suite.Require().NoError(suite.db.Create(user).Error)

	// Create test collection
	collection := &database.Collection{
		UserID:     user.ID,
		Name:       "Test Collection",
		Visibility: "private",
	}
	suite.Require().NoError(suite.db.Create(collection).Error)

	// Create test share
	testShare := &CollectionShare{
		CollectionID: collection.ID,
		UserID:       user.ID,
		ShareType:    ShareTypePublic,
		Permission:   PermissionView,
		ShareToken:   "test-token",
		IsActive:     true,
	}
	suite.Require().NoError(suite.db.Create(testShare).Error)

	err := suite.service.RecordActivity(context.Background(), testShare.ID, &user.ID, "view", "192.168.1.1", "Mozilla/5.0", nil)

	suite.NoError(err)

	// Verify activity was recorded
	var activity ShareActivity
	err = suite.db.First(&activity, "share_id = ?", testShare.ID).Error
	suite.NoError(err)
	suite.Equal(testShare.ID, activity.ShareID)
	suite.Equal(&user.ID, activity.UserID)
	suite.Equal("view", activity.ActivityType)
}

// Run the test suite
func TestSharingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SharingServiceTestSuite))
}

// Test validation functions
func TestCreateShareRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request *CreateShareRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &CreateShareRequest{
				CollectionID: 1,
				ShareType:    ShareTypePublic,
				Permission:   PermissionView,
			},
			wantErr: nil,
		},
		{
			name: "missing collection ID",
			request: &CreateShareRequest{
				ShareType:  ShareTypePublic,
				Permission: PermissionView,
			},
			wantErr: ErrInvalidCollectionID,
		},
		{
			name: "missing share type",
			request: &CreateShareRequest{
				CollectionID: 1,
				Permission:   PermissionView,
			},
			wantErr: ErrInvalidShareType,
		},
		{
			name: "missing permission",
			request: &CreateShareRequest{
				CollectionID: 1,
				ShareType:    ShareTypePublic,
			},
			wantErr: ErrInvalidPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCollaboratorRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request *CollaboratorRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &CollaboratorRequest{
				Email:      "test@example.com",
				Permission: PermissionEdit,
			},
			wantErr: nil,
		},
		{
			name: "missing email",
			request: &CollaboratorRequest{
				Permission: PermissionEdit,
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "missing permission",
			request: &CollaboratorRequest{
				Email: "test@example.com",
			},
			wantErr: ErrInvalidPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestForkRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request *ForkRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &ForkRequest{
				Name: "Forked Collection",
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			request: &ForkRequest{
				Description: "Test description",
			},
			wantErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
