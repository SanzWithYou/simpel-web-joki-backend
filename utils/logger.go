package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Ensure log dir
func ensureLogDir() error {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return os.Mkdir(logDir, 0755)
	}
	return nil
}

// Log request
func LogRequest(c *fiber.Ctx, duration time.Duration, err error) {
	// Create dir
	if err := ensureLogDir(); err != nil {
		log.Println("Error creating log directory:", err)
		return
	}

	// Format log
	logEntry := fmt.Sprintf(
		"[%s] %s %s - %d - %v",
		time.Now().Format("2006-01-02 15:04:05"),
		c.Method(),
		c.Path(),
		c.Response().StatusCode(),
		duration,
	)

	if err != nil {
		logEntry += fmt.Sprintf(" - Error: %v", err)
	}

	// Write to file
	logFile := filepath.Join("logs", "app.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Println(logEntry)
}

// Log error
func LogError(err error) {
	// Create dir
	if dirErr := ensureLogDir(); dirErr != nil {
		log.Println("Error creating log directory:", dirErr)
		return
	}

	// Format log
	logEntry := fmt.Sprintf(
		"[%s] ERROR: %v",
		time.Now().Format("2006-01-02 15:04:05"),
		err,
	)

	// Write to file
	errorLogFile := filepath.Join("logs", "error.log")
	f, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening error log file:", err)
		return
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Println(logEntry)
}
