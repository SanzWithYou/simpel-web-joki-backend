package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Log request
func LogRequest(c *fiber.Ctx, duration time.Duration, err error) {
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

	// Cetak langsung ke console
	log.Println(logEntry)
}

// Log error
func LogError(err error) {
	// Format log
	logEntry := fmt.Sprintf(
		"[%s] ERROR: %v",
		time.Now().Format("2006-01-02 15:04:05"),
		err,
	)

	// Cetak langsung ke console
	log.Println(logEntry)
}
