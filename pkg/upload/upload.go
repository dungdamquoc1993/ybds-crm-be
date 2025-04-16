package upload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadResult represents the result of a file upload
type UploadResult struct {
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Path        string `json:"path"`
	URL         string `json:"url"`
}

// MultipleUploadResult represents the result of multiple file uploads
type MultipleUploadResult struct {
	Files []*UploadResult `json:"files"`
}

// Service handles file uploads
type Service struct {
	config   *Config
	s3Client *S3Client
}

// NewService creates a new upload service
func NewService(config *Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid upload configuration: %w", err)
	}

	service := &Service{
		config: config,
	}

	// Initialize S3 client if using S3 storage
	if config.StorageType == StorageTypeS3 {
		s3Client, err := NewS3Client(
			config.S3Config.AccessKey,
			config.S3Config.SecretKey,
			config.S3Config.Region,
			config.S3Config.Bucket,
			config.S3Config.Prefix,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
		}
		service.s3Client = s3Client
	} else if config.StorageType == StorageTypeLocal {
		// Ensure upload directory exists for local storage
		uploadDir := config.GetUploadDir()
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create upload directory: %w", err)
		}
	}

	return service, nil
}

// Upload handles a file upload from a multipart form
func (s *Service) Upload(file *multipart.FileHeader, subDir string) (*UploadResult, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Reset the file pointer
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)

	// Validate file type
	if !s.config.AllowedTypes[contentType] {
		return nil, fmt.Errorf("file type %s is not allowed", contentType)
	}

	// Validate file size
	if file.Size > s.config.MaxSize*1024*1024 {
		return nil, fmt.Errorf("file size exceeds the limit of %d MB", s.config.MaxSize)
	}

	// Based on storage type, upload to either local storage or S3
	if s.config.StorageType == StorageTypeS3 {
		return s.s3Client.UploadFile(file, subDir)
	}

	// Otherwise, upload to local storage
	return s.uploadToLocalStorage(file, src, contentType, subDir)
}

// uploadToLocalStorage saves the file to the local filesystem
func (s *Service) uploadToLocalStorage(file *multipart.FileHeader, src io.Reader, contentType, subDir string) (*UploadResult, error) {
	// Generate a unique filename
	filename := s.generateFilename(file.Filename, subDir)

	// Create the destination file
	uploadDir := s.config.GetUploadDir()
	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the file
	_, err = io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Construct the URL path
	urlPath := filepath.Join("/uploads", s.config.SubDir, filename)
	urlPath = strings.ReplaceAll(urlPath, "\\", "/") // Ensure forward slashes for URLs

	// Return the result
	return &UploadResult{
		Filename:    filename,
		Size:        file.Size,
		ContentType: contentType,
		Path:        filepath.Join(uploadDir, filename),
		URL:         urlPath,
	}, nil
}

// UploadMultiple handles multiple file uploads
func (s *Service) UploadMultiple(files []*multipart.FileHeader, subDir string) (*MultipleUploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// If using S3, we can use the batch upload capability
	if s.config.StorageType == StorageTypeS3 {
		results, err := s.s3Client.UploadMultipleFiles(files, subDir)
		if err != nil {
			return nil, err
		}
		return &MultipleUploadResult{Files: results}, nil
	}

	// Otherwise, upload each file individually to local storage
	results := make([]*UploadResult, 0, len(files))
	for _, file := range files {
		result, err := s.Upload(file, subDir)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return &MultipleUploadResult{Files: results}, nil
}

// Delete removes a file from storage
func (s *Service) Delete(filename string) error {
	// Validate filename to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename")
	}

	// Based on storage type, delete from either local storage or S3
	if s.config.StorageType == StorageTypeS3 {
		return s.s3Client.DeleteFile(filename)
	}

	// Otherwise, delete from local storage
	// Get the full path
	path := filepath.Join(s.config.GetUploadDir(), filename)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	// Delete the file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// generateFilename creates a unique filename for the uploaded file
func (s *Service) generateFilename(originalFilename, subDir string) string {
	// Extract file extension
	ext := filepath.Ext(originalFilename)

	// Generate random bytes
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to timestamp if random generation fails
		return fmt.Sprintf("%s_%d%s", subDir, time.Now().UnixNano(), ext)
	}

	// Format the filename with prefix, timestamp, and random string
	timestamp := time.Now().Format("20060102_150405")
	randomStr := hex.EncodeToString(randomBytes)

	if subDir != "" {
		return fmt.Sprintf("%s_%s_%s%s", subDir, timestamp, randomStr, ext)
	}

	return fmt.Sprintf("%s_%s%s", timestamp, randomStr, ext)
}
