package repository

import (
	"testing"
	"time"

	"github.com/youthtrouble/symmetrical-giggle/internal/models"
)

func TestSQLiteRepository_CreateAndGetReviews(t *testing.T) {

	repo, err := NewSQLiteRepository(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	defer repo.Close()

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

	err = repo.CreateReview(review)
	if err != nil {
		t.Fatalf("Failed to create review: %v", err)
	}

	reviews, err := repo.GetReviews("123456", 24, 10)
	if err != nil {
		t.Fatalf("Failed to get reviews: %v", err)
	}

	if len(reviews) != 1 {
		t.Fatalf("Expected 1 review, got %d", len(reviews))
	}

	if reviews[0].Author != "Test User" {
		t.Errorf("Expected author 'Test User', got '%s'", reviews[0].Author)
	}

	exists, err := repo.ReviewExists("test-review-1")
	if err != nil {
		t.Fatalf("Failed to check if review exists: %v", err)
	}

	if !exists {
		t.Error("Review should exist")
	}
}

func stringPtr(s string) *string {
	return &s
}
