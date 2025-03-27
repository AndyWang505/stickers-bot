package DiscordBot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Sticker struct {
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Tags      []string `json:"tags"`
	AddedBy   string   `json:"added_by"`
	AddedAt   string   `json:"added_at"`
	LocalPath string   `json:"local_path,omitempty"`
}

type StickerManager struct {
	stickers map[string]Sticker
	mutex    sync.RWMutex
	filePath string
}

var (
	manager *StickerManager
	once    sync.Once
)

func InitStickerManager() error {
	var initErr error
	once.Do(func() {
		// Use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			initErr = fmt.Errorf("Failed to get current directory: %v", err)
			return
		}

		// Ensure resources directory exists
		resourcesDir := filepath.Join(currentDir, "resources")
		if err := os.MkdirAll(resourcesDir, 0755); err != nil {
			initErr = fmt.Errorf("Failed to create resources directory: %v", err)
			return
		}

		// Set stickers.json path
		filePath := filepath.Join(currentDir, "stickers.json")
		manager = &StickerManager{
			stickers: make(map[string]Sticker),
			filePath: filePath,
		}

		// Try to load existing stickers
		if err := manager.loadStickers(); err != nil {
			log.Printf("Warning: Failed to load stickers: %v", err)
		}
	})
	return initErr
}

func GetStickerManager() *StickerManager {
	return manager
}

func (sm *StickerManager) loadStickers() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(sm.filePath); os.IsNotExist(err) {
		log.Printf("Sticker file does not exist, will create new file: %s", sm.filePath)
		return nil
	}

	// Read file
	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return fmt.Errorf("Failed to read sticker file: %v", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, &sm.stickers); err != nil {
		return fmt.Errorf("Failed to parse sticker data: %v", err)
	}

	log.Printf("Successfully loaded %d stickers", len(sm.stickers))
	return nil
}

func (sm *StickerManager) saveStickers() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Serialize data to JSON
	data, err := json.MarshalIndent(sm.stickers, "", "  ")
	if err != nil {
		return fmt.Errorf("Failed to serialize sticker data: %v", err)
	}

	// Write to file
	if err := os.WriteFile(sm.filePath, data, 0644); err != nil {
		return fmt.Errorf("Failed to save sticker file: %v", err)
	}

	log.Printf("Successfully saved %d stickers", len(sm.stickers))
	return nil
}

func (sm *StickerManager) AddSticker(sticker Sticker) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if sticker exists
	if _, exists := sm.stickers[sticker.Name]; exists {
		return fmt.Errorf("Sticker '%s' already exists", sticker.Name)
	}

	// Save sticker
	sm.stickers[sticker.Name] = sticker
	if err := sm.saveStickers(); err != nil {
		return err
	}

	log.Printf("Successfully added sticker: %s", sticker.Name)
	return nil
}

func (sm *StickerManager) GetSticker(name string) (Sticker, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sticker, exists := sm.stickers[name]
	if !exists {
		return Sticker{}, fmt.Errorf("Sticker '%s' not found", name)
	}

	return sticker, nil
}

func (sm *StickerManager) ListStickers() []Sticker {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stickers := make([]Sticker, 0, len(sm.stickers))
	for _, sticker := range sm.stickers {
		stickers = append(stickers, sticker)
	}

	return stickers
}

func (sm *StickerManager) DeleteSticker(name string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.stickers[name]; !exists {
		return fmt.Errorf("Sticker '%s' not found", name)
	}

	delete(sm.stickers, name)
	if err := sm.saveStickers(); err != nil {
		return err
	}

	log.Printf("Successfully deleted sticker: %s", name)
	return nil
}

// Handle uploaded image
func (sm *StickerManager) HandleUploadedImage(name string, attachment *discordgo.MessageAttachment, addedBy string) error {
	// Download image
	resp, err := http.Get(attachment.URL)
	if err != nil {
		return fmt.Errorf("Failed to download image: %v", err)
	}
	defer resp.Body.Close()

	// Ensure filename is safe
	safeName := filepath.Base(name)
	ext := filepath.Ext(attachment.Filename)
	if ext == "" {
		ext = ".png" // Default extension
	}

	// Create local file path
	localPath := filepath.Join("resources", safeName+ext)

	// Create file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Copy image data
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("Failed to save image: %v", err)
	}

	// Add sticker record
	sticker := Sticker{
		Name:      name,
		URL:       attachment.URL,
		LocalPath: localPath,
		AddedBy:   addedBy,
		AddedAt:   time.Now().Format(time.RFC3339),
	}

	sm.mutex.Lock()
	sm.stickers[name] = sticker
	sm.mutex.Unlock()

	// Save updates
	if err := sm.saveStickers(); err != nil {
		return err
	}

	log.Printf("Successfully saved uploaded image: %s", name)
	return nil
} 