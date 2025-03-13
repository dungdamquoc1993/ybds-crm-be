package upload

import (
	"bytes"
	"io"
	"mime/multipart"
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

	// Create a test file
	testFile := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test image content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a multipart file header
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Create a buffer to write the file content
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy the file content to the form file
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	// Close the writer
	writer.Close()

	// Create a multipart file header for testing
	fileHeader := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     int64(buf.Len()),
		Header:   make(map[string][]string),
	}
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	// Test the upload function
	t.Run("Upload", func(t *testing.T) {
		// This is a simplified test since we can't easily create a real multipart.FileHeader
		// In a real test, you would use httptest to simulate a file upload

		// For now, we'll just test the filename generation and other functions
		filename := service.generateFilename("test.jpg", "prefix")
		if filename == "" {
			t.Errorf("Generated filename is empty")
		}

		if filepath.Ext(filename) != ".jpg" {
			t.Errorf("Expected extension .jpg, got %s", filepath.Ext(filename))
		}
	})

	// Test the delete function
	t.Run("Delete", func(t *testing.T) {
		// Create a test file to delete
		testDeleteFile := filepath.Join(tempDir, "delete_test.jpg")
		if err := os.WriteFile(testDeleteFile, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Delete the file
		err := service.Delete("delete_test.jpg")
		if err != nil {
			t.Errorf("Failed to delete file: %v", err)
		}

		// Check if the file was deleted
		_, err = os.Stat(testDeleteFile)
		if !os.IsNotExist(err) {
			t.Errorf("File was not deleted")
		}
	})
}
