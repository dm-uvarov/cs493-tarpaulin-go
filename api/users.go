package api

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tarpaulin/auth"
	"tarpaulin/datastore"
	"tarpaulin/models"
	"tarpaulin/storage"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Remove Email field
}

// LoginHandler handles the /users/login endpoint
func LoginHandler(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			RespondWithError(c, http.StatusBadRequest) // Uses "The request body is invalid"
			return
		}

		// Use username for authentication
		username := req.Username

		// Validate username and password
		if username == "" || req.Password == "" {
			RespondWithError(c, http.StatusBadRequest) // Uses "The request body is invalid"
			return
		}

		// Authenticate with Auth0 using username
		token, err := authService.AuthenticateUser(username, req.Password)
		if err != nil {
			RespondWithError(c, http.StatusUnauthorized) // Uses "Unauthorized"
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// GetAllUsersHandler handles the GET /users endpoint
func GetAllUsersHandler(userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("DEBUG: GetAllUsersHandler started")

		// Get all users from datastore
		fmt.Println("DEBUG: Retrieving all users from datastore")
		users, err := userStore.ListUsers(c.Request.Context())
		if err != nil {
			fmt.Printf("DEBUG: Failed to retrieve users: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		fmt.Printf("DEBUG: Retrieved %d users from datastore\n", len(users))

		// Convert to summary format (only id, sub, role)
		userSummaries := make([]models.UserSummary, len(users))
		for i, user := range users {
			userSummaries[i] = models.UserSummary{
				ID:   user.ID,
				Sub:  user.Sub,
				Role: user.Role,
			}
		}

		fmt.Println("DEBUG: Returning user summaries")
		c.JSON(http.StatusOK, userSummaries)
	}
}

// GetUserHandler handles the GET /users/:id endpoint
func GetUserHandler(userStore *datastore.UserStore, courseStore *datastore.CourseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from URL parameter
		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)

		if err != nil {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Get the user from datastore
		user, err := userStore.GetUser(c.Request.Context(), userID)
		if err != nil {
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Get current user from context (set by auth middleware)
		currentUser, exists := c.Get("user")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized)
			return
		}

		currentUserObj := currentUser.(*models.User)

		// Check authorization: Admin or user with matching JWT
		if currentUserObj.Role != "admin" && currentUserObj.ID != userID {
			RespondWithError(c, http.StatusForbidden)
			return
		}

		// Create detailed user response
		userDetail := models.UserDetail{
			ID:   user.ID,
			Sub:  user.Sub,
			Role: user.Role,
		}

		// Set avatar URL in the expected format if user has an avatar
		if user.AvatarURL != "" {
			// Use the stored avatar URL directly (it should already be the full URL)
			userDetail.AvatarURL = user.AvatarURL
		}

		// Get courses for instructors and students
		if user.Role == "instructor" {
			// Get courses taught by this instructor
			courseIDs, err := courseStore.GetCoursesByInstructor(c.Request.Context(), user.ID)
			if err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to get courses for instructor %d: %v\n", user.ID, err)
				courseIDs = []int64{}
			}
			userDetail.Courses = &courseIDs // Set pointer to slice (even if empty)
		} else if user.Role == "student" {
			// Get courses this student is enrolled in
			courseIDs, err := userStore.GetUserCourses(c.Request.Context(), user.ID)
			if err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: Failed to get courses for student %d: %v\n", user.ID, err)
				courseIDs = []int64{}
			}
			userDetail.Courses = &courseIDs // Set pointer to slice (even if empty)
		}
		// Admin users don't have courses (Courses remains nil, so omitted from JSON)

		c.JSON(http.StatusOK, userDetail)
	}
}

// UploadAvatarHandler handles POST /users/:id/avatar
func UploadAvatarHandler(userStore *datastore.UserStore, storageService *storage.StorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Starting avatar upload request\n")
		fmt.Printf("DEBUG: Content-Type: %s\n", c.GetHeader("Content-Type"))

		// Print all form field names for debugging
		err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max memory
		if err != nil {
			fmt.Printf("DEBUG: Failed to parse multipart form: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "The request body is invalid"})
			return
		}

		fmt.Printf("DEBUG: Available form fields: %v\n", c.Request.MultipartForm.File)
		fmt.Printf("DEBUG: Available form values: %v\n", c.Request.MultipartForm.Value)

		userID := c.Param("id")
		fmt.Printf("DEBUG: User ID from URL: %s\n", userID)

		// Convert userID string to int64
		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			fmt.Printf("DEBUG: Invalid user ID format: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Check if user is authorized
		userFromContext, exists := c.Get("user")
		if !exists {
			fmt.Printf("DEBUG: No user found in context\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user := userFromContext.(*models.User)
		fmt.Printf("DEBUG: User from context: ID=%d\n", user.ID)

		if user.ID != userIDInt {
			fmt.Printf("DEBUG: User ID mismatch - Context: %d, URL: %d\n", user.ID, userIDInt)
			RespondWithError(c, http.StatusForbidden)
			return
		}
		fmt.Printf("DEBUG: Authorization successful\n")

		// Get the uploaded file using the correct form field name 'file'
		fmt.Printf("DEBUG: Attempting to get form file 'file'\n")
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			fmt.Printf("DEBUG: Failed to get form file 'file': %v\n", err)
			RespondWithError(c, http.StatusBadRequest)
			return
		}
		defer file.Close()
		fmt.Printf("DEBUG: Successfully got file: %s, size: %d\n", header.Filename, header.Size)

		// Validate file type
		contentType := header.Header.Get("Content-Type")
		fmt.Printf("DEBUG: File content type: %s\n", contentType)
		if !strings.HasPrefix(contentType, "image/") {
			fmt.Printf("DEBUG: Invalid content type: %s\n", contentType)
			RespondWithError(c, http.StatusBadRequest)
			return
		}
		fmt.Printf("DEBUG: File type validation successful\n")

		// Upload to Google Cloud Storage - FIX: Use int64 format consistently
		objectName := fmt.Sprintf("avatars/%d/avatar", userIDInt)
		fmt.Printf("DEBUG: Uploading to GCS with object name: %s\n", objectName)
		avatarURL, err := storageService.UploadFile(c.Request.Context(), objectName, file)
		if err != nil {
			fmt.Printf("DEBUG: Failed to upload to GCS: %v\n", err)
			RespondWithError(c, http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUG: Successfully uploaded to GCS: %s\n", avatarURL)

		// Update user's avatar URL in datastore using int64 ID
		fmt.Printf("DEBUG: Updating user avatar URL in datastore\n")
		// Get base URL from environment variable or construct from request
		baseURL := os.Getenv("APP_BASE_URL")
		if baseURL == "" {
			// Fallback to constructing from request
			// For App Engine, always use HTTPS as it handles TLS termination
			scheme := "https"
			// Check X-Forwarded-Proto header for the original protocol
			if proto := c.Request.Header.Get("X-Forwarded-Proto"); proto != "" {
				scheme = proto
			} else if c.Request.TLS == nil && c.Request.Host == "localhost" || strings.Contains(c.Request.Host, "127.0.0.1") {
				// Only use HTTP for local development
				scheme = "http"
			}
			baseURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
		}
		avatarAPIURL := fmt.Sprintf("%s/users/%s/avatar", baseURL, userID)
		fmt.Printf("DEBUG: Generated avatar URL: %s\n", avatarAPIURL)
		err = userStore.UpdateAvatar(c.Request.Context(), userIDInt, avatarAPIURL)
		if err != nil {
			fmt.Printf("DEBUG: Failed to update avatar in datastore: %v\n", err)
			RespondWithError(c, http.StatusInternalServerError)
			return
		}
		fmt.Printf("DEBUG: Successfully updated avatar URL in datastore: %s\n", avatarAPIURL)

		fmt.Printf("DEBUG: Avatar upload completed successfully\n")
		c.JSON(http.StatusOK, gin.H{"avatar_url": avatarAPIURL})
	}
}

// GetAvatarHandler handles GET /users/:id/avatar
func GetAvatarHandler(userStore *datastore.UserStore, storageService *storage.StorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: GetAvatarHandler started\n")

		// Get user ID from URL parameter
		userIDStr := c.Param("id")
		fmt.Printf("DEBUG: User ID from URL: %s\n", userIDStr)

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			fmt.Printf("DEBUG: Invalid user ID format: %v\n", err)
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Get current user from context (set by auth middleware)
		currentUser, exists := c.Get("user")
		if !exists {
			fmt.Printf("DEBUG: No user in context (unauthorized)\n")
			RespondWithError(c, http.StatusUnauthorized)
			return
		}

		currentUserObj := currentUser.(*models.User)
		fmt.Printf("DEBUG: Current user ID: %d, Requested user ID: %d\n", currentUserObj.ID, userID)

		// Check authorization: User with matching JWT
		if currentUserObj.ID != userID {
			fmt.Printf("DEBUG: Access forbidden - user %d trying to access user %d's avatar\n", currentUserObj.ID, userID)
			RespondWithError(c, http.StatusForbidden)
			return
		}

		// Get user to check if avatar exists
		fmt.Printf("DEBUG: Retrieving user from datastore\n")
		user, err := userStore.GetUser(c.Request.Context(), userID)
		if err != nil {
			fmt.Printf("DEBUG: User not found in datastore: %v\n", err)
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Check if user has an avatar
		fmt.Printf("DEBUG: User avatar URL in datastore: %s\n", user.AvatarURL)
		if user.AvatarURL == "" {
			fmt.Printf("DEBUG: User has no avatar URL set\n")
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Create object name for storage - FIXED: should match upload format
		objectName := fmt.Sprintf("avatars/%d/avatar", userID)
		fmt.Printf("DEBUG: Looking for object in GCS: %s\n", objectName)

		// Get file from Google Cloud Storage
		fmt.Printf("DEBUG: Calling storage service to get file\n")
		fileReader, err := storageService.GetFile(c.Request.Context(), objectName)
		if err != nil {
			fmt.Printf("DEBUG: Failed to get avatar from GCS: %v\n", err)
			RespondWithError(c, http.StatusNotFound)
			return
		}

		fmt.Printf("DEBUG: Successfully retrieved file from GCS, serving to client\n")
		// Set appropriate headers and stream the file
		c.Header("Content-Type", "image/jpeg") // You might want to store and retrieve the actual content type
		c.DataFromReader(http.StatusOK, -1, "image/jpeg", fileReader, nil)
		fmt.Printf("DEBUG: File served successfully\n")
	}
}

// DeleteAvatarHandler handles DELETE /users/:id/avatar
func DeleteAvatarHandler(userStore *datastore.UserStore, storageService *storage.StorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from URL parameter
		userIDStr := c.Param("id")
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Get current user from context (set by auth middleware)
		currentUser, exists := c.Get("user")
		if !exists {
			RespondWithError(c, http.StatusUnauthorized)
			return
		}

		currentUserObj := currentUser.(*models.User)

		// Check authorization: User with matching JWT
		if currentUserObj.ID != userID {
			RespondWithError(c, http.StatusForbidden)
			return
		}

		// Get user to check if avatar exists
		user, err := userStore.GetUser(c.Request.Context(), userID)
		if err != nil {
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Check if user has an avatar
		if user.AvatarURL == "" {
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Create object name for storage - FIXED: should match upload format
		objectName := fmt.Sprintf("avatars/%d/avatar", userID)

		// Delete file from Google Cloud Storage
		err = storageService.DeleteFile(c.Request.Context(), objectName)
		if err != nil {
			fmt.Printf("Failed to delete avatar: %v\n", err)
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		// Update user's avatar URL in datastore (remove it)
		err = userStore.UpdateAvatar(c.Request.Context(), userID, "")
		if err != nil {
			fmt.Printf("Failed to update avatar URL: %v\n", err)
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		// Return 204 No Content with no body
		c.Status(http.StatusNoContent)
	}
}
