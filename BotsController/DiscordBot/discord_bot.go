package DiscordBot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token     string
	BotPrefix string = "!!"
)

func Start() {
	// Get token from environment variable
	Token = os.Getenv("DISCORD_BOT_TOKEN")
	if Token == "" {
		log.Fatal("No token provided. Set DISCORD_BOT_TOKEN environment variable.")
	}

	// Initialize sticker manager
	if err := InitStickerManager(); err != nil {
		log.Printf("Warning: Failed to initialize sticker manager: %v", err)
	}

	// Create Discord session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	// Add message handler
	dg.AddHandler(messageCreate)

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if message starts with prefix
	if !strings.HasPrefix(m.Content, BotPrefix) {
		return
	}

	// Remove prefix and split command
	content := strings.TrimPrefix(m.Content, BotPrefix)
	args := strings.Fields(content)
	if len(args) == 0 {
		return
	}

	command := args[0]
	args = args[1:]

	// Handle commands
	switch command {
	case "sticker":
		handleStickerCommand(s, m, args)
	case "list":
		handleListCommand(s, m)
	case "help":
		handleHelpCommand(s, m)
	case "debug":
		handleDebugCommand(s, m)
	case "reload":
		handleReloadCommand(s, m)
	default:
		// Try to get sticker directly by name
		handleDirectStickerCommand(s, m, command)
	}
}

func handleStickerCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a sticker name")
		return
	}

	name := args[0]
	sm := GetStickerManager()

	// Check for attachments
	if len(m.Attachments) > 0 {
		// Handle uploaded image
		err := sm.HandleUploadedImage(name, m.Attachments[0], m.Author.Username)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully added sticker: %s", name))
		return
	}

	// Get existing sticker
	sticker, err := sm.GetSticker(name)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	// Send sticker
	s.ChannelMessageSend(m.ChannelID, sticker.URL)
}

func handleListCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	sm := GetStickerManager()
	stickers := sm.ListStickers()

	if len(stickers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No stickers available")
		return
	}

	var message strings.Builder
	message.WriteString("Available stickers:\n")
	for _, sticker := range stickers {
		message.WriteString(fmt.Sprintf("- %s\n", sticker.Name))
	}

	s.ChannelMessageSend(m.ChannelID, message.String())
}

func handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	helpText := `Available commands:
!![name] - Get a sticker by name
!![name] + upload image - Add a new sticker
!!list - List all stickers
!!help - Show this help message
!!debug - Show debug information
!!reload - Reload stickers from file

Examples:
!!happy - Get the "happy" sticker
!!happy + upload image - Add a new "happy" sticker
!!list - Show all available stickers`

	s.ChannelMessageSend(m.ChannelID, helpText)
}

func handleDebugCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	sm := GetStickerManager()
	
	// Get absolute path of sticker file
	absPath, _ := filepath.Abs(sm.filePath)
	
	// Check if file exists
	fileInfo, err := os.Stat(sm.filePath)
	fileStatus := "File not found"
	if err == nil {
		fileStatus = fmt.Sprintf("File exists, size: %d bytes, modified: %s", 
			fileInfo.Size(), fileInfo.ModTime().Format(time.RFC3339))
	}
	
	// Get current stickers in memory
	stickers := sm.ListStickers()
	stickerList := "None"
	if len(stickers) > 0 {
		var sb strings.Builder
		for _, s := range stickers {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", s.Name, s.URL))
		}
		stickerList = sb.String()
	}
	
	// Get working directory
	wd, _ := os.Getwd()
	
	debugInfo := fmt.Sprintf("Debug Information:\n"+
		"- Working Directory: %s\n"+
		"- Sticker File Path: %s\n"+
		"- File Status: %s\n"+
		"- Stickers in Memory: %d\n"+
		"- Sticker List:\n%s",
		wd, absPath, fileStatus, len(stickers), stickerList)
	
	s.ChannelMessageSend(m.ChannelID, debugInfo)
}

func handleReloadCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	sm := GetStickerManager()
	
	// Re-initialize the sticker manager
	count := len(sm.stickers)
	err := InitStickerManager()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error reloading stickers: %v", err))
		return
	}
	
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Stickers reloaded. Previous count: %d, New count: %d", 
		count, len(sm.stickers)))
}

// New function to handle direct sticker retrieval
func handleDirectStickerCommand(s *discordgo.Session, m *discordgo.MessageCreate, name string) {
	sm := GetStickerManager()
	
	// Check for attachments (adding new sticker)
	if len(m.Attachments) > 0 {
		err := sm.HandleUploadedImage(name, m.Attachments[0], m.Author.Username)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully added sticker: %s", name))
		return
	}

	// Try to get existing sticker
	sticker, err := sm.GetSticker(name)
	if err != nil {
		// If sticker not found, send help message
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sticker '%s' not found. Use !!help to see available commands.", name))
		return
	}

	// Send sticker
	s.ChannelMessageSend(m.ChannelID, sticker.URL)
} 