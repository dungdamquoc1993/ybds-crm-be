package upload

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// StaticMiddleware creates a middleware to serve uploaded files
func StaticMiddleware(baseDir string) fiber.Handler {
	// Ensure the directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			panic(err)
		}
	}

	// Create a filesystem that serves files from the upload directory
	fs := http.Dir(baseDir)

	// Configure the filesystem middleware
	config := filesystem.Config{
		Root:         fs,
		Browse:       false,
		Index:        "index.html",
		MaxAge:       3600,
		NotFoundFile: "",
	}

	return filesystem.New(config)
}

// RegisterStaticRoutes registers routes to serve uploaded files
func RegisterStaticRoutes(app *fiber.App, baseDir string) {
	// Register the static middleware at /uploads
	app.Use("/uploads", StaticMiddleware(baseDir))
}

// GetFileURL returns the URL for a file
func GetFileURL(filename, subDir string) string {
	if subDir != "" {
		return filepath.Join("/uploads", subDir, filename)
	}
	return filepath.Join("/uploads", filename)
}
