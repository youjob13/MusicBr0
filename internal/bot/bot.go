package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"music-bot/internal/database"
	"music-bot/internal/spotify"
	"net/http"
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

	welcomeText := fmt.Sprintf(`🎵 Welcome to Music Bot!

Hello %s! I'm your personal music assistant.

What I can do:
• Add your favorite artists (/add)
• Generate personalized playlists (/playlist)
• Preview tracks directly in Telegram
• Send daily music recommendations

Ready to start? Use /add to add your first artist!

For help, use /help anytime.`, message.From.FirstName)

	msg := tgbotapi.NewMessage(message.Chat.ID, welcomeText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

// sendHelp sends the help message
func (b *Bot) sendHelp(chatID int64) {
	helpText := `🎵 Music Bot Commands

Available commands:
/start - Get started with the bot
/help - Show this help message
/add - Add a favorite artist
/list - View your favorite artists
/remove - Remove an artist from favorites
/playlist - Generate a personalized playlist

How it works:
1. Add your favorite artists with /add
2. Generate playlists with /playlist
3. Preview tracks directly in Telegram
4. Discover new music based on your taste

Note: Preview availability depends on Spotify's data - not all tracks have 30-second previews available.`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

// handleAddArtist handles adding a favorite artist
func (b *Bot) handleAddArtist(message *tgbotapi.Message) {
	b.userStates[message.From.ID] = "waiting_artist_name"
	
	promptText := `🎤✨ **ADD FAVORITE ARTIST** ✨🎤

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 **ARTIST SEARCH:**

Type the name of any artist you love!

*Examples:*
• Taylor Swift
• Drake  
• The Beatles
• Billie Eilish

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *I'll search Spotify and show you options to choose from*`

	msg := tgbotapi.NewMessage(message.Chat.ID, promptText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

// handleArtistSearch processes artist search and selection
func (b *Bot) handleArtistSearch(message *tgbotapi.Message) {
	artistName := strings.TrimSpace(message.Text)
	if artistName == "" {
		errorText := `❌ **INVALID INPUT**

Please enter a valid artist name.

Try again with an artist you love! 🎵`

		msg := tgbotapi.NewMessage(message.Chat.ID, errorText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	// Send searching message
	searchingText := fmt.Sprintf(`🔍✨ **SEARCHING FOR:** *%s*

⏳ *Scanning Spotify database...*
🎵 *Finding the perfect match...*`, artistName)

	searchMsg := tgbotapi.NewMessage(message.Chat.ID, searchingText)
	searchMsg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(searchMsg)

	// Search for artists on Spotify
	artists, err := b.spotifyClient.SearchArtists(artistName, 5)
	if err != nil {
		log.Printf("Failed to search artists: %v", err)
		errorText := `❌ **SEARCH ERROR**

🚫 *Couldn't search for artists right now*

Please try again in a moment! 🔄`

		msg := tgbotapi.NewMessage(message.Chat.ID, errorText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	if len(artists) == 0 {
		noResultsText := fmt.Sprintf(`🚫 **NO RESULTS FOUND**

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 *Searched for:* **%s**

💡 **TRY:**
• Check spelling
• Use the artist's main name
• Try different variations

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, artistName)

		msg := tgbotapi.NewMessage(message.Chat.ID, noResultsText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	// Create beautiful artist selection interface
	resultsText := fmt.Sprintf(`🎯✨ **SEARCH RESULTS** ✨🎯

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍 *Found %d artists for:* **%s**

👇 *Choose the right one:*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, len(artists), artistName)

	// Create inline keyboard with artist options
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, artist := range artists {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🎤 %s", artist.Name),
			fmt.Sprintf("add_artist:%s:%s", artist.SpotifyID, artist.Name),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})

		// Limit to 5 results
		if i >= 4 {
			break
		}
	}

	// Add cancel button
	cancelButton := tgbotapi.NewInlineKeyboardButtonData("❌ CANCEL SEARCH", "cancel_add_artist")
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{cancelButton})

	reply := tgbotapi.NewMessage(message.Chat.ID, resultsText)
	reply.ParseMode = tgbotapi.ModeMarkdown
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
	case "preview":
		if len(parts) >= 3 {
			previewURL := parts[1]
			trackInfo := strings.Join(parts[2:], ":")
			b.sendAudioPreview(callback.Message.Chat.ID, previewURL, trackInfo)
		}
	case "no_preview":
		// Handle "No Preview" button click
		b.api.Request(tgbotapi.NewCallbackWithAlert(callback.ID, "🚫 Sorry, no preview is available for this track. Try opening it in Spotify instead!"))
	}
}

// sendAudioPreview downloads and sends a 30-second audio preview
func (b *Bot) sendAudioPreview(chatID int64, previewURL, trackInfo string) {
	if previewURL == "" {
		msg := tgbotapi.NewMessage(chatID, "🚫 No preview available for this track. Try opening it in Spotify!")
		b.api.Send(msg)
		return
	}

	// Send loading message
	loadingText := fmt.Sprintf("⏳ Loading preview for %s...", trackInfo)
	loadingMsg := tgbotapi.NewMessage(chatID, loadingText)
	b.api.Send(loadingMsg)

	// Download the audio preview
	resp, err := http.Get(previewURL)
	if err != nil {
		log.Printf("Failed to download preview: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Failed to download preview. Try opening the track in Spotify!")
		b.api.Send(msg)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bad response downloading preview: %d", resp.StatusCode)
		msg := tgbotapi.NewMessage(chatID, "❌ Preview temporarily unavailable. Try opening the track in Spotify!")
		b.api.Send(msg)
		return
	}

	// Read the audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read audio data: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Error processing the audio file. Try opening the track in Spotify!")
		b.api.Send(msg)
		return
	}

	// Send as audio file with simple caption
	audioMsg := tgbotapi.NewAudio(chatID, tgbotapi.FileBytes{
		Name:  "preview.mp3",
		Bytes: audioData,
	})
	audioMsg.Caption = fmt.Sprintf("🎵 %s\n⏱️ 30-second preview", trackInfo)
	audioMsg.Duration = 30 // 30-second preview

	_, err = b.api.Send(audioMsg)
	if err != nil {
		log.Printf("Failed to send audio: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Couldn't send the preview. Try opening the track in Spotify!")
		b.api.Send(msg)
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
			msg := tgbotapi.NewMessage(chatID, "❌ Error setting up your account. Please try /start first!")
			b.api.Send(msg)
			return
		}
	}

	// Add artist to favorites
	err = b.db.AddUserArtist(user.ID, artistName, spotifyID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔄 %s is already in your favorites! Try adding a different artist.", artistName))
			b.api.Send(msg)
		} else {
			log.Printf("Failed to add artist: %v", err)
			msg := tgbotapi.NewMessage(chatID, "❌ Error adding the artist. Please try again!")
			b.api.Send(msg)
		}
		return
	}

	// Clear user state
	delete(b.userStates, userTelegramID)

	successText := fmt.Sprintf(`✅ Artist Added!

🎤 %s has been added to your favorites!

What's next?
• Use /add to add more artists
• Use /list to view your collection  
• Use /playlist to generate music!`, artistName)

	msg := tgbotapi.NewMessage(chatID, successText)
	b.api.Send(msg)
}

// handleListArtists shows user's favorite artists
func (b *Bot) handleListArtists(message *tgbotapi.Message) {
	user, err := b.db.GetUserByTelegramID(message.From.ID)
	if err != nil {
		// User doesn't exist, auto-register them
		user, err = b.db.CreateUser(message.From.ID, message.From.UserName, message.From.FirstName, message.From.LastName)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			errorText := `❌ **ACCOUNT ERROR**

🚫 *Error setting up your account*

Please try /start first! 🚀`

			msg := tgbotapi.NewMessage(message.Chat.ID, errorText)
			msg.ParseMode = tgbotapi.ModeMarkdown
			b.api.Send(msg)
			return
		}
	}

	artists, err := b.db.GetUserArtists(user.ID)
	if err != nil {
		log.Printf("Failed to get user artists: %v", err)
		errorText := `❌ **RETRIEVAL ERROR**

🚫 *Error retrieving your artists*

Please try again! 🔄`

		msg := tgbotapi.NewMessage(message.Chat.ID, errorText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	if len(artists) == 0 {
		emptyText := `🎤✨ **YOUR ARTIST COLLECTION** ✨🎤

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📭 **COLLECTION EMPTY**

You haven't added any favorite artists yet!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 **GET STARTED:**

Use /add to add your first artist and start building your music profile!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

		msg := tgbotapi.NewMessage(message.Chat.ID, emptyText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString(`🎤✨ **YOUR ARTIST COLLECTION** ✨🎤

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

`)

	for i, artist := range artists {
		text.WriteString(fmt.Sprintf("🎵 **%d.** %s\n", i+1, artist.ArtistName))
	}
	
	text.WriteString(fmt.Sprintf(`
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 **Total Artists:** %d
🎶 **Status:** Ready for playlists!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚀 **ACTIONS:**

🗑️ Use /remove to remove artists
🎵 Use /playlist to generate music!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`, len(artists)))

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
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
		emptyArtistsText := `🎤 **ARTIST COLLECTION EMPTY**

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 *You need favorite artists to generate playlists!*

🚀 **GET STARTED:**

Use /add to add artists, then try generating a playlist.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

		msg := tgbotapi.NewMessage(message.Chat.ID, emptyArtistsText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	// Send beautiful generating message
	generatingText := `🎵✨ **GENERATING YOUR PLAYLIST** ✨🎵

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🤖 *AI is analyzing your music taste...*
🎯 *Finding perfect track combinations...*
🔍 *Discovering similar artists...*
🎶 *Creating 30 personalized tracks...*

⏳ *This may take a moment...*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	generatingMsg := tgbotapi.NewMessage(message.Chat.ID, generatingText)
	generatingMsg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(generatingMsg)

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
		errorText := `❌ **PLAYLIST GENERATION FAILED**

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🚫 *Couldn't generate a playlist right now*

💡 **POSSIBLE CAUSES:**
• Spotify API temporarily unavailable
• Network connection issues
• Service maintenance

🔄 *Please try again in a moment!*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

		msg := tgbotapi.NewMessage(message.Chat.ID, errorText)
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	if len(tracks) == 0 {
		noTracksText := `🚫 No tracks found

Couldn't find enough tracks to create a playlist.

Suggestions:
• Add more artists to your collection with /add
• Try different artist genres
• Check your artist spellings

Use /add to expand your music taste!`

		msg := tgbotapi.NewMessage(message.Chat.ID, noTracksText)
		b.api.Send(msg)
		return
	}

	// Send playlist with preview functionality
	b.sendPlaylistWithPreviews(message.Chat.ID, tracks)

	// Save playlist to database
	playlistJSON, _ := json.Marshal(tracks)
	b.db.SavePlaylist(user.ID, string(playlistJSON))
}

// sendPlaylistWithPreviews sends a playlist with individual track cards and preview buttons
func (b *Bot) sendPlaylistWithPreviews(chatID int64, tracks []spotify.Track) {
	// Send simple playlist header
	headerText := fmt.Sprintf(`🎵 Your Personalized Playlist

%d tracks curated based on your music taste
Tap ▶️ to play 30-second previews directly in Telegram!`, len(tracks))

	msg := tgbotapi.NewMessage(chatID, headerText)
	b.api.Send(msg)

	// Send each track with simple design
	for i, track := range tracks {
		// Create simple track display
		trackText := fmt.Sprintf(`🎵 Track %d

🎤 Artist: %s
🎶 Song: %s
💿 Album: %s`,
			i+1,
			track.Artist,
			track.Name,
			track.Album)

		// Create simple inline keyboard
		var keyboard [][]tgbotapi.InlineKeyboardButton
		var row1 []tgbotapi.InlineKeyboardButton
		
		if track.PreviewURL != "" {
			previewBtn := tgbotapi.NewInlineKeyboardButtonData(
				"▶️ Play Preview",
				fmt.Sprintf("preview:%s:%s - %s", track.PreviewURL, track.Artist, track.Name),
			)
			row1 = append(row1, previewBtn)
		} else {
			// Add a disabled-style button when no preview
			noPreviewBtn := tgbotapi.NewInlineKeyboardButtonData(
				"🚫 No Preview",
				"no_preview",
			)
			row1 = append(row1, noPreviewBtn)
		}
		
		if track.SpotifyURL != "" {
			spotifyBtn := tgbotapi.NewInlineKeyboardButtonURL("🎵 Open in Spotify", track.SpotifyURL)
			row1 = append(row1, spotifyBtn)
		}
		
		keyboard = append(keyboard, row1)

		// Send track message
		msg := tgbotapi.NewMessage(chatID, trackText)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		
		b.api.Send(msg)
	}

	// Send simple completion message
	footerText := `🎉 Playlist complete! Enjoy your music discovery.`
	footerMsg := tgbotapi.NewMessage(chatID, footerText)
	b.api.Send(footerMsg)
}

// formatPlaylist formats tracks into a readable message (legacy format, kept for daily playlists)
func (b *Bot) formatPlaylist(tracks []spotify.Track) string {
	var text strings.Builder
	text.WriteString("🎵 Your Personalized Playlist:\n\n")

	for i, track := range tracks {
		text.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, track.Artist, track.Name))
		if track.SpotifyURL != "" {
			text.WriteString(fmt.Sprintf("   🎧 [Listen on Spotify](%s)\n", track.SpotifyURL))
		}
		if track.PreviewURL != "" {
			text.WriteString("   ▶️ Preview available\n")
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