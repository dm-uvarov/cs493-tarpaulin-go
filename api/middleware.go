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
func AdminOnlyMiddleware(authService *auth.AuthService, userStore *datastore.UserStore) gin.HandlerFunc {
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

		// Use the userStore parameter directly instead of getting from context
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
func CourseInstructorMiddleware(authService *auth.AuthService, userStore *datastore.UserStore, courseStore *datastore.CourseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user sub from context (set by AuthMiddleware)
		sub, exists := c.Get("user_sub")
		if !exists {
			fmt.Println("No user_sub found in context")
			RespondWithError(c, http.StatusUnauthorized)
			c.Abort()
			return
		}

		// Use the userStore and courseStore parameters directly instead of getting from context
		// Look up the user by sub in the datastore
		user, err := userStore.GetUserBySub(c.Request.Context(), sub.(string))
		if err != nil {
			fmt.Printf("Failed to find user with sub %s: %v\n", sub, err)
			RespondWithError(c, http.StatusForbidden)
			c.Abort()
			return
		}

		// Check if user has admin role (admins can access all courses)
		if user.Role == "admin" {
			c.Next()
			return
		}

		// For instructors, check if they are assigned to this course
		if user.Role == "instructor" {
			// Get course ID from URL parameter
			courseIDStr := c.Param("id")
			courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
			if err != nil {
				RespondWithError(c, http.StatusBadRequest)
				c.Abort()
				return
			}

			// Check if instructor is assigned to this course
			course, err := courseStore.GetCourse(c.Request.Context(), courseID)
			if err != nil {
				RespondWithError(c, http.StatusNotFound)
				c.Abort()
				return
			}

			if course.InstructorID == user.ID {
				c.Next()
				return
			}
		}

		// User is not authorized
		RespondWithError(c, http.StatusForbidden)
		c.Abort()
	}
}
