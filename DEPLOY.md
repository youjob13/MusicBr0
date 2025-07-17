# Heroku Deployment Guide

This guide will help you deploy your Music Bot to Heroku with automatic deploys from GitHub.

## Prerequisites

1. [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed
2. [Git](https://git-scm.com/) installed
3. [GitHub account](https://github.com/)
4. [Heroku account](https://heroku.com/)
5. Your bot credentials (Telegram token, Spotify client ID & secret)

## Step 1: Prepare Your GitHub Repository

### 1.1 Create GitHub Repository

```bash
# Initialize git repository (if not already done)
git init

# Add all files
git add .

# Commit changes
git commit -m "Initial commit: Music Bot for Heroku deployment"

# Create repository on GitHub and add remote
git remote add origin https://github.com/YOUR_USERNAME/MusicBr0.git

# Push to GitHub
git push -u origin main
```

### 1.2 Create Release Branch

```bash
# Create and switch to release branch
git checkout -b release

# Push release branch to GitHub
git push -u origin release
```

## Step 2: Create Heroku Application

### 2.1 Using Heroku CLI

```bash
# Login to Heroku
heroku login

# Create new Heroku app (replace 'your-music-bot' with your desired app name)
heroku create your-music-bot --region us

# Or use the one-click deploy button (see Step 2.2)
```

### 2.2 One-Click Deploy (Alternative)

Click this button to deploy directly from GitHub:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/YOUR_USERNAME/MusicBr0)

**Note:** Replace `YOUR_USERNAME` with your GitHub username in the app.json file.

## Step 3: Configure Environment Variables

Set your environment variables on Heroku:

```bash
# Using Heroku CLI
heroku config:set TELEGRAM_TOKEN="your_telegram_bot_token" --app your-music-bot
heroku config:set SPOTIFY_CLIENT_ID="your_spotify_client_id" --app your-music-bot
heroku config:set SPOTIFY_CLIENT_SECRET="your_spotify_client_secret" --app your-music-bot
heroku config:set DATABASE_PATH="/tmp/music_bot.db" --app your-music-bot
heroku config:set LOG_LEVEL="info" --app your-music-bot
```

**Or using Heroku Dashboard:**

1. Go to your [Heroku Dashboard](https://dashboard.heroku.com/apps)
2. Click on your app
3. Go to **Settings** tab
4. Click **Reveal Config Vars**
5. Add the environment variables:

| Key                     | Value                          |
| ----------------------- | ------------------------------ |
| `TELEGRAM_TOKEN`        | Your bot token from @BotFather |
| `SPOTIFY_CLIENT_ID`     | Your Spotify app client ID     |
| `SPOTIFY_CLIENT_SECRET` | Your Spotify app client secret |
| `DATABASE_PATH`         | `/tmp/music_bot.db`            |
| `LOG_LEVEL`             | `info`                         |

## Step 4: Set Up Automatic Deploys from GitHub

### 4.1 Connect GitHub Repository

1. Go to your [Heroku Dashboard](https://dashboard.heroku.com/apps)
2. Click on your app
3. Go to **Deploy** tab
4. In **Deployment method** section, click **GitHub**
5. Click **Connect to GitHub** and authorize Heroku
6. Search for your repository name and click **Connect**

### 4.2 Enable Automatic Deploys

1. In the **Automatic deploys** section:
   - Choose the `release` branch
   - ✅ Check **Wait for CI to pass before deploy** (if you have CI setup)
   - Click **Enable Automatic Deploys**

### 4.3 Deploy Manually (First Time)

1. In the **Manual deploy** section:
   - Choose the `release` branch
   - Click **Deploy Branch**

## Step 5: Verify Deployment

### 5.1 Check Application Logs

```bash
# View live logs
heroku logs --tail --app your-music-bot

# You should see output like:
# 2025/01/15 10:00:00 Authorized on account YourBotName
# 2025/01/15 10:00:00 Scheduler started - daily playlists will be sent at 9 AM
# 2025/01/15 10:00:00 Music bot started successfully!
```

### 5.2 Test Bot Functionality

1. Open Telegram and find your bot
2. Send `/start` command
3. Try other commands like `/help`, `/add`, etc.

## Step 6: Workflow for Future Updates

### 6.1 Development Workflow

```bash
# Make changes on main branch
git checkout main

# ... make your changes ...

# Commit changes
git add .
git commit -m "Add new feature"
git push origin main
```

### 6.2 Release Workflow

```bash
# Merge main into release branch for deployment
git checkout release
git merge main
git push origin release

# This will automatically trigger deployment to Heroku
```

## Important Notes

### Database Considerations

- Heroku's filesystem is ephemeral, so SQLite data is lost on app restart
- For production use, consider upgrading to [Heroku Postgres](https://devcenter.heroku.com/articles/heroku-postgresql)
- Current setup is suitable for testing and development

### App Scaling

```bash
# Scale to 1 dyno (free tier)
heroku ps:scale web=1 --app your-music-bot

# Check dyno status
heroku ps --app your-music-bot
```

### Free Tier Limitations

- App sleeps after 30 minutes of inactivity
- Limited to 1000 dyno hours per month
- Consider upgrading to Hobby tier ($7/month) for 24/7 operation

## Troubleshooting

### Common Issues

1. **App crashing on startup**

   ```bash
   heroku logs --tail --app your-music-bot
   ```

   Check for missing environment variables or configuration errors.

2. **Build failures**

   - Ensure go.mod and go.sum are committed
   - Check Go version compatibility
   - Verify all dependencies are properly declared

3. **Bot not responding**
   - Verify Telegram token is correct
   - Check Heroku logs for errors
   - Ensure app is not sleeping (upgrade to Hobby tier)

### Useful Heroku Commands

```bash
# Restart app
heroku restart --app your-music-bot

# Check app info
heroku info --app your-music-bot

# Open app in browser (for debugging)
heroku open --app your-music-bot

# Access Heroku bash shell
heroku run bash --app your-music-bot
```

## Cost Optimization

### Free Tier Usage

- Monitor dyno hours: [Heroku Dashboard](https://dashboard.heroku.com/account/billing)
- App sleeps after 30 minutes of inactivity (saves hours)
- 1000 free dyno hours per month (approximately 23 days)

### Upgrading

Consider upgrading if you need:

- 24/7 operation: Hobby tier ($7/month)
- Better performance: Standard tier ($25/month)
- Persistent database: Postgres add-on ($9/month)

---

**Your Music Bot is now deployed to Heroku with automatic deploys! 🎉**

Every time you push to the `release` branch, Heroku will automatically deploy the updated version of your bot.
