package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Client handles operations with AWS S3
type S3Client struct {
	session *session.Session
	bucket  string
	region  string
	prefix  string
}

// NewS3Client creates a new S3Client
func NewS3Client(accessKey, secretKey, region, bucket string, prefix string) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Client{
		session: sess,
		bucket:  bucket,
		region:  region,
		prefix:  prefix,
	}, nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(file *multipart.FileHeader, subDir string) (*UploadResult, error) {
	fmt.Println("Uploading file to S3")

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

	// Create a temp file
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "s3upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the file content to the temp file
	_, err = io.Copy(tempFile, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file to temp file: %w", err)
	}

	// Reset the temp file pointer
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to reset temp file pointer: %w", err)
	}

	// Read the entire file into memory
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	size := fileInfo.Size()
	fileBuffer := make([]byte, size)
	_, err = tempFile.Read(fileBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}

	// Extract base filename without directory structure
	baseFilename := filepath.Base(file.Filename)

	// Generate timestamp for the filename
	timestamp := time.Now().Format("20060102_150405")

	// Extract file extension
	ext := filepath.Ext(baseFilename)
	filenameWithoutExt := baseFilename[:len(baseFilename)-len(ext)]

	// Construct the S3 key with proper path structure
	// Add the filename with timestamp
	finalFilename := fmt.Sprintf("%s_%s%s", filenameWithoutExt, timestamp, ext)

	// Join all path segments properly (prefix, subDir, finalFilename)
	s3Key := joinS3Path(s.prefix, subDir, finalFilename)

	fmt.Println("Final S3 key:", s3Key)

	// Upload to S3
	_, err = s3.New(s.session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(s3Key),
		Body:                 bytes.NewReader(fileBuffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct the URL
	s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, s3Key)
	fmt.Println("Final S3 URL:", s3URL)

	// Return the result
	return &UploadResult{
		Filename:    finalFilename,
		Size:        size,
		ContentType: contentType,
		Path:        s3Key,
		URL:         s3URL,
	}, nil
}

// UploadMultipleFiles uploads multiple files to S3
func (s *S3Client) UploadMultipleFiles(files []*multipart.FileHeader, subDir string) ([]*UploadResult, error) {
	fmt.Println("Uploading multiple files to S3 with subdirectory:", subDir)
	results := make([]*UploadResult, 0, len(files))

	for _, file := range files {
		result, err := s.UploadFile(file, subDir)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(filename string) error {
	fmt.Println("Deleting file from S3:", filename)

	// Determine if this is a full path or just a filename
	var s3Key string

	// If it contains a slash, assume it's already a path
	if strings.Contains(filename, "/") {
		s3Key = filename
		fmt.Println("Using provided path as S3 key for deletion:", s3Key)
	} else {
		// Otherwise, apply the global prefix if set
		s3Key = joinS3Path(s.prefix, filename)
		fmt.Println("Applied global prefix to filename, final S3 key for deletion:", s3Key)
	}

	_, err := s3.New(s.session).DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	// Wait for the object to be deleted
	err = s3.New(s.session).WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to wait for file deletion: %w", err)
	}

	fmt.Println("Successfully deleted file from S3:", s3Key)
	return nil
}

// joinS3Path joins S3 path segments without creating double slashes
func joinS3Path(segments ...string) string {
	var result []string

	for _, segment := range segments {
		if segment == "" {
			continue
		}
		// Trim any leading or trailing slashes
		trimmed := strings.Trim(segment, "/")
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return strings.Join(result, "/")
}
