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

// HandleHealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is up and running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
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
