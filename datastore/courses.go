package datastore

import (
	"context"
	"fmt"
	"os"
	"tarpaulin/models"
	"time"

	"cloud.google.com/go/datastore"
	//"google.golang.org/api/iterator"
)

// CourseStore handles course data operations with Google Cloud Datastore
type CourseStore struct {
	Client    *datastore.Client
	ProjectID string
}

// NewCourseStore creates a new course store
func NewCourseStore() (*CourseStore, error) {
	ctx := context.Background()

	// Get GCP project ID from environment variables
	projectID := os.Getenv("GCP_PROJECT_ID")

	if projectID == "" {
		return nil, fmt.Errorf("missing required GCP_PROJECT_ID environment variable")
	}

	// Create a datastore client
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore client: %v", err)
	}

	return &CourseStore{
		Client:    client,
		ProjectID: projectID,
	}, nil
}

// CreateCourse creates a new course in Datastore
func (cs *CourseStore) CreateCourse(ctx context.Context, course *models.Course) (*models.Course, error) {
	// Create a new key
	key := datastore.IncompleteKey("Course", nil)

	// Set creation time
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	// Save the course
	newKey, err := cs.Client.Put(ctx, key, course)
	if err != nil {
		return nil, fmt.Errorf("failed to create course: %v", err)
	}

	// Set the ID
	course.ID = newKey.ID

	return course, nil
}

// GetCourse gets a course by ID
func (cs *CourseStore) GetCourse(ctx context.Context, id int64) (*models.Course, error) {
	// Create a key
	key := datastore.IDKey("Course", id, nil)

	// Get the course
	var course models.Course
	if err := cs.Client.Get(ctx, key, &course); err != nil {
		return nil, fmt.Errorf("failed to get course: %v", err)
	}

	// Set the ID
	course.ID = id

	return &course, nil
}

// UpdateCourse updates a course
func (cs *CourseStore) UpdateCourse(ctx context.Context, id int64, updates *models.CourseUpdates) (*models.Course, error) {
	// Get the course
	course, err := cs.GetCourse(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Subject != "" {
		course.Subject = updates.Subject
	}
	if updates.Number != 0 { // Check for non-zero instead of empty string
		course.Number = updates.Number
	}
	if updates.Title != "" {
		course.Title = updates.Title
	}
	if updates.Term != "" {
		course.Term = updates.Term
	}
	if updates.InstructorID != 0 {
		course.InstructorID = updates.InstructorID
	}

	// Update timestamp
	course.UpdatedAt = time.Now()

	// Create a key
	key := datastore.IDKey("Course", id, nil)

	// Save the course
	_, err = cs.Client.Put(ctx, key, course)
	if err != nil {
		return nil, fmt.Errorf("failed to update course: %v", err)
	}

	return course, nil
}

// DeleteCourse deletes a course
func (cs *CourseStore) DeleteCourse(ctx context.Context, id int64) error {
	// Create a key
	key := datastore.IDKey("Course", id, nil)

	// Delete the course
	if err := cs.Client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete course: %v", err)
	}

	return nil
}

// ListCourses lists all courses
func (cs *CourseStore) ListCourses(ctx context.Context) ([]*models.Course, error) {
	// Create a query
	query := datastore.NewQuery("Course")

	// Get all courses
	var courses []*models.Course
	keys, err := cs.Client.GetAll(ctx, query, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to list courses: %v", err)
	}

	// Set the IDs
	for i, key := range keys {
		courses[i].ID = key.ID
	}

	return courses, nil
}

// UpdateEnrollment updates a course's enrollment
func (cs *CourseStore) UpdateEnrollment(ctx context.Context, courseID int64, add []int64, remove []int64) error {
	// Verify the course exists
	_, err := cs.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	// Add enrollments
	for _, userID := range add {
		// Check if enrollment already exists
		query := datastore.NewQuery("Enrollment").Filter("userId =", userID).Filter("courseId =", courseID).Limit(1)
		var enrollments []*models.Enrollment
		_, err := cs.Client.GetAll(ctx, query, &enrollments)
		if err != nil {
			return fmt.Errorf("failed to check enrollment: %v", err)
		}

		// If enrollment doesn't exist, create it
		if len(enrollments) == 0 {
			enrollment := &models.Enrollment{
				UserID:    userID,
				CourseID:  courseID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create a new key
			key := datastore.IncompleteKey("Enrollment", nil)

			// Save the enrollment
			_, err := cs.Client.Put(ctx, key, enrollment)
			if err != nil {
				return fmt.Errorf("failed to create enrollment: %v", err)
			}
		}
	}

	// Remove enrollments
	for _, userID := range remove {
		// Find enrollment
		query := datastore.NewQuery("Enrollment").Filter("userId =", userID).Filter("courseId =", courseID)
		var enrollments []*models.Enrollment
		keys, err := cs.Client.GetAll(ctx, query, &enrollments)
		if err != nil {
			return fmt.Errorf("failed to find enrollment: %v", err)
		}

		// Delete enrollments
		for _, key := range keys {
			if err := cs.Client.Delete(ctx, key); err != nil {
				return fmt.Errorf("failed to delete enrollment: %v", err)
			}
		}
	}

	return nil
}

// GetEnrollment gets a course's enrollment (list of user IDs)
func (cs *CourseStore) GetEnrollment(ctx context.Context, courseID int64) ([]int64, error) {
	// Verify the course exists
	_, err := cs.GetCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// Query enrollments for this course
	query := datastore.NewQuery("Enrollment").Filter("courseId =", courseID)
	var enrollments []*models.Enrollment
	_, err = cs.Client.GetAll(ctx, query, &enrollments)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollments: %v", err)
	}

	// Extract user IDs
	userIDs := make([]int64, len(enrollments))
	for i, enrollment := range enrollments {
		userIDs[i] = enrollment.UserID
	}

	return userIDs, nil
}

// GetCoursesByInstructor gets all courses taught by an instructor
func (cs *CourseStore) GetCoursesByInstructor(ctx context.Context, instructorID int64) ([]int64, error) {
	// Query courses for this instructor
	query := datastore.NewQuery("Course").Filter("instructorId =", instructorID)
	var courses []*models.Course
	keys, err := cs.Client.GetAll(ctx, query, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses by instructor: %v", err)
	}

	// Extract course IDs
	courseIDs := make([]int64, len(courses))
	for i, key := range keys {
		courseIDs[i] = key.ID
	}

	return courseIDs, nil
}

// GetAllCoursesPaginated gets courses with pagination, ordered by subject
func (cs *CourseStore) GetAllCoursesPaginated(ctx context.Context, offset, limit int) ([]*models.Course, error) {
	// Create a query ordered by subject
	query := datastore.NewQuery("Course").Order("subject").Offset(offset).Limit(limit)

	// Get courses
	var courses []*models.Course
	keys, err := cs.Client.GetAll(ctx, query, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %v", err)
	}

	// Set the IDs
	for i, key := range keys {
		courses[i].ID = key.ID
	}

	return courses, nil
}
