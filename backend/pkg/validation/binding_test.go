package validation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookmark-sync-service/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ValidationTestSuite struct {
	suite.Suite
	validator *RequestValidator
	router    *gin.Engine
}

func (suite *ValidationTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.validator = NewRequestValidator()
	suite.router = gin.New()
}

func (suite *ValidationTestSuite) TestUserIDFromHeader_Success() {
	suite.router.GET("/test", func(c *gin.Context) {
		userID, err := suite.validator.UserIDFromHeader(c)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), uint(123), userID)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "123")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ValidationTestSuite) TestUserIDFromHeader_Missing() {
	suite.router.GET("/test", func(c *gin.Context) {
		userID, err := suite.validator.UserIDFromHeader(c)
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), uint(0), userID)
		assert.Contains(suite.T(), err.Error(), config.ErrUserNotAuthenticated)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *ValidationTestSuite) TestUserIDFromHeader_Invalid() {
	suite.router.GET("/test", func(c *gin.Context) {
		userID, err := suite.validator.UserIDFromHeader(c)
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), uint(0), userID)
		assert.Contains(suite.T(), err.Error(), config.ErrInvalidUserID)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "invalid")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ValidationTestSuite) TestIDFromParam_Success() {
	suite.router.GET("/test/:id", func(c *gin.Context) {
		id, err := suite.validator.IDFromParam(c, "id")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), uint(456), id)
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	req := httptest.NewRequest("GET", "/test/456", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ValidationTestSuite) TestIDFromParam_Invalid() {
	suite.router.GET("/test/:id", func(c *gin.Context) {
		id, err := suite.validator.IDFromParam(c, "id")
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), uint(0), id)
		assert.Contains(suite.T(), err.Error(), "invalid id parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	req := httptest.NewRequest("GET", "/test/invalid", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ValidationTestSuite) TestBindAndValidateJSON_Success() {
	type TestStruct struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	suite.router.POST("/test", func(c *gin.Context) {
		var data TestStruct
		err := suite.validator.BindAndValidateJSON(c, &data)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "John", data.Name)
		assert.Equal(suite.T(), "john@example.com", data.Email)
		c.JSON(http.StatusOK, data)
	})

	testData := TestStruct{
		Name:  "John",
		Email: "john@example.com",
	}
	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ValidationTestSuite) TestBindAndValidateJSON_ValidationError() {
	type TestStruct struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
	}

	suite.router.POST("/test", func(c *gin.Context) {
		var data TestStruct
		err := suite.validator.BindAndValidateJSON(c, &data)
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), config.ErrInvalidData)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	testData := TestStruct{
		Name:  "", // Missing required field
		Email: "invalid-email",
	}
	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ValidationTestSuite) TestValidatePagination_Success() {
	suite.router.GET("/test", func(c *gin.Context) {
		params, err := suite.validator.ValidatePagination(c)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 2, params.Page)
		assert.Equal(suite.T(), 10, params.PageSize)
		assert.Equal(suite.T(), 10, params.Offset) // (2-1) * 10
		c.JSON(http.StatusOK, params)
	})

	req := httptest.NewRequest("GET", "/test?page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ValidationTestSuite) TestValidatePagination_Defaults() {
	suite.router.GET("/test", func(c *gin.Context) {
		params, err := suite.validator.ValidatePagination(c)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, params.Page)      // Default
		assert.Equal(suite.T(), 20, params.PageSize) // Default
		assert.Equal(suite.T(), 0, params.Offset)    // (1-1) * 20
		c.JSON(http.StatusOK, params)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ValidationTestSuite) TestValidatePagination_ValidationError() {
	suite.router.GET("/test", func(c *gin.Context) {
		params, err := suite.validator.ValidatePagination(c)
		assert.Error(suite.T(), err)
		assert.Nil(suite.T(), params)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	})

	req := httptest.NewRequest("GET", "/test?page=0&page_size=200", nil) // Invalid values
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
