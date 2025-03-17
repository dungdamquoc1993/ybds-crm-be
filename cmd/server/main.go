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
	"github.com/ybds/internal/middleware"
	"github.com/ybds/internal/services"
	"github.com/ybds/pkg/config"
	pkgdb "github.com/ybds/pkg/database"
	pkgjwt "github.com/ybds/pkg/jwt"
	pkgupload "github.com/ybds/pkg/upload"
	pkgws "github.com/ybds/pkg/websocket"
)

// @title YBDS API
// @version 1.0
// @description YBDS API Documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @externalDocs.description Find out more about Swagger
// @externalDocs.url https://swagger.io/docs/

// @tag.name products
// @tag.description Product management endpoints

// @tag.name product-images
// @tag.description Product image management endpoints. Images are stored at /uploads/products/ and can be accessed via this path.

// @x-upload-info {"base-path":"/uploads","products-path":"/uploads/products","max-size":"10MB","allowed-types":["image/jpeg","image/png","image/gif","image/webp"]}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Debug: Print database configuration
	log.Printf("DB_HOST: %s", cfg.AccountDB.Host)
	log.Printf("DB_PORT: %s", cfg.AccountDB.Port)
	log.Printf("DB_USER: %s", cfg.AccountDB.User)
	log.Printf("DB_ACCOUNT_NAME: %s", cfg.AccountDB.Name)
	log.Printf("DB_NOTIFICATION_NAME: %s", cfg.NotificationDB.Name)
	log.Printf("DB_ORDER_NAME: %s", cfg.OrderDB.Name)
	log.Printf("DB_PRODUCT_NAME: %s", cfg.ProductDB.Name)
	log.Printf("DB_SSL_MODE: %s", cfg.AccountDB.SSLMode)

	// Initialize multiple database connections
	dbConnections, err := pkgdb.NewDatabaseConnections(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}

	// Initialize databases for internal use
	if err := database.InitDatabases(dbConnections); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	// Initialize JWT service
	jwtService, err := pkgjwt.NewJWTService(&cfg.JWT)
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	// Initialize websocket hub
	hub := pkgws.NewHub()
	go hub.Run()

	// Initialize upload service
	uploadConfig := pkgupload.NewConfig("./uploads")
	uploadConfig.WithSubDir("products")
	uploadService, err := pkgupload.NewService(uploadConfig)
	if err != nil {
		log.Fatalf("Failed to initialize upload service: %v", err)
	}

	// Initialize services in the correct order to respect dependencies
	notificationService := services.NewNotificationService(dbConnections.NotificationDB, hub)
	userService := services.NewUserService(dbConnections.AccountDB, notificationService)
	productService := services.NewProductService(dbConnections.ProductDB, notificationService, uploadService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(dbConnections.AccountDB, jwtService, userService)
	userHandler := handlers.NewUserHandler(dbConnections.AccountDB, notificationService)
	productHandler := handlers.NewProductHandler(dbConnections.ProductDB, notificationService, uploadService)
	orderHandler := handlers.NewOrderHandler(dbConnections.OrderDB, productService, userService, notificationService)
	notificationHandler := handlers.NewNotificationHandler(dbConnections.NotificationDB, notificationService, hub)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "YBDS API",
		ErrorHandler: customErrorHandler,
	})

	// Register static routes for serving uploaded files
	pkgupload.RegisterStaticRoutes(app, uploadConfig.BaseDir)

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

	// Public routes that don't require authentication
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/register", authHandler.Register)

	// Register websocket route with its own middleware
	wsHandler := pkgws.NewHandler(hub, pkgws.JWTAuthFunc(
		// Function to extract token from request
		func(c *fiber.Ctx) string {
			token := c.Query("token")
			if token != "" {
				fmt.Printf("[WebSocket] Token from query: %s\n", token[:10]+"...")
			} else {
				fmt.Printf("[WebSocket] No token provided in query\n")
			}
			return token
		},
		// Function to validate token
		func(tokenString string) (string, []string, error) {
			// Use the same JWT validation logic as regular API endpoints
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				fmt.Printf("[WebSocket] Token validation error: %v\n", err)
				return "", nil, err
			}

			fmt.Printf("[WebSocket] Token validated successfully for user %s with roles %v\n", claims.UserID, claims.Roles)
			return claims.UserID, claims.Roles, nil
		},
	))

	wsGroup := api.Group("/ws")
	wsGroup.Use(wsHandler.Middleware())
	wsGroup.Get("/", fiberwsocket.New(wsHandler.HandleConnection))

	// Protected routes that require authentication
	// Create authenticated routes group
	authenticated := api.Group("/")
	authenticated.Use(middleware.JWTAuth(jwtService))

	// Create admin-only routes
	adminRoutes := authenticated.Group("/")
	adminRoutes.Use(middleware.AdminGuard())

	// Create routes for both admin and agent
	adminOrAgentRoutes := authenticated.Group("/")
	adminOrAgentRoutes.Use(middleware.AdminOrAgentGuard())

	// Register user routes - Admin only
	adminRoutes.Get("/users", userHandler.GetUsers)
	adminRoutes.Get("/users/:id", userHandler.GetUserByID)

	// Register product routes using the RegisterRoutes method
	productHandler.RegisterRoutes(adminOrAgentRoutes, middleware.JWTAuth(jwtService))

	// Register order routes using the RegisterRoutes method
	orderHandler.RegisterRoutes(adminOrAgentRoutes, middleware.JWTAuth(jwtService))

	// Admin-only order routes
	adminRoutes.Put("/orders/:id/status", orderHandler.UpdateOrderStatus)
	adminRoutes.Delete("/orders/:id", orderHandler.DeleteOrder)

	// Register notification routes - Admin or Agent
	adminOrAgentRoutes.Get("/notifications", notificationHandler.GetNotifications)
	adminOrAgentRoutes.Get("/notifications/unread", notificationHandler.GetUnreadNotifications)
	adminOrAgentRoutes.Put("/notifications/:id/read", notificationHandler.MarkAsRead)
	adminOrAgentRoutes.Put("/notifications/read-all", notificationHandler.MarkAllAsRead)

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
