package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/youthtrouble/symmetrical-giggle/internal/models"
)

type SQLiteRepository struct {
	db *sqlx.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS reviews (
		id TEXT PRIMARY KEY,
		app_id TEXT NOT NULL,
		author TEXT NOT NULL,
		rating INTEGER NOT NULL,
		title TEXT,
		content TEXT NOT NULL,
		submitted_date DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_reviews_app_date ON reviews(app_id, submitted_date DESC);
	CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(app_id, rating DESC);

	CREATE TABLE IF NOT EXISTS app_configs (
		app_id TEXT PRIMARY KEY,
		poll_interval INTEGER DEFAULT 300000000000, -- nanoseconds (5 minutes = 300000000000 ns)
		last_poll DATETIME,
		is_active BOOLEAN DEFAULT TRUE
	);
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return err
	}

	var count int
	err = r.db.Get(&count, "SELECT COUNT(*) FROM app_configs")
	if err != nil {
		return err
	}

	if count == 0 {
		// Inserts a default app config for the default app ID used in the frontend
		defaultAppID := "595068606"
		defaultInterval := int64(5 * time.Minute) // 5 minutes in nanoseconds

		_, err = r.db.Exec(`
			INSERT INTO app_configs (app_id, poll_interval, is_active) 
			VALUES (?, ?, TRUE)
		`, defaultAppID, defaultInterval)

		if err != nil {
			return fmt.Errorf("failed to insert default app config: %w", err)
		}
	}

	return nil
}

func (r *SQLiteRepository) CreateReview(review *models.Review) error {
	query := `
		INSERT OR IGNORE INTO reviews 
		(id, app_id, author, rating, title, content, submitted_date, created_at) 
		VALUES (:id, :app_id, :author, :rating, :title, :content, :submitted_date, :created_at)
	`
	_, err := r.db.NamedExec(query, review)
	return err
}

func (r *SQLiteRepository) GetReviews(appID string, hours int, limit int) ([]models.Review, error) {
	query := `
		SELECT * FROM reviews 
		WHERE app_id = ? AND submitted_date >= datetime('now', '-' || ? || ' hours')
		ORDER BY submitted_date DESC 
		LIMIT ?
	`

	var reviews []models.Review
	err := r.db.Select(&reviews, query, appID, hours, limit)
	return reviews, err
}

func (r *SQLiteRepository) ReviewExists(id string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM reviews WHERE id = ?", id)
	return count > 0, err
}

func (r *SQLiteRepository) GetAppConfig(appID string) (*models.AppConfig, error) {
	var config struct {
		AppID        string     `db:"app_id"`
		PollInterval int64      `db:"poll_interval"`
		LastPoll     *time.Time `db:"last_poll"`
		IsActive     bool       `db:"is_active"`
	}

	err := r.db.Get(&config, "SELECT * FROM app_configs WHERE app_id = ?", appID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &models.AppConfig{
		AppID:        config.AppID,
		PollInterval: time.Duration(config.PollInterval),
		LastPoll:     config.LastPoll,
		IsActive:     config.IsActive,
	}, nil
}

func (r *SQLiteRepository) UpsertAppConfig(config *models.AppConfig) error {

	pol1Interval := int64(config.PollInterval)

	query := `
		INSERT OR REPLACE INTO app_configs 
		(app_id, poll_interval, last_poll, is_active) 
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, config.AppID, pol1Interval, config.LastPoll, config.IsActive)
	return err
}

func (r *SQLiteRepository) GetActiveApps() ([]string, error) {
	var appIDs []string
	err := r.db.Select(&appIDs, "SELECT app_id FROM app_configs WHERE is_active = TRUE")
	return appIDs, err
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
