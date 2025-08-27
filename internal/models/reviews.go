package models

import (
	"time"
)

type Review struct {
	ID            string    `json:"id" db:"id"`
	AppID         string    `json:"app_id" db:"app_id"`
	Author        string    `json:"author" db:"author"`
	Rating        int       `json:"rating" db:"rating"`
	Title         *string   `json:"title" db:"title"`
	Content       string    `json:"content" db:"content"`
	SubmittedDate time.Time `json:"submitted_date" db:"submitted_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type AppConfig struct {
	AppID        string        `json:"app_id" db:"app_id"`
	PollInterval time.Duration `json:"poll_interval" db:"poll_interval"`
	LastPoll     *time.Time    `json:"last_poll" db:"last_poll"`
	IsActive     bool          `json:"is_active" db:"is_active"`
}

type RSSFeed struct {
	Feed struct {
		Entry []RSSEntry `json:"entry"`
	} `json:"feed"`
}

type RSSEntry struct {
	ID struct {
		Label string `json:"label"`
	} `json:"id"`
	Author struct {
		Name struct {
			Label string `json:"label"`
		} `json:"name"`
	} `json:"author"`
	Rating struct {
		Label string `json:"label"`
	} `json:"im:rating"`
	Title struct {
		Label string `json:"label"`
	} `json:"title"`
	Content struct {
		Label string `json:"label"`
	} `json:"content"`
	Updated struct {
		Label string `json:"label"`
	} `json:"updated"`
}
