# Reusable Packages

This directory contains reusable packages that can be used across the application.

## Upload Package

The upload package provides functionality for handling file uploads, including:

- Configuration for upload settings (allowed file types, max size, etc.)
- File upload service with validation
- HTTP handlers for upload endpoints
- Static file serving middleware
- Utility functions for generating file URLs

### Usage

```go
// Create a new upload configuration
config := upload.NewConfig("./uploads")
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

// Register the upload routes with a Fiber router
app := fiber.New()
uploadHandler.RegisterRoutes(app.Group("/api/uploads"))

// Register the static middleware to serve uploaded files
upload.RegisterStaticRoutes(app, "./uploads")

// Get a file URL
fileURL := upload.GetFileURL("example.jpg", "products")
```

## WebSocket Package

The WebSocket package provides functionality for real-time communication, including:

- WebSocket hub for managing connections and broadcasting messages
- Client management with authentication and authorization
- Topic-based subscriptions
- Message broadcasting to specific topics, users, or roles
- Middleware for authenticating WebSocket connections
- Support for JWT, query parameter, and header-based authentication

### Usage

```go
// Create a new hub
hub := websocket.NewHub()
hub = hub.WithCleanupInterval(5 * time.Minute)
hub = hub.WithInactiveTimeout(15 * time.Minute)

// Configure topic authorization
hub = hub.WithTopicAuth(func(client *websocket.Client, topic string) bool {
    // Only allow subscriptions to topics that match the user's roles
    for _, role := range client.Roles {
        if strings.HasPrefix(topic, role+".") {
            return true
        }
    }
    return false
})

// Configure message handling
hub = hub.WithMessageHandler(func(client *websocket.Client, message []byte) error {
    // Process the message
    log.Printf("Received message from %s: %s", client.UserID, string(message))
    return nil
})

// Start the hub
go hub.Run()

// Create a JWT validator function
validateToken := func(tokenString string) (string, []string, error) {
    // Validate the token and return user ID and roles
    return "user123", []string{"user"}, nil
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

// Register the WebSocket routes
app := fiber.New()
handler.RegisterRoutes(app, "/ws")

// Broadcast a message to a topic
hub.Broadcast <- &websocket.BroadcastMessage{
    Topic:   "notifications",
    Message: []byte(`{"type":"notification","message":"Hello, world!"}`),
}
```

## Integration Example

See the `examples` package for a complete example of how to integrate the upload and WebSocket packages. 