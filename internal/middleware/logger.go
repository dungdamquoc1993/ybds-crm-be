package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger creates a middleware that logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request and response details
		method := c.Method()
		path := c.Path()
		status := c.Response().StatusCode()
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		// Format log message
		logMessage := fmt.Sprintf("[%s] %s %s %d %s %s %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			status,
			latency,
			ip,
			userAgent,
		)

		// Log based on status code
		if status >= 500 {
			fmt.Printf("\x1b[31m%s\x1b[0m\n", logMessage) // Red for server errors
		} else if status >= 400 {
			fmt.Printf("\x1b[33m%s\x1b[0m\n", logMessage) // Yellow for client errors
		} else if status >= 300 {
			fmt.Printf("\x1b[36m%s\x1b[0m\n", logMessage) // Cyan for redirects
		} else {
			fmt.Printf("\x1b[32m%s\x1b[0m\n", logMessage) // Green for success
		}

		return err
	}
}
