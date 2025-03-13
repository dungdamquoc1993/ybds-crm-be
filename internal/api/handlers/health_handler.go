package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ybds/internal/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HandleHealthCheck handles the health check endpoint
func (h *HealthHandler) HandleHealthCheck(c *fiber.Ctx) error {
	return utils.SuccessResponse(c, fiber.StatusOK, "Service is healthy", map[string]string{
		"status":  "up",
		"version": "1.0.0",
	})
}

// RegisterRoutes registers the health check routes
func (h *HealthHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/health", h.HandleHealthCheck)
}
