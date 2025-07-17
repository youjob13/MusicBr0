@echo off
echo 🎵 Music Bot Quick Start
echo =======================

REM Check if .env file exists
if not exist ".env" (
    echo 📝 Creating .env file from template...
    copy env_example.txt .env >nul
    echo.
    echo ⚠️  Please edit .env file with your tokens before continuing!
    echo.
    echo Required values:
    echo - TELEGRAM_TOKEN (from @BotFather)
    echo - SPOTIFY_CLIENT_ID (from Spotify Developer Dashboard)
    echo - SPOTIFY_CLIENT_SECRET (from Spotify Developer Dashboard)
    echo.
    echo Run this script again after setting up your .env file.
    pause
    exit /b 1
)

REM Load environment variables from .env file
echo 📖 Loading environment variables...
for /f "usebackq tokens=1,2 delims==" %%i in (".env") do (
    if not "%%i"=="" if not "%%j"=="" (
        set "%%i=%%j"
    )
)

REM Check required variables
if "%TELEGRAM_TOKEN%"=="" (
    echo ❌ TELEGRAM_TOKEN not set in .env file
    pause
    exit /b 1
)
if "%SPOTIFY_CLIENT_ID%"=="" (
    echo ❌ SPOTIFY_CLIENT_ID not set in .env file
    pause
    exit /b 1
)
if "%SPOTIFY_CLIENT_SECRET%"=="" (
    echo ❌ SPOTIFY_CLIENT_SECRET not set in .env file
    pause
    exit /b 1
)

echo ✅ Environment variables loaded

REM Create data directory
if not exist "data" mkdir data
echo 📁 Created data directory

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go is not installed. Please install Go 1.19+ and try again.
    pause
    exit /b 1
)

echo 🔧 Building application...
go build -o music-bot.exe

if %errorlevel% equ 0 (
    echo ✅ Build successful
    echo.
    echo 🚀 Starting Music Bot...
    echo Press Ctrl+C to stop
    echo.
    
    REM Set additional environment variables
    set DATABASE_PATH=data\music_bot.db
    set LOG_LEVEL=info
    
    music-bot.exe
) else (
    echo ❌ Build failed. Please check the error messages above.
    pause
    exit /b 1
) 