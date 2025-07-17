package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Database represents the database connection and operations
type Database struct {
	db *sql.DB
}

// User represents a Telegram user
type User struct {
	ID          int       `json:"id"`
	TelegramID  int64     `json:"telegram_id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	CreatedAt   time.Time `json:"created_at"`
	LastActive  time.Time `json:"last_active"`
}

// Artist represents a user's favorite artist
type Artist struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ArtistName string    `json:"artist_name"`
	SpotifyID  string    `json:"spotify_id"`
	AddedAt    time.Time `json:"added_at"`
}

// Playlist represents a generated playlist
type Playlist struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	PlaylistData string    `json:"playlist_data"` // JSON string of tracks
	CreatedAt    time.Time `json:"created_at"`
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{db: db}

	// Initialize tables
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables creates the necessary database tables
func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER UNIQUE NOT NULL,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_active DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_artists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			artist_name TEXT NOT NULL,
			spotify_id TEXT,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id),
			UNIQUE(user_id, artist_name)
		)`,
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			playlist_data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// CreateUser creates or updates a user in the database
func (d *Database) CreateUser(telegramID int64, username, firstName, lastName string) (*User, error) {
	query := `INSERT OR REPLACE INTO users (telegram_id, username, first_name, last_name, last_active) 
			  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`
	
	_, err := d.db.Exec(query, telegramID, username, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return d.GetUserByTelegramID(telegramID)
}

// GetUserByTelegramID retrieves a user by their Telegram ID
func (d *Database) GetUserByTelegramID(telegramID int64) (*User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, created_at, last_active 
			  FROM users WHERE telegram_id = ?`
	
	row := d.db.QueryRow(query, telegramID)
	
	user := &User{}
	err := row.Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, 
					&user.LastName, &user.CreatedAt, &user.LastActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// AddUserArtist adds an artist to user's favorites
func (d *Database) AddUserArtist(userID int, artistName, spotifyID string) error {
	query := `INSERT OR IGNORE INTO user_artists (user_id, artist_name, spotify_id) VALUES (?, ?, ?)`
	
	_, err := d.db.Exec(query, userID, artistName, spotifyID)
	if err != nil {
		return fmt.Errorf("failed to add artist: %w", err)
	}

	return nil
}

// RemoveUserArtist removes an artist from user's favorites
func (d *Database) RemoveUserArtist(userID int, artistName string) error {
	query := `DELETE FROM user_artists WHERE user_id = ? AND artist_name = ?`
	
	result, err := d.db.Exec(query, userID, artistName)
	if err != nil {
		return fmt.Errorf("failed to remove artist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artist not found in favorites")
	}

	return nil
}

// GetUserArtists retrieves all favorite artists for a user
func (d *Database) GetUserArtists(userID int) ([]Artist, error) {
	query := `SELECT id, user_id, artist_name, spotify_id, added_at 
			  FROM user_artists WHERE user_id = ? ORDER BY added_at DESC`
	
	rows, err := d.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user artists: %w", err)
	}
	defer rows.Close()

	var artists []Artist
	for rows.Next() {
		var artist Artist
		err := rows.Scan(&artist.ID, &artist.UserID, &artist.ArtistName, 
						&artist.SpotifyID, &artist.AddedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan artist: %w", err)
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

// SavePlaylist saves a generated playlist
func (d *Database) SavePlaylist(userID int, playlistData string) error {
	query := `INSERT INTO playlists (user_id, playlist_data) VALUES (?, ?)`
	
	_, err := d.db.Exec(query, userID, playlistData)
	if err != nil {
		return fmt.Errorf("failed to save playlist: %w", err)
	}

	return nil
}

// GetAllActiveUsers retrieves all users for daily playlist generation
func (d *Database) GetAllActiveUsers() ([]User, error) {
	query := `SELECT id, telegram_id, username, first_name, last_name, created_at, last_active 
			  FROM users WHERE last_active > datetime('now', '-7 days')`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, 
						&user.LastName, &user.CreatedAt, &user.LastActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUserActivity updates the user's last active timestamp
func (d *Database) UpdateUserActivity(telegramID int64) error {
	query := `UPDATE users SET last_active = CURRENT_TIMESTAMP WHERE telegram_id = ?`
	
	_, err := d.db.Exec(query, telegramID)
	if err != nil {
		return fmt.Errorf("failed to update user activity: %w", err)
	}

	return nil
} 