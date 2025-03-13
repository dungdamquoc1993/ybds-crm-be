package websocket

import (
	"errors"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// AuthFunc is a function that authenticates a WebSocket connection
type AuthFunc func(c *fiber.Ctx) (userID string, roles []string, err error)

// Handler handles WebSocket connections
type Handler struct {
	hub      *Hub
	authFunc AuthFunc
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authFunc AuthFunc) *Handler {
	return &Handler{
		hub:      hub,
		authFunc: authFunc,
	}
}

// WithDefaultAuth sets a default authentication function that allows anonymous access
func (h *Handler) WithDefaultAuth() *Handler {
	h.authFunc = func(c *fiber.Ctx) (string, []string, error) {
		return "anonymous", []string{"guest"}, nil
	}
	return h
}

// HandleConnection handles a WebSocket connection
func (h *Handler) HandleConnection(c *websocket.Conn) {
	// Get user ID and roles from the context
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		log.Println("user_id not found in context")
		return
	}

	rolesInterface, ok := c.Locals("roles").([]string)
	if !ok {
		log.Println("roles not found in context")
		rolesInterface = []string{"guest"}
	}

	// Create a new client
	client := NewClient(c, h.hub, userID, rolesInterface)

	// Register the client
	h.hub.Register <- client

	// Start the client's read and write pumps
	go client.WritePump()
	client.ReadPump()
}

// Middleware creates a middleware that authenticates WebSocket connections
func (h *Handler) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Upgrade middleware
		if websocket.IsWebSocketUpgrade(c) {
			// Authenticate the connection
			userID, roles, err := h.authFunc(c)
			if err != nil {
				return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
			}

			// Store user ID and roles in the context
			c.Locals("user_id", userID)
			c.Locals("roles", roles)

			// Allow the upgrade
			return c.Next()
		}

		return fiber.NewError(fiber.StatusUpgradeRequired, "Upgrade required")
	}
}

// RegisterRoutes registers WebSocket routes
func (h *Handler) RegisterRoutes(app *fiber.App, path string) {
	// Register the WebSocket route
	app.Use(path, h.Middleware())
	app.Get(path, websocket.New(h.HandleConnection))
}

// JWTAuthFunc creates an authentication function that uses JWT tokens
func JWTAuthFunc(getTokenFromRequest func(*fiber.Ctx) string, validateToken func(string) (string, []string, error)) AuthFunc {
	return func(c *fiber.Ctx) (string, []string, error) {
		// Get the token from the request
		token := getTokenFromRequest(c)
		if token == "" {
			return "", nil, errors.New("no token provided")
		}

		// Validate the token
		userID, roles, err := validateToken(token)
		if err != nil {
			return "", nil, err
		}

		return userID, roles, nil
	}
}

// QueryAuthFunc creates an authentication function that uses query parameters
func QueryAuthFunc(validateToken func(string) (string, []string, error)) AuthFunc {
	return func(c *fiber.Ctx) (string, []string, error) {
		// Get the token from the query
		token := c.Query("token")
		if token == "" {
			return "", nil, errors.New("no token provided")
		}

		// Validate the token
		userID, roles, err := validateToken(token)
		if err != nil {
			return "", nil, err
		}

		return userID, roles, nil
	}
}

// HeaderAuthFunc creates an authentication function that uses headers
func HeaderAuthFunc(header string, validateToken func(string) (string, []string, error)) AuthFunc {
	return func(c *fiber.Ctx) (string, []string, error) {
		// Get the token from the header
		token := c.Get(header)
		if token == "" {
			return "", nil, errors.New("no token provided")
		}

		// Validate the token
		userID, roles, err := validateToken(token)
		if err != nil {
			return "", nil, err
		}

		return userID, roles, nil
	}
}
