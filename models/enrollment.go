package models

import "time"

// Enrollment represents a student's enrollment in a course
type Enrollment struct {
	ID        int64     `json:"id" datastore:"-"`
	UserID    int64     `json:"userId" datastore:"userId"`     // ID of the enrolled student
	CourseID  int64     `json:"courseId" datastore:"courseId"` // ID of the course
	CreatedAt time.Time `json:"createdAt" datastore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" datastore:"updatedAt"`
}
