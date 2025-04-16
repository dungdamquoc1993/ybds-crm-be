package upload

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "upload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config
	config := NewConfig(tempDir)
	config.WithMaxSize(1) // 1MB

	// Create the service
	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create upload service: %v", err)
	}

	// Test the filename generation
	t.Run("FilenameGeneration", func(t *testing.T) {
		// Test with prefix
		filename := service.generateFilename("test.jpg", "prefix")
		if filename == "" {
			t.Errorf("Generated filename is empty")
		}
		if filepath.Ext(filename) != ".jpg" {
			t.Errorf("Expected extension .jpg, got %s", filepath.Ext(filename))
		}
		if len(filename) < 20 { // Should have prefix, timestamp, and random string
			t.Errorf("Filename too short: %s", filename)
		}

		// Test without prefix
		filename = service.generateFilename("test.jpg", "")
		if filename == "" {
			t.Errorf("Generated filename is empty")
		}
		if filepath.Ext(filename) != ".jpg" {
			t.Errorf("Expected extension .jpg, got %s", filepath.Ext(filename))
		}
	})

	// Test file deletion
	t.Run("FileDelete", func(t *testing.T) {
		// Create a test file
		testFilename := "delete_test.jpg"
		testPath := filepath.Join(tempDir, testFilename)
		if err := os.WriteFile(testPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Fatalf("Test file doesn't exist: %s", testPath)
		}

		// Delete the file
		err := service.Delete(testFilename)
		if err != nil {
			t.Errorf("Failed to delete file: %v", err)
		}

		// Check if file was deleted
		if _, err := os.Stat(testPath); !os.IsNotExist(err) {
			t.Errorf("File was not deleted: %s", testPath)
		}
	})

	// Test file deletion with invalid filename
	t.Run("InvalidFileDelete", func(t *testing.T) {
		err := service.Delete("../../../etc/passwd")
		if err == nil {
			t.Error("Should have rejected path traversal attempt")
		}

		err = service.Delete("nonexistent.jpg")
		if err == nil {
			t.Error("Should have reported an error for non-existent file")
		}
	})
}

func TestS3Upload(t *testing.T) {
	// Skip this test as it requires AWS credentials
	t.Skip("Skipping S3 tests - requires AWS credentials")
}
