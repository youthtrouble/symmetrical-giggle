// internal/integration/integration_test.go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"github.com/youthtrouble/symmetrical-giggle/internal/api"
	"github.com/youthtrouble/symmetrical-giggle/internal/models"
	"github.com/youthtrouble/symmetrical-giggle/internal/repository"
	"github.com/youthtrouble/symmetrical-giggle/internal/services"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

type IntegrationTestSuite struct {
	suite.Suite
	router   *gin.Engine
	repo     repository.Repository
	handlers *api.Handlers
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	repo, err := repository.NewSQLiteRepository(":memory:")
	s.Require().NoError(err)
	s.repo = repo

	logger := logger.New("error")
	rssService := services.NewRSSService(logger)
	pollingManager := services.NewPollingManager(repo, rssService, logger)

	s.handlers = api.NewHandlers(repo, pollingManager, logger)

	s.router = gin.New()
	api.SetupRoutes(s.router, s.handlers)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.repo.Close()
}

func (s *IntegrationTestSuite) TestGetReviewsEndpoint() {
	review := &models.Review{
		ID:            "test-review-1",
		AppID:         "123456",
		Author:        "Test User",
		Rating:        5,
		Title:         stringPtr("Great App"),
		Content:       "This is a test review",
		SubmittedDate: time.Now(),
		CreatedAt:     time.Now(),
	}

	err := s.repo.CreateReview(review)
	s.Require().NoError(err)

	req, _ := http.NewRequest("GET", "/api/reviews/123456", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Assert().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)

	reviews, ok := response["reviews"].([]interface{})
	s.Assert().True(ok)
	s.Assert().Len(reviews, 1)
}

func (s *IntegrationTestSuite) TestConfigureAppEndpoint() {
	configData := map[string]interface{}{
		"poll_interval": "10m",
		"is_active":     true,
	}

	jsonData, _ := json.Marshal(configData)
	req, _ := http.NewRequest("POST", "/api/apps/123456/configure", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Assert().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)

	s.Assert().Equal("Configuration updated successfully", response["message"])
}

func (s *IntegrationTestSuite) TestHealthCheckEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Assert().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)

	s.Assert().Equal("healthy", response["status"])
}

func stringPtr(s string) *string {
	return &s
}
