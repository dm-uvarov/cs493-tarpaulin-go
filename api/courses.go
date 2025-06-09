package api

import (
	"net/http"
	"strconv"
	"tarpaulin/datastore"
	"tarpaulin/models"

	"github.com/gin-gonic/gin"
)

// CreateCourseHandler handles the POST /courses endpoint
func CreateCourseHandler(courseStore *datastore.CourseStore, userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var course models.Course
		if err := c.ShouldBindJSON(&course); err != nil {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if course.Subject == "" || course.Number == 0 || course.Title == "" || course.Term == "" || course.InstructorID == 0 {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Validate that instructor_id refers to a user with role 'instructor'
		instructor, err := userStore.GetUser(c.Request.Context(), course.InstructorID)
		if err != nil {
			// User not found
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		if instructor.Role != "instructor" {
			// User exists but is not an instructor
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Create the course
		createdCourse, err := courseStore.CreateCourse(c.Request.Context(), &course)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		// Create response with self URL
		response := gin.H{
			"id":            createdCourse.ID,
			"subject":       createdCourse.Subject,
			"number":        createdCourse.Number,
			"title":         createdCourse.Title,
			"term":          createdCourse.Term,
			"instructor_id": createdCourse.InstructorID,
			"created_at":    createdCourse.CreatedAt,
			"updated_at":    createdCourse.UpdatedAt,
			"self":          "https://" + c.Request.Host + "/courses/" + strconv.FormatInt(createdCourse.ID, 10),
		}

		c.JSON(http.StatusCreated, response)
	}
}

// GetAllCoursesHandler handles the GET /courses endpoint
func GetAllCoursesHandler(courseStore *datastore.CourseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters
		offset := 0
		limit := 3 // Default page size as specified

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Get paginated courses
		courses, err := courseStore.GetAllCoursesPaginated(c.Request.Context(), offset, limit)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		// Convert to summary format with self URLs
		courseSummaries := make([]gin.H, len(courses))
		for i, course := range courses {
			courseSummaries[i] = gin.H{
				"id":            course.ID,
				"subject":       course.Subject,
				"number":        course.Number,
				"title":         course.Title,
				"term":          course.Term,
				"instructor_id": course.InstructorID,
				"self":          "https://" + c.Request.Host + "/courses/" + strconv.FormatInt(course.ID, 10),
			}
		}

		// Create response with pagination
		response := gin.H{
			"courses": courseSummaries,
		}

		// Add next link if there are more courses
		if len(courses) == limit {
			nextOffset := offset + limit
			nextURL := "https://" + c.Request.Host + "/courses?offset=" + strconv.Itoa(nextOffset) + "&limit=" + strconv.Itoa(limit)
			response["next"] = nextURL
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCourseHandler handles the GET /courses/:id endpoint
func GetCourseHandler(courseStore *datastore.CourseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get course ID from URL parameter
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// Get the course from datastore
		course, err := courseStore.GetCourse(c.Request.Context(), id)
		if err != nil {
			// Course not found
			RespondWithError(c, http.StatusNotFound)
			return
		}

		// Create response with self URL
		response := gin.H{
			"id":            course.ID,
			"subject":       course.Subject,
			"number":        course.Number,
			"title":         course.Title,
			"term":          course.Term,
			"instructor_id": course.InstructorID,
			"created_at":    course.CreatedAt,
			"updated_at":    course.UpdatedAt,
			"self":          "https://" + c.Request.Host + "/courses/" + strconv.FormatInt(course.ID, 10),
		}

		c.JSON(http.StatusOK, response)
	}
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
