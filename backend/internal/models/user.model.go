package models

import "time"

// User represents a user in the system.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Mobile       string    `json:"mobile"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // don't expose passwords in JSON
	JoiningDate  time.Time `json:"joiningDate"`
	Address      string    `json:"address"`
	AvatarLink   string    `json:"avatarLink,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
