package upload

import (
	"github.com/gofiber/fiber/v2"
)

// Handler provides HTTP handlers for file uploads
type Handler struct {
	service *Service
}

// NewHandler creates a new upload handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// UploadResponse represents the response for a file upload
type UploadResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    *UploadResult `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// DeleteResponse represents the response for a file deletion
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// HandleUpload handles file upload requests
func (h *Handler) HandleUpload(c *fiber.Ctx) error {
	// Get the prefix from query parameters (optional)
	prefix := c.Query("prefix", "")

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: "Failed to get file from request",
			Error:   err.Error(),
		})
	}

	// Upload the file
	result, err := h.service.Upload(file, prefix)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(UploadResponse{
			Success: false,
			Message: "Failed to upload file",
			Error:   err.Error(),
		})
	}

	// Return the result
	return c.Status(fiber.StatusOK).JSON(UploadResponse{
		Success: true,
		Message: "File uploaded successfully",
		Data:    result,
	})
}

// HandleDelete handles file deletion requests
func (h *Handler) HandleDelete(c *fiber.Ctx) error {
	// Get the filename from the request
	filename := c.Params("filename")
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(DeleteResponse{
			Success: false,
			Message: "Filename is required",
			Error:   "Missing filename parameter",
		})
	}

	// Delete the file
	err := h.service.Delete(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(DeleteResponse{
			Success: false,
			Message: "Failed to delete file",
			Error:   err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(DeleteResponse{
		Success: true,
		Message: "File deleted successfully",
	})
}

// RegisterRoutes registers the upload routes
func (h *Handler) RegisterRoutes(router fiber.Router) {
	uploadGroup := router.Group("/upload")

	// Upload endpoint
	uploadGroup.Post("/", h.HandleUpload)

	// Delete endpoint
	uploadGroup.Delete("/:filename", h.HandleDelete)
}
