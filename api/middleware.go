package api

import (
	"net/http"
	//"strings"
	"fmt"
	"strconv" // Add this import for string conversion
	"tarpaulin/auth"
	"tarpaulin/datastore" // Add this import

	//"tarpaulin/models"

	"github.com/gin-gonic/gin"
)

// AdminOnlyMiddleware restricts access to admin users only
func AdminOnlyMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("DEBUG: AdminOnlyMiddleware started")
		// Get user sub from context (set by AuthMiddleware)
		sub, exists := c.Get("user_sub")
		if !exists {
			fmt.Println("DEBUG: No user_sub found in context")
			// Use the standardized error response
			RespondWithError(c, http.StatusUnauthorized)
			c.Abort()
			return
		}
		fmt.Printf("DEBUG: Found user_sub in context: %v (type: %T)\n", sub, sub)

		// Get userStore from the application context
		userStore, exists := c.MustGet("userStore").(*datastore.UserStore) // Fix this line
		if !exists {
			fmt.Println("DEBUG: UserStore not found in application context")
			// Use the standardized error response
			RespondWithError(c, http.StatusInternalServerError)
			c.Abort()
			return
		}

		// Look up the user by sub in the datastore
		fmt.Printf("DEBUG: Looking up user with sub: %s\n", sub.(string))
		user, err := userStore.GetUserBySub(c.Request.Context(), sub.(string))
		if err != nil {
			fmt.Printf("DEBUG: Failed to find user with sub %s: %v\n", sub, err)
			// Use the standardized error response
			RespondWithError(c, http.StatusForbidden)
			c.Abort()
			return
		}

		// Check if user has admin role
		fmt.Printf("DEBUG: User found - ID: %d, Role: %s\n", user.ID, user.Role)
		if user.Role != "admin" {
			fmt.Printf("DEBUG: User with sub %s has role %s, not admin\n", sub, user.Role)
			// Use the standardized error response
			RespondWithError(c, http.StatusForbidden)
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: User with sub %s has admin role, access granted\n", sub)
		c.Next()
	}
}

// CourseInstructorMiddleware restricts access to course instructors or admin users
func CourseInstructorMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user sub from context (set by AuthMiddleware)
		sub, exists := c.Get("user_sub")
		if !exists {
			fmt.Println("No user_sub found in context")
			RespondWithError(c, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Get userStore from the application context
		userStore, exists := c.MustGet("userStore").(*datastore.UserStore)
		if !exists {
			fmt.Println("UserStore not found in application context")
			RespondWithError(c, http.StatusInternalServerError)
			c.Abort()
			return
		}

		// Get courseStore from the application context
		courseStore, exists := c.MustGet("courseStore").(*datastore.CourseStore)
		if !exists {
			fmt.Println("CourseStore not found in application context")
			RespondWithError(c, http.StatusInternalServerError)
			c.Abort()
			return
		}

		// Look up the user by sub in the datastore
		user, err := userStore.GetUserBySub(c.Request.Context(), sub.(string))
		if err != nil {
			fmt.Printf("Failed to find user with sub %s: %v\n", sub, err)
			RespondWithError(c, http.StatusForbidden)
			c.Abort()
			return
		}

		// Admin users have access to all courses
		if user.Role == "admin" {
			fmt.Printf("User with sub %s has admin role, access granted\n", sub)
			c.Next()
			return
		}

		// Get course ID from URL parameter
		courseID := c.Param("id")
		if courseID == "" {
			fmt.Println("No course ID found in URL parameters")
			RespondWithError(c, http.StatusBadRequest)
			c.Abort()
			return
		}

		// Convert string courseID to int64
		courseIDInt, err := strconv.ParseInt(courseID, 10, 64)
		if err != nil {
			fmt.Printf("Invalid course ID format: %s, error: %v\n", courseID, err)
			RespondWithError(c, http.StatusBadRequest)
			c.Abort()
			return
		}

		// Get the course with the converted int64 ID
		course, err := courseStore.GetCourse(c.Request.Context(), courseIDInt)
		if err != nil {
			fmt.Printf("Failed to find course with ID %d: %v\n", courseIDInt, err)
			RespondWithError(c, http.StatusNotFound)
			c.Abort()
			return
		}

		// Check if user is the instructor of the course
		if course.InstructorID != user.ID {
			fmt.Printf("User %d is not the instructor of course %d\n", user.ID, courseIDInt)
			RespondWithError(c, http.StatusForbidden)
			c.Abort()
			return
		}

		fmt.Printf("User %d is the instructor of course %d, access granted\n", user.ID, courseIDInt)
		c.Next()
	}
}
