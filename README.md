# 🎵 Music Bot - Telegram Music Recommendation Bot

A Telegram bot that generates personalized music playlists based on your favorite artists using the Spotify API.

## 🎯 Features

- **Artist Management**: Add and remove your favorite artists
- **Smart Recommendations**: Discover new music based on your taste using Spotify's recommendation engine
- **Daily Playlists**: Automatically receive personalized playlists every morning
- **Artist Discovery**: Find similar artists and explore new music
- **User-Friendly Interface**: Easy-to-use Telegram commands and inline keyboards

## 🏗️ Architecture

- **Go Backend**: Fast and efficient server application
- **SQLite Database**: Lightweight database for user data and preferences
- **Spotify API**: Music data and recommendations
- **Telegram Bot API**: User interaction and messaging
- **Cron Scheduler**: Automated daily playlist delivery

## 🚀 Quick Start

### Prerequisites

1. **Go 1.19+** installed on your system
2. **Telegram Bot Token** from [@BotFather](https://t.me/botfather)
3. **Spotify Developer Account** and app credentials

### Setup Instructions

#### 1. Create Telegram Bot

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` command
3. Follow the prompts to create your bot
4. Save the bot token for later

#### 2. Create Spotify Application

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/)
2. Click "Create App"
3. Fill in app details:
   - **App Name**: Your bot name
   - **App Description**: Music recommendation bot
   - **Website**: Optional
   - **Redirect URI**: `http://localhost:8080/callback` (not used but required)
4. Save your **Client ID** and **Client Secret**

#### 3. Clone and Setup Project

```bash
# Clone the repository (or download the files)
git clone <your-repo-url>
cd music-bot

# Install dependencies
go mod download

# Build the application
go build -o music-bot
```

#### 4. Configure Environment Variables

Create a `.env` file or set environment variables:

```bash
# Copy the example configuration
cp env_example.txt .env

# Edit .env with your values:
TELEGRAM_TOKEN=your_telegram_bot_token_here
SPOTIFY_CLIENT_ID=your_spotify_client_id_here
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret_here
DATABASE_PATH=music_bot.db
LOG_LEVEL=info
```

**Windows (PowerShell):**

```powershell
$env:TELEGRAM_TOKEN="your_telegram_bot_token_here"
$env:SPOTIFY_CLIENT_ID="your_spotify_client_id_here"
$env:SPOTIFY_CLIENT_SECRET="your_spotify_client_secret_here"
```

**Linux/Mac:**

```bash
export TELEGRAM_TOKEN="your_telegram_bot_token_here"
export SPOTIFY_CLIENT_ID="your_spotify_client_id_here"
export SPOTIFY_CLIENT_SECRET="your_spotify_client_secret_here"
```

#### 5. Run the Bot

```bash
# Run directly
./music-bot

# Or with Go
go run main.go
```

You should see output like:

```
2024/01/15 10:00:00 Authorized on account YourBotName
2024/01/15 10:00:00 Scheduler started - daily playlists will be sent at 9 AM
2024/01/15 10:00:00 Music bot started successfully!
```

## 🎮 Bot Commands

Once your bot is running, users can interact with it using these commands:

- `/start` - Start using the bot and register
- `/help` - Show help message and available commands
- `/add` - Add a favorite artist
- `/list` - Show your favorite artists
- `/remove` - Remove an artist from favorites
- `/playlist` - Generate a new playlist instantly

## 🔧 Bot Usage Flow

1. **Start the bot** with `/start`
2. **Add artists** with `/add` and search for your favorites
3. **Generate playlists** with `/playlist` to get immediate recommendations
4. **Receive daily playlists** automatically at 9 AM every day

## 📁 Project Structure

```
music-bot/
├── main.go                 # Application entry point
├── internal/
│   ├── bot/                # Telegram bot logic
│   │   └── bot.go
│   ├── config/             # Configuration management
│   │   └── config.go
│   ├── database/           # Database operations
│   │   └── database.go
│   ├── scheduler/          # Cron scheduler
│   │   └── scheduler.go
│   └── spotify/            # Spotify API client
│       └── spotify.go
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md               # This file
└── env_example.txt         # Environment variables example
```

## 🐳 Docker Deployment

### Build Docker Image

Create a `Dockerfile`:

```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .

CMD ["./main"]
```

Build and run:

```bash
# Build the image
docker build -t music-bot .

# Run with environment variables
docker run -d \
  --name music-bot \
  -e TELEGRAM_TOKEN="your_token" \
  -e SPOTIFY_CLIENT_ID="your_client_id" \
  -e SPOTIFY_CLIENT_SECRET="your_client_secret" \
  -v $(pwd)/data:/root/data \
  music-bot
```

## ☁️ Cloud Deployment

### Heroku

1. Create a new Heroku app
2. Set environment variables in Heroku dashboard
3. Deploy using Git or GitHub integration

### VPS Deployment

1. Copy files to your server
2. Set environment variables
3. Run as a systemd service:

```bash
# Create service file
sudo nano /etc/systemd/system/music-bot.service

# Add service configuration
[Unit]
Description=Music Bot
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/music-bot
ExecStart=/path/to/music-bot/music-bot
Restart=always
Environment=TELEGRAM_TOKEN=your_token
Environment=SPOTIFY_CLIENT_ID=your_client_id
Environment=SPOTIFY_CLIENT_SECRET=your_client_secret

[Install]
WantedBy=multi-user.target

# Enable and start service
sudo systemctl enable music-bot
sudo systemctl start music-bot
```

## 🔧 Configuration Options

| Environment Variable    | Description                              | Default        | Required |
| ----------------------- | ---------------------------------------- | -------------- | -------- |
| `TELEGRAM_TOKEN`        | Bot token from BotFather                 | -              | ✅       |
| `SPOTIFY_CLIENT_ID`     | Spotify app client ID                    | -              | ✅       |
| `SPOTIFY_CLIENT_SECRET` | Spotify app client secret                | -              | ✅       |
| `DATABASE_PATH`         | SQLite database file path                | `music_bot.db` | ❌       |
| `LOG_LEVEL`             | Logging level (debug, info, warn, error) | `info`         | ❌       |

## 📊 Database Schema

The bot uses SQLite with the following tables:

- **users**: Telegram user information
- **user_artists**: User's favorite artists
- **playlists**: Generated playlist history

## 🐛 Troubleshooting

### Common Issues

1. **"Failed to load configuration" error**

   - Check that all required environment variables are set
   - Verify your tokens are correct

2. **"Failed to initialize Spotify client" error**

   - Verify your Spotify credentials
   - Check your internet connection

3. **Bot not responding**

   - Check bot token is correct
   - Verify bot is not blocked by user
   - Check logs for error messages

4. **"No artists found" when searching**
   - Check Spotify API is accessible
   - Try different artist names
   - Verify Spotify credentials

### Logs

Check application logs for detailed error information:

```bash
# If running directly
./music-bot 2>&1 | tee bot.log

# If running with systemd
sudo journalctl -u music-bot -f
```

## 🛠️ Development

### Adding New Features

1. **Add new commands** in `internal/bot/bot.go`
2. **Extend database schema** in `internal/database/database.go`
3. **Add Spotify features** in `internal/spotify/spotify.go`

### Testing

```bash
# Run tests
go test ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o music-bot-linux
GOOS=windows GOARCH=amd64 go build -o music-bot-windows.exe
```

## 📈 Performance Tips

- **Database**: Regular SQLite database maintenance
- **API Limits**: Respect Spotify rate limits (built-in)
- **Memory**: Monitor memory usage for large user bases
- **Concurrency**: Bot handles multiple users concurrently

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Spotify Web API](https://developer.spotify.com/documentation/web-api/)
- [zmb3/spotify](https://github.com/zmb3/spotify) - Go Spotify library
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - Go Telegram library

## 📞 Support

If you encounter any issues or have questions:

1. Check the troubleshooting section above
2. Review the logs for error messages
3. Create an issue on GitHub
4. Check Telegram Bot API and Spotify API documentation

---

**Made with ❤️ and Go** 🎵🤖
