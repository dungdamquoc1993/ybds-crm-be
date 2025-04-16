package upload

import (
	"fmt"
	"path/filepath"
)

// StorageType defines the type of storage to use
type StorageType string

const (
	// StorageTypeLocal indicates local file storage
	StorageTypeLocal StorageType = "local"
	// StorageTypeS3 indicates AWS S3 storage
	StorageTypeS3 StorageType = "s3"
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

	// StorageType determines where to store files (local or s3)
	StorageType StorageType

	// S3Config contains AWS S3 configuration
	S3Config *S3Config
}

// S3Config contains configuration for AWS S3
type S3Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
	Prefix    string
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
		MaxSize:     10, // 10MB default
		StorageType: StorageTypeLocal,
	}
}

// WithS3 configures the upload service to use AWS S3
func (c *Config) WithS3(accessKey, secretKey, region, bucket string, prefix string) *Config {
	c.StorageType = StorageTypeS3
	c.S3Config = &S3Config{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Bucket:    bucket,
		Prefix:    prefix,
	}
	return c
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
	if c.StorageType == StorageTypeLocal && c.BaseDir == "" {
		return fmt.Errorf("base directory is required for local storage")
	}
	if c.StorageType == StorageTypeS3 && c.S3Config == nil {
		return fmt.Errorf("S3 configuration is required for S3 storage")
	}
	if c.StorageType == StorageTypeS3 && (c.S3Config.AccessKey == "" || c.S3Config.SecretKey == "" || c.S3Config.Bucket == "" || c.S3Config.Region == "") {
		return fmt.Errorf("incomplete S3 configuration")
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("max size must be greater than 0")
	}
	if len(c.AllowedTypes) == 0 {
		return fmt.Errorf("at least one allowed type is required")
	}
	return nil
}
