// internal/services/rss_test.go
package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/youthtrouble/symmetrical-giggle/internal/models"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

func TestRSSService_FetchReviews(t *testing.T) {
	// Mock RSS response
	mockRSSData := models.RSSFeed{
		Feed: struct {
			Entry []models.RSSEntry `json:"entry"`
		}{
			Entry: []models.RSSEntry{
				{
					ID: struct {
						Label string `json:"label"`
					}{Label: "review-1"},
					Author: struct {
						Name struct {
							Label string `json:"label"`
						} `json:"name"`
					}{Name: struct {
						Label string `json:"label"`
					}{Label: "Test User"}},
					Rating: struct {
						Label string `json:"label"`
					}{Label: "5"},
					Title: struct {
						Label string `json:"label"`
					}{Label: "Great App!"},
					Content: struct {
						Label string `json:"label"`
					}{Label: "This app is amazing!"},
					Updated: struct {
						Label string `json:"label"`
					}{Label: time.Now().Format(time.RFC3339)},
				},
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockRSSData)
	}))
	defer server.Close()

	// Override URL for testing
	logger := logger.New("info")
	service := NewRSSServiceWithURL(logger, server.URL+"/%s")

	// Test with mock server
	ctx := context.Background()
	reviews, err := service.FetchReviews(ctx, "123456")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf("Expected 1 review, got %d", len(reviews))
	}

	review := reviews[0]
	if review.Author != "Test User" {
		t.Errorf("Expected author 'Test User', got '%s'", review.Author)
	}

	if review.Rating != 5 {
		t.Errorf("Expected rating 5, got %d", review.Rating)
	}
}
