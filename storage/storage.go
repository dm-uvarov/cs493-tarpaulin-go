package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// StorageService handles file operations with Google Cloud Storage
type StorageService struct {
	Client     *storage.Client
	BucketName string
	ProjectID  string
}

// NewStorageService creates a new storage service
func NewStorageService() (*StorageService, error) {
	fmt.Printf("DEBUG: Creating new storage service\n")
	ctx := context.Background()

	// Get GCP project ID and bucket name from environment variables
	projectID := os.Getenv("GCP_PROJECT_ID")
	bucketName := os.Getenv("GCP_BUCKET_NAME")

	fmt.Printf("DEBUG: Project ID: %s, Bucket Name: %s\n", projectID, bucketName)

	if projectID == "" || bucketName == "" {
		fmt.Printf("DEBUG: Missing GCP environment variables\n")
		return nil, fmt.Errorf("missing required GCP environment variables")
	}

	// Create a storage client
	fmt.Printf("DEBUG: Creating storage client\n")
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("DEBUG: Failed to create storage client: %v\n", err)
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	fmt.Printf("DEBUG: Storage service created successfully\n")
	return &StorageService{
		Client:     client,
		BucketName: bucketName,
		ProjectID:  projectID,
	}, nil
}

// UploadFile uploads a file to Google Cloud Storage
func (s *StorageService) UploadFile(ctx context.Context, objectName string, file io.Reader) (string, error) {
	fmt.Printf("DEBUG: Starting file upload to bucket: %s, object: %s\n", s.BucketName, objectName)

	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)
	fmt.Printf("DEBUG: Got bucket handle\n")

	// Test bucket accessibility
	if _, err := bucket.Attrs(ctx); err != nil {
		fmt.Printf("DEBUG: Cannot access bucket %s: %v\n", s.BucketName, err)
		return "", fmt.Errorf("cannot access bucket %s: %v", s.BucketName, err)
	}
	fmt.Printf("DEBUG: Bucket %s is accessible\n", s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)
	fmt.Printf("DEBUG: Got object handle\n")

	// Create a writer to the object
	wc := obj.NewWriter(ctx)
	fmt.Printf("DEBUG: Created object writer\n")

	// Copy the file data to the object
	fmt.Printf("DEBUG: Starting file copy\n")
	bytesWritten, err := io.Copy(wc, file)
	if err != nil {
		fmt.Printf("DEBUG: io.Copy failed: %v\n", err)
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	fmt.Printf("DEBUG: Copied %d bytes\n", bytesWritten)

	// Close the writer to flush the data and save the object
	fmt.Printf("DEBUG: Closing writer\n")
	if err := wc.Close(); err != nil {
		fmt.Printf("DEBUG: Writer.Close failed: %v\n", err)
		return "", fmt.Errorf("Writer.Close: %v", err)
	}
	fmt.Printf("DEBUG: Writer closed successfully\n")

	// Note: Skipping ACL setting because uniform bucket-level access is enabled
	// With uniform bucket-level access, permissions are managed at the bucket level
	fmt.Printf("DEBUG: Skipping ACL setting (uniform bucket-level access enabled)\n")

	// Return the public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.BucketName, objectName)
	fmt.Printf("DEBUG: File uploaded successfully, public URL: %s\n", publicURL)
	return publicURL, nil
}

// GetFile gets a file from Google Cloud Storage
func (s *StorageService) GetFile(ctx context.Context, objectName string) (io.Reader, error) {
	fmt.Printf("DEBUG: Getting file from bucket: %s, object: %s\n", s.BucketName, objectName)

	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)

	// Create a reader for the object
	r, err := obj.NewReader(ctx)
	if err != nil {
		fmt.Printf("DEBUG: Object(%q).NewReader failed: %v\n", objectName, err)
		return nil, fmt.Errorf("Object(%q).NewReader: %v", objectName, err)
	}

	fmt.Printf("DEBUG: File reader created successfully\n")
	return r, nil
}

// DeleteFile deletes a file from Google Cloud Storage
func (s *StorageService) DeleteFile(ctx context.Context, objectName string) error {
	fmt.Printf("DEBUG: Deleting file from bucket: %s, object: %s\n", s.BucketName, objectName)

	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)

	// Delete the object
	if err := obj.Delete(ctx); err != nil {
		fmt.Printf("DEBUG: Object(%q).Delete failed: %v\n", objectName, err)
		return fmt.Errorf("Object(%q).Delete: %v", objectName, err)
	}

	fmt.Printf("DEBUG: File deleted successfully\n")
	return nil
}

// GenerateSignedURL generates a signed URL for uploading a file
func (s *StorageService) GenerateSignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	fmt.Printf("DEBUG: Generating signed URL for bucket: %s, object: %s\n", s.BucketName, objectName)

	// Generate a signed URL
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(expires),
		Scheme:  storage.SigningSchemeV4,
	}

	// Use the storage.SignedURL function
	url, err := storage.SignedURL(s.BucketName, objectName, opts)
	if err != nil {
		fmt.Printf("DEBUG: SignedURL failed: %v\n", err)
		return "", fmt.Errorf("SignedURL: %v", err)
	}

	fmt.Printf("DEBUG: Signed URL generated successfully: %s\n", url)
	return url, nil
}
