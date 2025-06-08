package models

import "time"

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" datastore:"-"`
	Sub       string    `json:"sub" datastore:"sub"`
	Role      string    `json:"role" datastore:"role"` // admin, instructor, or student
	AvatarURL string    `json:"avatarUrl,omitempty" datastore:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt" datastore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" datastore:"updatedAt"`
}

// UserSummary represents a summary of a user (for GET /users endpoint)
type UserSummary struct {
	ID   int64  `json:"id"`
	Sub  string `json:"sub"`
	Role string `json:"role"`
}

// UserUpdates represents fields that can be updated for a user
type UserUpdates struct {
	Role string `json:"role,omitempty"`
}

// UserDetail represents detailed user information for GET /users/:id
type UserDetail struct {
	ID        int64    `json:"id"`
	Sub       string   `json:"sub"`
	Role      string   `json:"role"`
	AvatarURL string   `json:"avatarUrl,omitempty"`
	Courses   *[]int64 `json:"courses,omitempty"` // Use pointer to distinguish between nil and empty slice
}
