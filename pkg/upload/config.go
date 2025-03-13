package upload

import (
	"fmt"
	"path/filepath"
)

// Config defines the configuration for file uploads
type Config struct {
	// BaseDir is the base directory for file uploads
	BaseDir string

	// AllowedTypes is a map of allowed MIME types
	AllowedTypes map[string]bool

	// MaxSize is the maximum file size in megabytes
	MaxSize int64

	// SubDir is an optional subdirectory within BaseDir
	SubDir string
}

// NewConfig creates a new upload configuration with default values
func NewConfig(baseDir string) *Config {
	return &Config{
		BaseDir: baseDir,
		AllowedTypes: map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		},
		MaxSize: 10, // 10MB default
	}
}

// WithSubDir sets the subdirectory for uploads
func (c *Config) WithSubDir(subDir string) *Config {
	c.SubDir = subDir
	return c
}

// WithMaxSize sets the maximum file size in megabytes
func (c *Config) WithMaxSize(maxSize int64) *Config {
	c.MaxSize = maxSize
	return c
}

// WithAllowedTypes sets the allowed MIME types
func (c *Config) WithAllowedTypes(types []string) *Config {
	c.AllowedTypes = make(map[string]bool)
	for _, t := range types {
		c.AllowedTypes[t] = true
	}
	return c
}

// AddAllowedType adds a MIME type to the allowed types
func (c *Config) AddAllowedType(mimeType string) *Config {
	if c.AllowedTypes == nil {
		c.AllowedTypes = make(map[string]bool)
	}
	c.AllowedTypes[mimeType] = true
	return c
}

// GetUploadDir returns the full upload directory path
func (c *Config) GetUploadDir() string {
	if c.SubDir != "" {
		return filepath.Join(c.BaseDir, c.SubDir)
	}
	return c.BaseDir
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BaseDir == "" {
		return fmt.Errorf("base directory is required")
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("max size must be greater than 0")
	}
	if len(c.AllowedTypes) == 0 {
		return fmt.Errorf("at least one allowed type is required")
	}
	return nil
}
