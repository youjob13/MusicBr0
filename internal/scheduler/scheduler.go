package scheduler

import (
	"log"
	"music-bot/internal/database"

	"github.com/robfig/cron/v3"
)

// Scheduler handles automated playlist delivery
type Scheduler struct {
	cron *cron.Cron
	bot  PlaylistSender
	db   *database.Database
}

// PlaylistSender interface for bot that can send playlists
type PlaylistSender interface {
	SendDailyPlaylist(userTelegramID int64) error
}

// New creates a new scheduler
func New(bot PlaylistSender, db *database.Database) *Scheduler {
	c := cron.New()
	return &Scheduler{
		cron: c,
		bot:  bot,
		db:   db,
	}
}

// Start begins the scheduled tasks
func (s *Scheduler) Start() {
	// Send daily playlists at 9 AM every day
	s.cron.AddFunc("0 9 * * *", s.sendDailyPlaylists)
	
	// Optional: Send weekly discovery playlists on Sunday at 10 AM
	s.cron.AddFunc("0 10 * * 0", s.sendWeeklyDiscovery)
	
	s.cron.Start()
	log.Println("Scheduler started - daily playlists will be sent at 9 AM")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// sendDailyPlaylists sends playlists to all active users
func (s *Scheduler) sendDailyPlaylists() {
	log.Println("Starting daily playlist delivery...")
	
	users, err := s.db.GetAllActiveUsers()
	if err != nil {
		log.Printf("Failed to get active users: %v", err)
		return
	}
	
	successCount := 0
	for _, user := range users {
		err := s.bot.SendDailyPlaylist(user.TelegramID)
		if err != nil {
			log.Printf("Failed to send daily playlist to user %d: %v", user.TelegramID, err)
		} else {
			successCount++
		}
	}
	
	log.Printf("Daily playlist delivery completed. Sent to %d/%d users", successCount, len(users))
}

// sendWeeklyDiscovery sends special discovery playlists on weekends
func (s *Scheduler) sendWeeklyDiscovery() {
	log.Println("Starting weekly discovery playlist delivery...")
	
	users, err := s.db.GetAllActiveUsers()
	if err != nil {
		log.Printf("Failed to get active users: %v", err)
		return
	}
	
	successCount := 0
	for _, user := range users {
		// For now, just send regular playlists on weekends too
		// In the future, this could be enhanced with special discovery logic
		err := s.bot.SendDailyPlaylist(user.TelegramID)
		if err != nil {
			log.Printf("Failed to send weekly discovery to user %d: %v", user.TelegramID, err)
		} else {
			successCount++
		}
	}
	
	log.Printf("Weekly discovery delivery completed. Sent to %d/%d users", successCount, len(users))
} 