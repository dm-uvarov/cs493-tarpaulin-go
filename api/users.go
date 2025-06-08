package api

import (
	//"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	//"strings"
	"tarpaulin/auth"
	"tarpaulin/datastore"
	"tarpaulin/models"

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
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid request body"})
			return
		}

		// Use username for authentication
		username := req.Username

		// Validate username and password
		if username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Username and password are required"})
			return
		}

		// Authenticate with Auth0 using username
		token, err := authService.AuthenticateUser(username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "Invalid credentials"})
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
			ID:        user.ID,
			Sub:       user.Sub,
			Role:      user.Role,
			AvatarURL: user.AvatarURL,
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
