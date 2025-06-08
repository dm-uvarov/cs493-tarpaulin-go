package api

import (
	"net/http"
	//"tarpaulin/auth"

	"github.com/gin-gonic/gin"
)

// CreateCourseHandler handles the POST /courses endpoint
func CreateCourseHandler(c *gin.Context) {
	// TODO: Implement course creation
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// GetAllCoursesHandler handles the GET /courses endpoint
func GetAllCoursesHandler(c *gin.Context) {
	// TODO: Implement fetching all courses
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// GetCourseHandler handles the GET /courses/:id endpoint
func GetCourseHandler(c *gin.Context) {
	// TODO: Implement fetching a specific course
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// UpdateCourseHandler handles the PATCH /courses/:id endpoint
func UpdateCourseHandler(c *gin.Context) {
	// TODO: Implement updating a course
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// DeleteCourseHandler handles the DELETE /courses/:id endpoint
func DeleteCourseHandler(c *gin.Context) {
	// TODO: Implement deleting a course
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// UpdateEnrollmentHandler handles the PATCH /courses/:id/students endpoint
func UpdateEnrollmentHandler(c *gin.Context) {
	// TODO: Implement updating course enrollment
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// GetEnrollmentHandler handles the GET /courses/:id/students endpoint
func GetEnrollmentHandler(c *gin.Context) {
	// TODO: Implement getting course enrollment
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}
