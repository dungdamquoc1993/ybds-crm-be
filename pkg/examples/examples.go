package examples

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ybds/pkg/upload"
	"github.com/ybds/pkg/websocket"
)

// ExampleUploadConfig demonstrates how to configure and use the upload package
func ExampleUploadConfig() {
	// Create a new upload configuration
	config := upload.NewConfig("./uploads")

	// Configure the upload settings
	config = config.WithMaxSize(5 * 1024 * 1024) // 5MB
	config = config.WithAllowedTypes([]string{"image/jpeg", "image/png"})
	config = config.WithSubDir("products")

	// Create a new upload service
	uploadService, err := upload.NewService(config)
	if err != nil {
		log.Fatalf("Failed to create upload service: %v", err)
	}

	// Create a new upload handler
	uploadHandler := upload.NewHandler(uploadService)

	// Create a new Fiber app
	app := fiber.New()

	// Register the upload routes
	uploadHandler.RegisterRoutes(app.Group("/api/uploads"))

	// Register the static middleware to serve uploaded files
	upload.RegisterStaticRoutes(app, "./uploads")

	// Example of how to get a file URL
	fileURL := upload.GetFileURL("example.jpg", "products")
	fmt.Println("File URL:", fileURL)
}

// ExampleWebSocketConfig demonstrates how to configure and use the WebSocket package
func ExampleWebSocketConfig() {
	// Create a new hub
	hub := websocket.NewHub()

	// Configure the hub
	hub = hub.WithCleanupInterval(5 * time.Minute)
	hub = hub.WithInactiveTimeout(15 * time.Minute)
	hub = hub.WithTopicAuth(func(client *websocket.Client, topic string) bool {
		// Only allow subscriptions to topics that match the user's roles
		for _, role := range client.Roles {
			if strings.HasPrefix(topic, role+".") {
				return true
			}
		}
		return false
	})
	hub = hub.WithMessageHandler(func(client *websocket.Client, message []byte) error {
		// Log the message
		log.Printf("Received message from %s: %s", client.UserID, string(message))
		return nil
	})

	// Start the hub
	go hub.Run()

	// Create a JWT validator function
	validateToken := func(tokenString string) (string, []string, error) {
		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the secret key
			return []byte("your-secret-key"), nil
		})

		if err != nil {
			return "", nil, err
		}

		// Get the claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Get the user ID
			userID, ok := claims["user_id"].(string)
			if !ok {
				return "", nil, fmt.Errorf("user_id not found in token")
			}

			// Get the roles
			rolesInterface, ok := claims["roles"].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("roles not found in token")
			}

			// Convert roles to strings
			roles := make([]string, len(rolesInterface))
			for i, role := range rolesInterface {
				roles[i] = role.(string)
			}

			return userID, roles, nil
		}

		return "", nil, fmt.Errorf("invalid token")
	}

	// Create a new WebSocket handler with JWT authentication
	handler := websocket.NewHandler(hub, websocket.JWTAuthFunc(
		func(c *fiber.Ctx) string {
			// Get the token from the Authorization header
			auth := c.Get("Authorization")
			if auth == "" {
				return ""
			}

			// Remove the "Bearer " prefix
			if strings.HasPrefix(auth, "Bearer ") {
				return auth[7:]
			}

			return auth
		},
		validateToken,
	))

	// Create a new Fiber app
	app := fiber.New()

	// Register the WebSocket routes
	handler.RegisterRoutes(app, "/ws")
}

// ExampleIntegration demonstrates how to integrate the upload and WebSocket packages
func ExampleIntegration() {
	// Create a new Fiber app
	app := fiber.New()

	// Create a new upload configuration
	uploadConfig := upload.NewConfig("./uploads")
	uploadConfig = uploadConfig.WithMaxSize(10 * 1024 * 1024)
	uploadConfig = uploadConfig.WithAllowedTypes([]string{"image/jpeg", "image/png", "image/gif", "image/webp"})

	// Create a new upload service
	uploadService, err := upload.NewService(uploadConfig)
	if err != nil {
		log.Fatalf("Failed to create upload service: %v", err)
	}

	// Create a new upload handler
	uploadHandler := upload.NewHandler(uploadService)

	// Register the upload routes
	uploadHandler.RegisterRoutes(app.Group("/api/uploads"))

	// Register the static middleware to serve uploaded files
	upload.RegisterStaticRoutes(app, "./uploads")

	// Create a new WebSocket hub
	hub := websocket.NewHub()

	// Configure the hub
	hub = hub.WithCleanupInterval(5 * time.Minute)
	hub = hub.WithInactiveTimeout(15 * time.Minute)

	// Start the hub
	go hub.Run()

	// Create a JWT validator function
	validateToken := func(tokenString string) (string, []string, error) {
		// Implementation omitted for brevity
		return "user123", []string{"user"}, nil
	}

	// Create a new WebSocket handler with JWT authentication
	wsHandler := websocket.NewHandler(hub, websocket.JWTAuthFunc(
		func(c *fiber.Ctx) string {
			// Get the token from the Authorization header
			auth := c.Get("Authorization")
			if auth == "" {
				return ""
			}

			// Remove the "Bearer " prefix
			if strings.HasPrefix(auth, "Bearer ") {
				return auth[7:]
			}

			return auth
		},
		validateToken,
	))

	// Register the WebSocket routes
	wsHandler.RegisterRoutes(app, "/ws")

	// Example of notifying clients when a new file is uploaded
	app.Post("/api/products", func(c *fiber.Ctx) error {
		// Parse the request
		var product struct {
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			ImageURL    string  `json:"image_url"`
		}

		if err := c.BodyParser(&product); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request",
				"error":   err.Error(),
			})
		}

		// Save the product to the database (implementation omitted)

		// Notify all clients subscribed to the "products" topic
		notification := struct {
			Type    string `json:"type"`
			Topic   string `json:"topic"`
			Payload struct {
				Action  string `json:"action"`
				Product struct {
					Name        string  `json:"name"`
					Description string  `json:"description"`
					Price       float64 `json:"price"`
					ImageURL    string  `json:"image_url"`
				} `json:"product"`
			} `json:"payload"`
		}{
			Type:  "notification",
			Topic: "products",
			Payload: struct {
				Action  string `json:"action"`
				Product struct {
					Name        string  `json:"name"`
					Description string  `json:"description"`
					Price       float64 `json:"price"`
					ImageURL    string  `json:"image_url"`
				} `json:"product"`
			}{
				Action: "created",
				Product: struct {
					Name        string  `json:"name"`
					Description string  `json:"description"`
					Price       float64 `json:"price"`
					ImageURL    string  `json:"image_url"`
				}{
					Name:        product.Name,
					Description: product.Description,
					Price:       product.Price,
					ImageURL:    product.ImageURL,
				},
			},
		}

		// Convert the notification to JSON
		data, err := json.Marshal(notification)
		if err != nil {
			log.Printf("Failed to marshal notification: %v", err)
		} else {
			// Broadcast the notification to all subscribers of the "products" topic
			hub.Broadcast <- &websocket.BroadcastMessage{
				Topic:   "products",
				Message: data,
			}
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Product created successfully",
			"data":    product,
		})
	})

	// Start the server
	log.Fatal(app.Listen(":3000"))
}
