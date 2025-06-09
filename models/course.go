package models

import "time"

// Course represents a course in the system
type Course struct {
	ID           int64     `json:"id" datastore:"-"`
	Subject      string    `json:"subject" datastore:"subject"`            // Subject code, up to 4 characters
	Number       int       `json:"number" datastore:"number"`              // Course number
	Title        string    `json:"title" datastore:"title"`                // Course title, up to 50 characters
	Term         string    `json:"term" datastore:"term"`                  // Term, up to 10 characters
	InstructorID int64     `json:"instructor_id" datastore:"instructorId"` // ID of the instructor (user with role instructor)
	CreatedAt    time.Time `json:"created_at" datastore:"createdAt"`
	UpdatedAt    time.Time `json:"updated_at" datastore:"updatedAt"`
}

// CourseSummary represents a summary of a course (for listing courses)
type CourseSummary struct {
	ID           int64  `json:"id"`
	Subject      string `json:"subject"`
	Number       int    `json:"number"`
	Title        string `json:"title"`
	Term         string `json:"term"`
	InstructorID int64  `json:"instructor_id"`
}

// CourseUpdates represents fields that can be updated for a course
type CourseUpdates struct {
	Subject      string `json:"subject,omitempty"`
	Number       int    `json:"number,omitempty"`
	Title        string `json:"title,omitempty"`
	Term         string `json:"term,omitempty"`
	InstructorID int64  `json:"instructor_id,omitempty"`
}
