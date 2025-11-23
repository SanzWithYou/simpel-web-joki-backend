package middleware

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// Rate limit config
type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

// Get config from env
func GetRateLimitConfig(endpoint string, defaultMax int, defaultExpiration time.Duration) RateLimitConfig {
	maxStr := os.Getenv(fmt.Sprintf("RATE_LIMIT_%s_MAX", endpoint))
	expirationStr := os.Getenv(fmt.Sprintf("RATE_LIMIT_%s_EXPIRATION", endpoint))

	max := defaultMax
	if maxStr != "" {
		if val, err := strconv.Atoi(maxStr); err == nil {
			max = val
		}
	}

	expiration := defaultExpiration
	if expirationStr != "" {
		if val, err := strconv.Atoi(expirationStr); err == nil {
			expiration = time.Duration(val) * time.Second
		}
	}

	return RateLimitConfig{
		Max:        max,
		Expiration: expiration,
	}
}

// Rate limit middleware
func RateLimit(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Expiration).Unix()))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak permintaan, silakan coba lagi nanti",
			})
		},
		Next: func(c *fiber.Ctx) bool {
			remaining := config.Max - 1

			if val := c.Locals("limiter_remaining"); val != nil {
				if r, ok := val.(int); ok {
					remaining = r
				}
			}

			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
			c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(config.Expiration).Unix()))

			return false
		},
	})
}
