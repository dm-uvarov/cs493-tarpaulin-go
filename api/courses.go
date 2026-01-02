package api

import (
	"fmt"
	"net/http"
	"strconv"
	"tarpaulin/datastore"
	"tarpaulin/models"
	"strings"
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
			"self":          "https://" + c.Request.Host + "/courses/" + strconv.FormatInt(course.ID, 10),
		}

		c.JSON(http.StatusOK, response)
	}
}

// UpdateCourseHandler handles the PATCH /courses/:id endpoint
func UpdateCourseHandler(courseStore *datastore.CourseStore, userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse course ID from URL
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest, "Invalid course ID")
			return
		}

		// Parse request body
		var updates models.CourseUpdates
		if bindErr := c.ShouldBindJSON(&updates); bindErr != nil {
			RespondWithError(c, http.StatusBadRequest)
			return
		}

		// If instructor_id is provided, validate that the instructor exists
		if updates.InstructorID != 0 {
			instructor, instructorErr := userStore.GetUser(c.Request.Context(), updates.InstructorID)
			if instructorErr != nil {
				if instructorErr.Error() == "user not found" {
					RespondWithError(c, http.StatusBadRequest)
					return
				}
				RespondWithError(c, http.StatusInternalServerError)
				return
			}
			if instructor.Role != "instructor" {
				RespondWithError(c, http.StatusBadRequest)
				return
			}
		}

		// Update the course
		updatedCourse, err := courseStore.UpdateCourse(c.Request.Context(), id, &updates)
		if err != nil {
			if err.Error() == "course not found" {
				RespondWithError(c, http.StatusNotFound)
				return
			}
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		// Return the updated course
		response := gin.H{
			"id":            updatedCourse.ID,
			"subject":       updatedCourse.Subject,
			"number":        updatedCourse.Number,
			"title":         updatedCourse.Title,
			"term":          updatedCourse.Term,
			"instructor_id": updatedCourse.InstructorID,
			"self":          "https://" + c.Request.Host + "/courses/" + strconv.FormatInt(updatedCourse.ID, 10),
		}

		c.JSON(http.StatusOK, response)
	}
}

// DeleteCourseHandler handles the DELETE /courses/:id endpoint
func DeleteCourseHandler(courseStore *datastore.CourseStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse course ID from URL
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			RespondWithError(c, http.StatusBadRequest, "Invalid course ID")
			return
		}

		// Check if course exists before deleting
		_, err = courseStore.GetCourse(c.Request.Context(), id)
		if err != nil {
			if err.Error() == "course not found" {
				RespondWithError(c, http.StatusNotFound, "Course not found")
			} else {
				RespondWithError(c, http.StatusInternalServerError, "Internal server error")
			}
			return
		}

		// Delete the course and all related data
		err = courseStore.DeleteCourse(c.Request.Context(), id)
		if err != nil {
			RespondWithError(c, http.StatusInternalServerError, "Failed to delete course")
			return
		}

		// Return 204 No Content on successful deletion
		c.Status(http.StatusNoContent)
	}
}

// UpdateEnrollmentHandler handles the PATCH /courses/:id/students endpoint
func UpdateEnrollmentHandler(courseStore *datastore.CourseStore, userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Starting enrollment update\n")
		
		// Parse course ID from URL
		courseIDStr := c.Param("id")
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Course ID string: %s\n", courseIDStr)
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Failed to parse course ID: %v\n", err)
			RespondWithError(c, http.StatusBadRequest)
			return
		}
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Parsed course ID: %d\n", courseID)

		// Parse request body
		var req struct {
			Add    []int64 `json:"add"`
			Remove []int64 `json:"remove"`
		}

		if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
			fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Failed to bind JSON: %v\n", bindErr)
			RespondWithError(c, http.StatusBadRequest)
			return
		}
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Request body - Add: %v, Remove: %v\n", req.Add, req.Remove)

		// Use the courseStore and userStore parameters directly instead of getting from context
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Using provided courseStore and userStore\n")

		// Validate that all users in add/remove lists exist and are students
		allUserIDs := append(req.Add, req.Remove...)
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Validating %d users\n", len(allUserIDs))
		for _, userID := range allUserIDs {
			fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Validating user ID: %d\n", userID)
			user, userErr := userStore.GetUser(c.Request.Context(), userID)
			if userErr != nil {
				fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Error getting user %d: %v\n", userID, userErr)
				if strings.Contains(userErr.Error(), "user not found") {
					RespondWithError(c, http.StatusBadRequest)
					return
				}
				RespondWithError(c, http.StatusInternalServerError)
				return
			}
			fmt.Printf("[DEBUG] UpdateEnrollmentHandler: User %d has role: %s\n", userID, user.Role)
			if user.Role != "student" {
				fmt.Printf("[DEBUG] UpdateEnrollmentHandler: User %d is not a student\n", userID)
				RespondWithError(c, http.StatusBadRequest)
				return
			}
		}

		// Check for conflicts (user in both add and remove)
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Checking for conflicts\n")
		for _, addID := range req.Add {
			for _, removeID := range req.Remove {
				if addID == removeID {
					fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Conflict detected - user %d in both add and remove\n", addID)
					RespondWithError(c, http.StatusConflict)
					return
				}
			}
		}

		// Update enrollment
		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Calling UpdateEnrollment for course %d\n", courseID)
		err = courseStore.UpdateEnrollment(c.Request.Context(), courseID, req.Add, req.Remove)
		if err != nil {
			fmt.Printf("[DEBUG] UpdateEnrollmentHandler: UpdateEnrollment failed: %v\n", err)
			if strings.Contains(err.Error(), "course not found") {
				RespondWithError(c, http.StatusNotFound)
				return
			}
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		fmt.Printf("[DEBUG] UpdateEnrollmentHandler: Successfully updated enrollment\n")
		c.Status(http.StatusOK)
	}
}

// GetEnrollmentHandler handles the GET /courses/:id/students endpoint
func GetEnrollmentHandler(courseStore *datastore.CourseStore, userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[DEBUG] GetEnrollmentHandler: Starting enrollment retrieval\n")
		
		// Parse course ID from URL
		courseIDStr := c.Param("id")
		fmt.Printf("[DEBUG] GetEnrollmentHandler: Course ID string: %s\n", courseIDStr)
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			fmt.Printf("[DEBUG] GetEnrollmentHandler: Failed to parse course ID: %v\n", err)
			RespondWithError(c, http.StatusBadRequest)
			return
		}
		fmt.Printf("[DEBUG] GetEnrollmentHandler: Parsed course ID: %d\n", courseID)

		// Use the courseStore parameter directly instead of getting from context
		fmt.Printf("[DEBUG] GetEnrollmentHandler: Using provided courseStore\n")

		// Get enrollment
		fmt.Printf("[DEBUG] GetEnrollmentHandler: Calling GetEnrollment for course %d\n", courseID)
		userIDs, err := courseStore.GetEnrollment(c.Request.Context(), courseID)
		if err != nil {
			fmt.Printf("[DEBUG] GetEnrollmentHandler: GetEnrollment failed: %v\n", err)
			if strings.Contains(err.Error(), "course not found") {
				RespondWithError(c, http.StatusNotFound)
				return
			}
			RespondWithError(c, http.StatusInternalServerError)
			return
		}

		fmt.Printf("[DEBUG] GetEnrollmentHandler: Successfully retrieved enrollment\n")
		c.JSON(http.StatusOK, gin.H{"students": userIDs})
	}
}
