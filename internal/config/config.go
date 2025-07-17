package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	TelegramToken       string
	SpotifyClientID     string
	SpotifyClientSecret string
	DatabasePath        string
	LogLevel            string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		TelegramToken:       os.Getenv("TELEGRAM_TOKEN"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		DatabasePath:        getEnvWithDefault("DATABASE_PATH", "music_bot.db"),
		LogLevel:            getEnvWithDefault("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN environment variable is required")
	}

	if cfg.SpotifyClientID == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID environment variable is required")
	}

	if cfg.SpotifyClientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_SECRET environment variable is required")
	}

	return cfg, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 