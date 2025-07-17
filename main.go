package main

import (
	"log"
	"music-bot/internal/bot"
	"music-bot/internal/config"
	"music-bot/internal/database"
	"music-bot/internal/scheduler"
	"music-bot/internal/spotify"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize Spotify client
	spotifyClient, err := spotify.New(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	if err != nil {
		log.Fatal("Failed to initialize Spotify client:", err)
	}

	// Initialize bot
	musicBot, err := bot.New(cfg.TelegramToken, db, spotifyClient)
	if err != nil {
		log.Fatal("Failed to initialize bot:", err)
	}

	// Start scheduler for daily playlists
	scheduler := scheduler.New(musicBot, db)
	scheduler.Start()

	log.Println("Music bot started successfully!")

	// Start bot
	musicBot.Start()
} 