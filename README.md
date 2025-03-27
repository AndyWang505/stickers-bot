# Discord Sticker Bot

A simple Discord bot for managing and sharing stickers in your server.

## Quick Start

1. **Installation**
```bash
# Clone the repository
git clone https://github.com/yourusername/stickers-bot.git
cd stickers-bot

# Install dependencies
go mod download

# Copy and edit config
cp config.json.example config.json
# Edit config.json with your Discord Bot Token
```

2. **Run the bot**
```bash
go run main.go
```

## Usage

### Basic Commands

```
!![name] - Get a sticker
!![name] + upload image - Add a new sticker
!!list - Show all stickers
!!help - Show help message
```

### Examples

```
!!happy - Get the "happy" sticker
!!happy + upload image - Add a new "happy" sticker
!!list - Show all stickers
```

## Configuration

Edit `config.json` with your settings:
```json
{
    "token": "your-bot-token",
    "resources_dir": "resources",
    "stickers_file": "stickers.json",
    "logs_dir": "logs",
    "command_prefix": "!!"
}
``` 