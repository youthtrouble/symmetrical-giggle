package repository

import (
	"github.com/youthtrouble/symmetrical-giggle/internal/models"
)

type Repository interface {
	CreateReview(review *models.Review) error
	GetReviews(appID string, hours int, limit int) ([]models.Review, error)
	ReviewExists(id string) (bool, error)

	GetAppConfig(appID string) (*models.AppConfig, error)
	UpsertAppConfig(config *models.AppConfig) error
	GetActiveApps() ([]string, error)

	Close() error
}
