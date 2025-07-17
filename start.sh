#!/bin/bash

# Music Bot Quick Start Script
echo "🎵 Music Bot Quick Start"
echo "======================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp env_example.txt .env
    echo ""
    echo "⚠️  Please edit .env file with your tokens before continuing!"
    echo ""
    echo "Required values:"
    echo "- TELEGRAM_TOKEN (from @BotFather)"
    echo "- SPOTIFY_CLIENT_ID (from Spotify Developer Dashboard)"
    echo "- SPOTIFY_CLIENT_SECRET (from Spotify Developer Dashboard)"
    echo ""
    echo "Run this script again after setting up your .env file."
    exit 1
fi

# Source environment variables
echo "📖 Loading environment variables..."
source .env

# Check required variables
if [ -z "$TELEGRAM_TOKEN" ] || [ -z "$SPOTIFY_CLIENT_ID" ] || [ -z "$SPOTIFY_CLIENT_SECRET" ]; then
    echo "❌ Missing required environment variables in .env file"
    echo "Please check your .env file and set all required values."
    exit 1
fi

echo "✅ Environment variables loaded"

# Create data directory
mkdir -p data
echo "📁 Created data directory"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.19+ and try again."
    exit 1
fi

echo "🔧 Building application..."
go build -o music-bot

if [ $? -eq 0 ]; then
    echo "✅ Build successful"
    echo ""
    echo "🚀 Starting Music Bot..."
    echo "Press Ctrl+C to stop"
    echo ""
    
    # Export environment variables for the application
    export TELEGRAM_TOKEN
    export SPOTIFY_CLIENT_ID
    export SPOTIFY_CLIENT_SECRET
    export DATABASE_PATH="data/music_bot.db"
    export LOG_LEVEL="info"
    
    ./music-bot
else
    echo "❌ Build failed. Please check the error messages above."
    exit 1
fi 