package community

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

// SocialMetricsServiceTestSuite tests the SocialMetricsService
type SocialMetricsServiceTestSuite struct {
	suite.Suite
	service   *SocialMetricsService
	mockDB    *MockDB
	mockRedis *MockRedisClient
	ctx       context.Context
}

func (suite *SocialMetricsServiceTestSuite) SetupTest() {
	suite.mockDB = new(MockDB)
	suite.mockRedis = new(MockRedisClient)
	suite.ctx = context.Background()
	logger := zaptest.NewLogger(suite.T())
	jsonHelper := NewJSONHelper()

	suite.service = NewSocialMetricsService(suite.mockDB, suite.mockRedis, jsonHelper, logger)
}

func (suite *SocialMetricsServiceTestSuite) TearDownTest() {
	suite.mockDB.AssertExpectations(suite.T())
	suite.mockRedis.AssertExpectations(suite.T())
}

func (suite *SocialMetricsServiceTestSuite) TestGetSocialMetrics_Success() {
	bookmarkID := uint(1)

	// Mock finding social metrics - populate the struct with expected data
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Run(func(args mock.Arguments) {
		metrics := args.Get(0).(*SocialMetrics)
		metrics.ID = 1
		metrics.BookmarkID = bookmarkID
		metrics.TotalViews = 10
		metrics.TotalClicks = 5
	}).Return(&gorm.DB{Error: nil})

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), metrics)
	assert.Equal(suite.T(), bookmarkID, metrics.BookmarkID)
}

func (suite *SocialMetricsServiceTestSuite) TestGetSocialMetrics_InvalidBookmarkID() {
	bookmarkID := uint(0)

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), metrics)
	assert.Equal(suite.T(), ErrInvalidBookmarkID, err)
}

func (suite *SocialMetricsServiceTestSuite) TestGetSocialMetrics_NotFound() {
	bookmarkID := uint(999)

	// Mock not finding social metrics
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})

	metrics, err := suite.service.GetSocialMetrics(suite.ctx, bookmarkID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), metrics)
	assert.Equal(suite.T(), ErrBookmarkNotFound, err)
}

func (suite *SocialMetricsServiceTestSuite) TestUpdateSocialMetrics_CreateNew() {
	bookmarkID := uint(1)
	actionType := "view"

	// Mock not finding existing metrics (will create new)
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
	suite.mockDB.On("Create", mock.AnythingOfType("*community.SocialMetrics")).Return(&gorm.DB{Error: nil})

	err := suite.service.UpdateSocialMetrics(suite.ctx, bookmarkID, actionType)

	assert.NoError(suite.T(), err)
}

func (suite *SocialMetricsServiceTestSuite) TestUpdateSocialMetrics_UpdateExisting() {
	bookmarkID := uint(1)
	actionType := "click"

	// Mock finding existing metrics (will update) - need to set ID to simulate existing record
	suite.mockDB.On("First", mock.AnythingOfType("*community.SocialMetrics"), mock.Anything).Run(func(args mock.Arguments) {
		metrics := args.Get(0).(*SocialMetrics)
		metrics.ID = 1 // Set ID to simulate existing record
		metrics.BookmarkID = bookmarkID
	}).Return(&gorm.DB{Error: nil})
	suite.mockDB.On("Save", mock.AnythingOfType("*community.SocialMetrics")).Return(&gorm.DB{Error: nil})

	err := suite.service.UpdateSocialMetrics(suite.ctx, bookmarkID, actionType)

	assert.NoError(suite.T(), err)
}

func (suite *SocialMetricsServiceTestSuite) TestUpdateSocialMetrics_InvalidBookmarkID() {
	bookmarkID := uint(0)
	actionType := "view"

	err := suite.service.UpdateSocialMetrics(suite.ctx, bookmarkID, actionType)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidBookmarkID, err)
}

func TestSocialMetricsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SocialMetricsServiceTestSuite))
}

// Unit tests for metric calculation logic
func TestSocialMetricsService_MetricCalculations(t *testing.T) {
	service := &SocialMetricsService{}

	tests := []struct {
		name           string
		metrics        SocialMetrics
		actionType     string
		expectedViews  int
		expectedClicks int
	}{
		{
			name: "View action increments views",
			metrics: SocialMetrics{
				TotalViews:  5,
				TotalClicks: 2,
			},
			actionType:     "view",
			expectedViews:  6,
			expectedClicks: 2,
		},
		{
			name: "Click action increments clicks",
			metrics: SocialMetrics{
				TotalViews:  5,
				TotalClicks: 2,
			},
			actionType:     "click",
			expectedViews:  5,
			expectedClicks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service.updateMetricsByAction(&tt.metrics, tt.actionType)
			assert.Equal(t, tt.expectedViews, tt.metrics.TotalViews)
			assert.Equal(t, tt.expectedClicks, tt.metrics.TotalClicks)
		})
	}
}

func TestSocialMetricsService_DerivedMetricsCalculation(t *testing.T) {
	service := &SocialMetricsService{}

	metrics := SocialMetrics{
		TotalViews:  100,
		TotalClicks: 20,
		TotalSaves:  10,
		TotalShares: 5,
		TotalLikes:  15,
	}

	service.calculateDerivedMetrics(&metrics)

	// Engagement rate should be (20+10+5+15)/100 = 0.5
	assert.Equal(t, 0.5, metrics.EngagementRate)

	// Virality score should be 5*2.0 + 10*1.5 = 25.0
	assert.Equal(t, 25.0, metrics.ViralityScore)

	// Quality score should be (0.5 * 0.6) + (15/100 * 0.4) = 0.36
	assert.Equal(t, 0.36, metrics.QualityScore)
}
