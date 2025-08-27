package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/youthtrouble/symmetrical-giggle/internal/models"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

type RSSService struct {
	client  *http.Client
	logger  *logger.Logger
	baseURL string
}

func NewRSSService(logger *logger.Logger) *RSSService {
	return &RSSService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		baseURL: "https://itunes.apple.com/us/rss/customerreviews/id=%s/sortBy=mostRecent/json",
	}
}

// NewRSSServiceWithURL creates a new RSS service with a custom base URL (useful for testing)
func NewRSSServiceWithURL(logger *logger.Logger, baseURL string) *RSSService {
	return &RSSService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		baseURL: baseURL,
	}
}

func (s *RSSService) FetchReviews(ctx context.Context, appID string) ([]models.Review, error) {
	url := fmt.Sprintf(s.baseURL, appID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	var rssData models.RSSFeed
	if err := json.NewDecoder(resp.Body).Decode(&rssData); err != nil {
		return nil, fmt.Errorf("failed to decode RSS feed: %w", err)
	}

	return s.parseReviews(rssData, appID)
}

func (s *RSSService) parseReviews(rssData models.RSSFeed, appID string) ([]models.Review, error) {
	var reviews []models.Review

	for _, entry := range rssData.Feed.Entry {
		// Skip the first entry as it's usually app metadata
		if len(reviews) == 0 && entry.Rating.Label == "" {
			continue
		}

		rating, err := strconv.Atoi(entry.Rating.Label)
		if err != nil {
			s.logger.Warn("Invalid rating format", "rating", entry.Rating.Label)
			continue
		}

		submittedDate, err := time.Parse(time.RFC3339, entry.Updated.Label)
		if err != nil {
			s.logger.Warn("Invalid date format", "date", entry.Updated.Label)
			continue
		}

		var title *string
		if entry.Title.Label != "" {
			title = &entry.Title.Label
		}

		review := models.Review{
			ID:            entry.ID.Label,
			AppID:         appID,
			Author:        entry.Author.Name.Label,
			Rating:        rating,
			Title:         title,
			Content:       entry.Content.Label,
			SubmittedDate: submittedDate,
			CreatedAt:     time.Now(),
		}

		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (s *RSSService) FetchWithRetry(ctx context.Context, appID string, maxRetries int) ([]models.Review, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		reviews, err := s.FetchReviews(ctx, appID)
		if err == nil {
			return reviews, nil
		}

		lastErr = err
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			s.logger.Warn("RSS fetch failed, retrying", "attempt", attempt, "backoff", backoff, "error", err)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
