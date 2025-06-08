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
	ctx := context.Background()

	// Get GCP project ID and bucket name from environment variables
	projectID := os.Getenv("GCP_PROJECT_ID")
	bucketName := os.Getenv("GCP_BUCKET_NAME")

	if projectID == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required GCP environment variables")
	}

	// Create a storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &StorageService{
		Client:     client,
		BucketName: bucketName,
		ProjectID:  projectID,
	}, nil
}

// UploadFile uploads a file to Google Cloud Storage
func (s *StorageService) UploadFile(ctx context.Context, objectName string, file io.Reader) (string, error) {
	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)

	// Create a writer to the object
	wc := obj.NewWriter(ctx)

	// Copy the file data to the object
	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}

	// Close the writer to flush the data and save the object
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	// Make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("ACL.Set: %v", err)
	}

	// Return the public URL
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.BucketName, objectName), nil
}

// GetFile gets a file from Google Cloud Storage
func (s *StorageService) GetFile(ctx context.Context, objectName string) (io.Reader, error) {
	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)

	// Create a reader for the object
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %v", objectName, err)
	}

	return r, nil
}

// DeleteFile deletes a file from Google Cloud Storage
func (s *StorageService) DeleteFile(ctx context.Context, objectName string) error {
	// Get a handle to the bucket
	bucket := s.Client.Bucket(s.BucketName)

	// Get a handle to the object
	obj := bucket.Object(objectName)

	// Delete the object
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", objectName, err)
	}

	return nil
}

// GenerateSignedURL generates a signed URL for uploading a file
func (s *StorageService) GenerateSignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	// Get a handle to the bucket
	//bucket := s.Client.Bucket(s.BucketName)

	// No need to get the object handle since we're using storage.SignedURL directly
	// Remove this line: obj := bucket.Object(objectName)

	// Generate a signed URL
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(expires),
		Scheme:  storage.SigningSchemeV4,
	}

	// Use the storage.SignedURL function
	url, err := storage.SignedURL(s.BucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("SignedURL: %v", err)
	}

	return url, nil
}
