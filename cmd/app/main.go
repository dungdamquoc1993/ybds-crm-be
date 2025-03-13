package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	fiberwsocket "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	_ "github.com/ybds/docs" // Import swagger docs
	"github.com/ybds/internal/api/handlers"
	"github.com/ybds/internal/database"
	"github.com/ybds/internal/services"
	"github.com/ybds/pkg/config"
	pkgdb "github.com/ybds/pkg/database"
	pkgjwt "github.com/ybds/pkg/jwt"
	pkgws "github.com/ybds/pkg/websocket"
)

// @title YBDS API
// @version 1.0
// @description API Server for YBDS Application
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email your-email@domain.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api
// @schemes http https
func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := pkgdb.NewDatabase(&cfg.AccountDB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize database for internal use
	if err := database.InitDatabase(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize JWT service
	jwtService, err := pkgjwt.NewJWTService(&cfg.JWT)
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	// Initialize websocket hub
	hub := pkgws.NewHub()
	go hub.Run()

	// Initialize services in the correct order to respect dependencies
	notificationService := services.NewNotificationService(db, hub)
	userService := services.NewUserService(db, notificationService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtService)
	userHandler := handlers.NewUserHandler(db, notificationService)
	productHandler := handlers.NewProductHandler(db, notificationService)
	orderHandler := handlers.NewOrderHandler(db, userService, notificationService)
	notificationHandler := handlers.NewNotificationHandler(db, hub)

	// Initialize JWT service wrapper for internal use
	internalJWTService := services.NewJWTServiceWrapper(jwtService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "YBDS API",
		ErrorHandler: customErrorHandler,
	})

	// Register middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New())

	// Setup Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "YBDS API Documentation",
	}))

	// Create API routes
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Register auth routes
	auth := api.Group("/auth")
	authHandler.RegisterRoutes(auth)

	// Register websocket route
	wsHandler := pkgws.NewHandler(hub, func(c *fiber.Ctx) (string, []string, error) {
		userID, ok := c.Locals("user_id").(string)
		if !ok {
			return "", nil, fmt.Errorf("user_id not found in context")
		}

		roles, ok := c.Locals("roles").([]string)
		if !ok {
			roles = []string{"user"}
		}

		return userID, roles, nil
	})

	api.Use("/ws", wsHandler.Middleware())
	api.Get("/ws", fiberwsocket.New(wsHandler.HandleConnection))

	// Create authenticated routes
	authenticated := api.Group("/")
	authenticated.Use(internalJWTService.AuthMiddleware)

	// Register user routes
	userHandler.RegisterRoutes(authenticated, internalJWTService.AuthMiddleware)

	// Register product routes
	productHandler.RegisterRoutes(authenticated, internalJWTService.AuthMiddleware)

	// Register order routes
	orderHandler.RegisterRoutes(authenticated, internalJWTService.AuthMiddleware)

	// Register notification routes
	notificationHandler.RegisterRoutes(authenticated, internalJWTService.AuthMiddleware)

	// Start server
	serverPort := fmt.Sprintf(":%s", cfg.Server.Port)
	go func() {
		if err := app.Listen(serverPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/", cfg.Server.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server gracefully stopped")
}

// customErrorHandler handles errors returned from routes
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Set Content-Type: application/json
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Return status code with error message
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": "Error occurred",
		"error":   err.Error(),
	})
}
