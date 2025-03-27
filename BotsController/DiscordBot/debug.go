package DiscordBot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Token        string `json:"token"`
	ResourcesDir string `json:"resources_dir"`
	StickersFile string `json:"stickers_file"`
	LogsDir      string `json:"logs_dir"`
	CommandPrefix string `json:"command_prefix"`
}

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
	logFile     *os.File
	config      Config
)

// LoadConfig loads configuration from the specified file
func LoadConfig(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return err
	}

	fmt.Printf("Config loaded from %s\n", file)
	return nil
}

// InitLoggers initializes the loggers for debugging
func InitLoggers() error {
	// Use logs directory from config
	logsDir := "logs"
	if config.LogsDir != "" {
		logsDir = config.LogsDir
	}
	
	// Ensure logs directory exists
	fmt.Printf("Creating logs directory: %s\n", logsDir)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Error creating logs directory: %v\n", err)
		return err
	}

	// Create log file with current timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logsDir, fmt.Sprintf("bot_%s.log", timestamp))
	
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return err
	}
	
	fmt.Printf("Logs will be written to: %s\n", logFilePath)

	// Initialize loggers
	debugLogger = log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Output to console as well
	debugLogger.SetOutput(os.Stdout)
	infoLogger.SetOutput(os.Stdout)
	errorLogger.SetOutput(os.Stderr)

	fmt.Printf("Loggers initialized successfully\n")
	return nil
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fmt.Printf("DEBUG: %s\n", message)
	if debugLogger != nil {
		debugLogger.Output(2, message)
	}
}

// LogInfo logs an info message
func LogInfo(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fmt.Printf("INFO: %s\n", message)
	if infoLogger != nil {
		infoLogger.Output(2, message)
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	fmt.Printf("ERROR: %s\n", message)
	if errorLogger != nil {
		errorLogger.Output(2, message)
	}
}

// CloseLogFile closes the log file
func CloseLogFile() {
	if logFile != nil {
		logFile.Close()
		fmt.Println("Log file closed")
	}
} 