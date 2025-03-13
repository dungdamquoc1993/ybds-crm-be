package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Test API",
	})

	// Register middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New())

	// Create API routes
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/login", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Login successful",
			"token":   "test-token",
		})
	})

	auth.Post("/register", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Registration successful",
			"user": fiber.Map{
				"id": "test-user-id",
			},
		})
	})

	// Start server
	port := "3001"
	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
