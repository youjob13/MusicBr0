package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"music-bot/internal/database"
	"music-bot/internal/spotify"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot represents the Telegram music bot
type Bot struct {
	api           *tgbotapi.BotAPI
	db            *database.Database
	spotifyClient *spotify.Client
	userStates    map[int64]string // Track user interaction states
}

// New creates a new music bot instance
func New(token string, db *database.Database, spotifyClient *spotify.Client) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	api.Debug = true
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:           api,
		db:            db,
		spotifyClient: spotifyClient,
		userStates:    make(map[int64]string),
	}, nil
}

// Start begins the bot's main message handling loop
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

// handleMessage processes text messages from users
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Update user activity
	b.db.UpdateUserActivity(message.From.ID)

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	// Handle text based on user state
	userState := b.userStates[message.From.ID]
	switch userState {
	case "waiting_artist_name":
		b.handleArtistSearch(message)
	default:
		b.sendHelp(message.Chat.ID)
	}
}

// handleCommand processes bot commands
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.sendHelp(message.Chat.ID)
	case "add":
		b.handleAddArtist(message)
	case "list":
		b.handleListArtists(message)
	case "remove":
		b.handleRemoveArtist(message)
	case "playlist":
		b.handleGeneratePlaylist(message)
	default:
		b.sendMessage(message.Chat.ID, "Unknown command. Use /help to see available commands.")
	}
}

// handleStart handles the /start command and user registration
func (b *Bot) handleStart(message *tgbotapi.Message) {
	// Register or update user
	_, err := b.db.CreateUser(
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LastName,
	)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, there was an error setting up your account.")
		return
	}

	welcomeText := fmt.Sprintf(`🎵 Welcome to Music Bot, %s!

I can help you discover new music based on your favorite artists.

Here's what I can do:
🎤 Add your favorite artists
📋 Generate personalized playlists
🔍 Discover similar artists
📅 Get daily music recommendations

Use /help to see all available commands.

Let's start by adding some of your favorite artists! Use /add to begin.`, message.From.FirstName)

	b.sendMessage(message.Chat.ID, welcomeText)
}

// sendHelp sends the help message
func (b *Bot) sendHelp(chatID int64) {
	helpText := `🎵 Music Bot Commands:

/start - Start using the bot
/help - Show this help message
/add - Add a favorite artist
/list - Show your favorite artists
/remove - Remove an artist from favorites
/playlist - Generate a new playlist

🎯 How it works:
1. Add your favorite artists with /add
2. Generate playlists with /playlist
3. Discover new music based on your taste!

Daily playlists are automatically sent every morning at 9 AM.`

	b.sendMessage(chatID, helpText)
}

// handleAddArtist handles adding a favorite artist
func (b *Bot) handleAddArtist(message *tgbotapi.Message) {
	b.userStates[message.From.ID] = "waiting_artist_name"
	b.sendMessage(message.Chat.ID, "🎤 Please send me the name of the artist you'd like to add:")
}

// handleArtistSearch processes artist search and selection
func (b *Bot) handleArtistSearch(message *tgbotapi.Message) {
	artistName := strings.TrimSpace(message.Text)
	if artistName == "" {
		b.sendMessage(message.Chat.ID, "Please enter a valid artist name.")
		return
	}

	// Search for artists on Spotify
	artists, err := b.spotifyClient.SearchArtists(artistName, 5)
	if err != nil {
		log.Printf("Failed to search artists: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, I couldn't search for artists right now. Please try again later.")
		return
	}

	if len(artists) == 0 {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("No artists found for '%s'. Please try a different name.", artistName))
		return
	}

	// Create inline keyboard with artist options
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, artist := range artists {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", artist.Name),
			fmt.Sprintf("add_artist:%s:%s", artist.SpotifyID, artist.Name),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})

		// Limit to 5 results
		if i >= 4 {
			break
		}
	}

	// Add cancel button
	cancelButton := tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel_add_artist")
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{cancelButton})

	reply := tgbotapi.NewMessage(message.Chat.ID, "🔍 Found these artists. Please select one:")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	b.api.Send(reply)
}

// handleCallbackQuery processes inline keyboard callbacks
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Acknowledge the callback
	b.api.Request(tgbotapi.NewCallback(callback.ID, ""))

	data := callback.Data
	parts := strings.Split(data, ":")

	switch parts[0] {
	case "add_artist":
		if len(parts) >= 3 {
			spotifyID := parts[1]
			artistName := strings.Join(parts[2:], ":")
			b.confirmAddArtist(callback.Message.Chat.ID, callback.From.ID, spotifyID, artistName)
		}
	case "cancel_add_artist":
		delete(b.userStates, callback.From.ID)
		b.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, "❌ Artist addition cancelled.")
	case "remove_artist":
		if len(parts) >= 2 {
			artistName := strings.Join(parts[1:], ":")
			b.confirmRemoveArtist(callback.Message.Chat.ID, callback.From.ID, artistName)
		}
	}
}

// confirmAddArtist adds the selected artist to user's favorites
func (b *Bot) confirmAddArtist(chatID int64, userTelegramID int64, spotifyID, artistName string) {
	// Get or create user in database
	user, err := b.db.GetUserByTelegramID(userTelegramID)
	if err != nil {
		// User doesn't exist, try to create them
		// We don't have full user info here, so we'll use basic info
		user, err = b.db.CreateUser(userTelegramID, "", "", "")
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			b.sendMessage(chatID, "Sorry, there was an error setting up your account. Please try /start first.")
			return
		}
	}

	// Add artist to favorites
	err = b.db.AddUserArtist(user.ID, artistName, spotifyID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			b.sendMessage(chatID, fmt.Sprintf("🎤 %s is already in your favorites!", artistName))
		} else {
			log.Printf("Failed to add artist: %v", err)
			b.sendMessage(chatID, "Sorry, there was an error adding the artist.")
		}
		return
	}

	// Clear user state
	delete(b.userStates, userTelegramID)

	successText := fmt.Sprintf("✅ Added %s to your favorites!\n\nUse /add to add more artists or /playlist to generate a playlist.", artistName)
	b.sendMessage(chatID, successText)
}

// handleListArtists shows user's favorite artists
func (b *Bot) handleListArtists(message *tgbotapi.Message) {
	user, err := b.db.GetUserByTelegramID(message.From.ID)
	if err != nil {
		// User doesn't exist, auto-register them
		user, err = b.db.CreateUser(message.From.ID, message.From.UserName, message.From.FirstName, message.From.LastName)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			b.sendMessage(message.Chat.ID, "Sorry, there was an error setting up your account. Please try /start first.")
			return
		}
	}

	artists, err := b.db.GetUserArtists(user.ID)
	if err != nil {
		log.Printf("Failed to get user artists: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, there was an error retrieving your artists.")
		return
	}

	if len(artists) == 0 {
		b.sendMessage(message.Chat.ID, "🎤 You haven't added any favorite artists yet.\n\nUse /add to add your first artist!")
		return
	}

	var text strings.Builder
	text.WriteString("🎤 Your Favorite Artists:\n\n")
	for i, artist := range artists {
		text.WriteString(fmt.Sprintf("%d. %s\n", i+1, artist.ArtistName))
	}
	text.WriteString(fmt.Sprintf("\nTotal: %d artists", len(artists)))
	text.WriteString("\n\nUse /remove to remove artists or /playlist to generate a playlist.")

	b.sendMessage(message.Chat.ID, text.String())
}

// handleRemoveArtist handles removing an artist
func (b *Bot) handleRemoveArtist(message *tgbotapi.Message) {
	user, err := b.db.GetUserByTelegramID(message.From.ID)
	if err != nil {
		// User doesn't exist, auto-register them
		user, err = b.db.CreateUser(message.From.ID, message.From.UserName, message.From.FirstName, message.From.LastName)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			b.sendMessage(message.Chat.ID, "Sorry, there was an error setting up your account. Please try /start first.")
			return
		}
	}

	artists, err := b.db.GetUserArtists(user.ID)
	if err != nil {
		log.Printf("Failed to get user artists: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, there was an error retrieving your artists.")
		return
	}

	if len(artists) == 0 {
		b.sendMessage(message.Chat.ID, "🎤 You don't have any favorite artists to remove.")
		return
	}

	// Create keyboard with artists to remove
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, artist := range artists {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("❌ %s", artist.ArtistName),
			fmt.Sprintf("remove_artist:%s", artist.ArtistName),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, "🎤 Select an artist to remove:")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	b.api.Send(reply)
}

// confirmRemoveArtist removes an artist from favorites
func (b *Bot) confirmRemoveArtist(chatID int64, userTelegramID int64, artistName string) {
	user, err := b.db.GetUserByTelegramID(userTelegramID)
	if err != nil {
		// User doesn't exist, auto-register them (though this shouldn't normally happen)
		user, err = b.db.CreateUser(userTelegramID, "", "", "")
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			b.sendMessage(chatID, "Sorry, there was an error setting up your account. Please try /start first.")
			return
		}
	}

	err = b.db.RemoveUserArtist(user.ID, artistName)
	if err != nil {
		log.Printf("Failed to remove artist: %v", err)
		b.sendMessage(chatID, "Sorry, there was an error removing the artist.")
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Removed %s from your favorites.", artistName))
}

// handleGeneratePlaylist generates and sends a playlist
func (b *Bot) handleGeneratePlaylist(message *tgbotapi.Message) {
	user, err := b.db.GetUserByTelegramID(message.From.ID)
	if err != nil {
		// User doesn't exist, auto-register them
		user, err = b.db.CreateUser(message.From.ID, message.From.UserName, message.From.FirstName, message.From.LastName)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			b.sendMessage(message.Chat.ID, "Sorry, there was an error setting up your account. Please try /start first.")
			return
		}
	}

	artists, err := b.db.GetUserArtists(user.ID)
	if err != nil {
		log.Printf("Failed to get user artists: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, there was an error retrieving your artists.")
		return
	}

	if len(artists) == 0 {
		b.sendMessage(message.Chat.ID, "🎤 You need to add some favorite artists first!\n\nUse /add to add artists, then try generating a playlist.")
		return
	}

	// Send "generating" message
	b.sendMessage(message.Chat.ID, "🎵 Generating your personalized playlist... This may take a moment.")

	// Extract Spotify IDs
	var spotifyIDs []string
	for _, artist := range artists {
		if artist.SpotifyID != "" {
			spotifyIDs = append(spotifyIDs, artist.SpotifyID)
		}
	}

	// Generate playlist
	tracks, err := b.spotifyClient.GeneratePlaylistForArtists(spotifyIDs, 30)
	if err != nil {
		log.Printf("Failed to generate playlist: %v", err)
		b.sendMessage(message.Chat.ID, "Sorry, I couldn't generate a playlist right now. Please try again later.")
		return
	}

	if len(tracks) == 0 {
		b.sendMessage(message.Chat.ID, "Sorry, I couldn't find enough tracks to create a playlist. Try adding more artists!")
		return
	}

	// Format and send playlist
	playlistText := b.formatPlaylist(tracks)
	b.sendMessage(message.Chat.ID, playlistText)

	// Save playlist to database
	playlistJSON, _ := json.Marshal(tracks)
	b.db.SavePlaylist(user.ID, string(playlistJSON))
}

// formatPlaylist formats tracks into a readable message
func (b *Bot) formatPlaylist(tracks []spotify.Track) string {
	var text strings.Builder
	text.WriteString("🎵 Your Personalized Playlist:\n\n")

	for i, track := range tracks {
		text.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, track.Artist, track.Name))
		if track.SpotifyURL != "" {
			text.WriteString(fmt.Sprintf("   🎧 [Listen on Spotify](%s)\n", track.SpotifyURL))
		}
		text.WriteString("\n")
	}

	text.WriteString(fmt.Sprintf("🎼 Total: %d tracks\n", len(tracks)))
	text.WriteString("💫 Based on your favorite artists and similar recommendations")

	return text.String()
}

// SendDailyPlaylist generates and sends a daily playlist to a user
func (b *Bot) SendDailyPlaylist(userTelegramID int64) error {
	user, err := b.db.GetUserByTelegramID(userTelegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	artists, err := b.db.GetUserArtists(user.ID)
	if err != nil || len(artists) == 0 {
		// Skip users without artists
		return nil
	}

	// Extract Spotify IDs
	var spotifyIDs []string
	for _, artist := range artists {
		if artist.SpotifyID != "" {
			spotifyIDs = append(spotifyIDs, artist.SpotifyID)
		}
	}

	if len(spotifyIDs) == 0 {
		return nil
	}

	// Generate playlist
	tracks, err := b.spotifyClient.GeneratePlaylistForArtists(spotifyIDs, 25)
	if err != nil {
		return fmt.Errorf("failed to generate playlist: %w", err)
	}

	if len(tracks) == 0 {
		return nil
	}

	// Send daily playlist message
	playlistText := "🌅 Good morning! Here's your daily music discovery:\n\n" + b.formatPlaylist(tracks)
	b.sendMessage(userTelegramID, playlistText)

	// Save playlist
	playlistJSON, _ := json.Marshal(tracks)
	b.db.SavePlaylist(user.ID, string(playlistJSON))

	return nil
}

// sendMessage sends a text message to a chat
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	b.api.Send(msg)
}

// editMessage edits an existing message
func (b *Bot) editMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	b.api.Send(edit)
} 