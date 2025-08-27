package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Polling  PollingConfig
	LogLevel string
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Path string
}

type PollingConfig struct {
	DefaultInterval time.Duration
	MaxConcurrent   int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "4000"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./reviews.db"),
		},
		Polling: PollingConfig{
			DefaultInterval: parseDuration(getEnv("POLL_INTERVAL", "5m")),
			MaxConcurrent:   parseInt(getEnv("MAX_CONCURRENT_POLLS", "10")),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 10
	}
	return i
}
