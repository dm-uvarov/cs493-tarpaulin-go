package datastore

import (
	"context"
	"fmt"
	"os"
	"tarpaulin/models"
	"time"

	"cloud.google.com/go/datastore"
)

// UserStore handles user data operations with Google Cloud Datastore
type UserStore struct {
	Client    *datastore.Client
	ProjectID string
}

// NewUserStore creates a new user store
func NewUserStore() (*UserStore, error) {
	ctx := context.Background()

	// Get GCP project ID from environment variables
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("missing required GCP_PROJECT_ID environment variable")
	}

	// Debug logging for database connection
	fmt.Printf("DEBUG: Connecting to Datastore - Project: %s, Database: (default)\n", projectID)

	// Create a datastore client for default database
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore client: %v", err)
	}

	fmt.Printf("DEBUG: Connected to default database successfully\n")

	return &UserStore{
		Client:    client,
		ProjectID: projectID,
	}, nil
}

// CreateUser creates a new user in Datastore
func (us *UserStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Create a new key
	key := datastore.IncompleteKey("users", nil)

	// Set creation time
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Save the user
	newKey, err := us.Client.Put(ctx, key, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Set the ID
	user.ID = newKey.ID

	return user, nil
}

// GetUser gets a user by ID
func (us *UserStore) GetUser(ctx context.Context, id int64) (*models.User, error) {
	// Create a key
	key := datastore.IDKey("users", id, nil)

	// Get the user
	var user models.User
	if err := us.Client.Get(ctx, key, &user); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	// Set the ID
	user.ID = id

	return &user, nil
}

// GetUserByEmail gets a user by email
// Remove this entire function (lines 84-95)
// func (us *UserStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
//     // Query for user by email
//     query := datastore.NewQuery("users").Filter("email =", email).Limit(1)
//     var users []*models.User
//     keys, err := us.Client.GetAll(ctx, query, &users)
//     if err != nil {
//         return nil, fmt.Errorf("failed to get user by email: %v", err)
//     }
//
//     if len(users) == 0 {
//         return nil, fmt.Errorf("user not found")
//     }
//
//     // Set the ID
//     users[0].ID = keys[0].ID
//
//     return users[0], nil
// }

// GetUserBySub gets a user by their Auth0 sub claim
func (us *UserStore) GetUserBySub(ctx context.Context, sub string) (*models.User, error) {
	fmt.Printf("DEBUG: Looking up user with sub: %s\n", sub)

	// Create a query
	query := datastore.NewQuery("users").Filter("sub =", sub).Limit(1)

	// Get the user
	var users []*models.User
	keys, err := us.Client.GetAll(ctx, query, &users)
	if err != nil {
		fmt.Printf("DEBUG: Error looking up user with sub %s: %v\n", sub, err)
		return nil, fmt.Errorf("failed to get user by sub: %v", err)
	}

	if len(users) == 0 {
		fmt.Printf("DEBUG: No user found with sub: %s\n", sub)
		return nil, fmt.Errorf("user not found")
	}

	// Set the ID
	users[0].ID = keys[0].ID
	fmt.Printf("DEBUG: Found user with sub %s: ID=%d, Role=%s\n", sub, users[0].ID, users[0].Role)

	return users[0], nil
}

// ListUsers lists all users
func (us *UserStore) ListUsers(ctx context.Context) ([]*models.User, error) {
	// Create a query
	query := datastore.NewQuery("users")

	// Get all users
	var users []*models.User
	keys, err := us.Client.GetAll(ctx, query, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}

	// Set the IDs
	for i, key := range keys {
		users[i].ID = key.ID
	}

	return users, nil
}

// UpdateUser updates a user
func (us *UserStore) UpdateUser(ctx context.Context, id int64, updates *models.UserUpdates) (*models.User, error) {
	// Get the user
	user, err := us.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Role != "" {
		user.Role = updates.Role
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Create a key
	key := datastore.IDKey("users", id, nil)

	// Save the user
	_, err = us.Client.Put(ctx, key, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (us *UserStore) DeleteUser(ctx context.Context, id int64) error {
	// Create a key
	key := datastore.IDKey("User", id, nil)

	// Delete the user
	if err := us.Client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}

// UpdateAvatar updates a user's avatar URL
func (us *UserStore) UpdateAvatar(ctx context.Context, id int64, avatarURL string) error {
	// Get the user
	user, err := us.GetUser(ctx, id)
	if err != nil {
		return err
	}

	// Update avatar URL
	user.AvatarURL = avatarURL

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Create a key
	key := datastore.IDKey("users", id, nil)

	// Save the user
	_, err = us.Client.Put(ctx, key, user)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %v", err)
	}

	return nil
}

// HealthCheck checks if the datastore connection is working
func (us *UserStore) HealthCheck(ctx context.Context) error {
	// Create a simple query that doesn't return any results but will verify connection
	query := datastore.NewQuery("User").Limit(1)

	// Run the query
	_, err := us.Client.Count(ctx, query)
	return err
}

// GetUserCourses gets all courses a user is enrolled in
func (us *UserStore) GetUserCourses(ctx context.Context, userID int64) ([]int64, error) {
	// Verify the user exists
	_, err := us.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Query enrollments for this user
	query := datastore.NewQuery("Enrollment").Filter("userId =", userID)
	var enrollments []*models.Enrollment
	_, err = us.Client.GetAll(ctx, query, &enrollments)
	if err != nil {
		return nil, fmt.Errorf("failed to get user enrollments: %v", err)
	}

	// Extract course IDs
	courseIDs := make([]int64, len(enrollments))
	for i, enrollment := range enrollments {
		courseIDs[i] = enrollment.CourseID
	}

	return courseIDs, nil
}
