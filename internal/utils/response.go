package utils

import (
	"github.com/gofiber/fiber/v2"
)

// Response is the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse returns a success response
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, err interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

// ValidationErrorResponse returns a validation error response
func ValidationErrorResponse(c *fiber.Ctx, err error) error {
	return ErrorResponse(c, fiber.StatusBadRequest, "Validation error", err.Error())
}

// ServerErrorResponse returns a server error response
func ServerErrorResponse(c *fiber.Ctx, err error) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, "Internal server error", err.Error())
}

// NotFoundResponse returns a not found error response
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, fiber.StatusNotFound, message, nil)
}

// UnauthorizedResponse returns an unauthorized error response
func UnauthorizedResponse(c *fiber.Ctx) error {
	return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized", nil)
}

// ForbiddenResponse returns a forbidden error response
func ForbiddenResponse(c *fiber.Ctx) error {
	return ErrorResponse(c, fiber.StatusForbidden, "Forbidden", nil)
}
