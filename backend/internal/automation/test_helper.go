package automation

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AutomationTestBase provides common test setup and teardown functionality
type AutomationTestBase struct {
	suite.Suite
	db      *gorm.DB
	service *Service
	userID  string
}

// SetupAutomationTest creates a fresh database and service for testing
func (base *AutomationTestBase) SetupAutomationTest() {
	base.userID = "test-user-123"

	// Create a new in-memory SQLite database for each test
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	base.Require().NoError(err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&WebhookEndpoint{},
		&WebhookDelivery{},
		&RSSFeed{},
		&BulkOperation{},
		&BackupJob{},
		&APIIntegration{},
		&AutomationRule{},
	)
	base.Require().NoError(err)

	base.db = db
	base.service = NewServiceForTesting(db)
}

// TearDownAutomationTest cleans up the database connection
func (base *AutomationTestBase) TearDownAutomationTest() {
	if base.db != nil {
		sqlDB, _ := base.db.DB()
		sqlDB.Close()
	}
}

// SetupGinRouter creates a configured Gin router for handler testing
func (base *AutomationTestBase) SetupGinRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", base.userID)
		c.Next()
	})

	return router
}

// GetTestUserID returns the test user ID
func (base *AutomationTestBase) GetTestUserID() string {
	return base.userID
}

// GetTestDB returns the test database
func (base *AutomationTestBase) GetTestDB() *gorm.DB {
	return base.db
}

// GetTestService returns the test service
func (base *AutomationTestBase) GetTestService() *Service {
	return base.service
}
