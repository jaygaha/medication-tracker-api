// internal/models/user.go
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// NotificationPreference represents the 'notification_preference' enum in the database.
type NotificationPreference string

const (
	NotificationNone  NotificationPreference = "none"
	NotificationEmail NotificationPreference = "email"
	NotificationPush  NotificationPreference = "push"
	NotificationAll   NotificationPreference = "all"
)

type User struct {
	ID                     uuid.UUID              `json:"id" db:"id"`
	FirstName              string                 `json:"first_name" db:"first_name"`
	LastName               string                 `json:"last_name" db:"last_name"`
	Email                  string                 `json:"email" db:"email"`
	PasswordHash           string                 `json:"-" db:"password_hash"`
	Timezone               string                 `json:"timezone" db:"timezone"`
	NotificationPreference NotificationPreference `json:"notification_preference" db:"notification_preference"`
	DeletedAt              time.Time              `json:"-" db:"deleted_at"`
	CreatedAt              time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
}

// UpdateUserRequest DTO represents the request body for updating a user.
// All fields are optional so users can update single fields (like just the timezone) without sending the whole profile
type UpdateUserRequest struct {
	FirstName              *string                 `json:"first_name"`
	LastName               *string                 `json:"last_name"`
	Timezone               *string                 `json:"timezone"`
	NotificationPreference *NotificationPreference `json:"notification_preference"`
}

func (r *UpdateUserRequest) Validate() error {
	if r.NotificationPreference != nil {
		switch *r.NotificationPreference {
		case NotificationNone, NotificationEmail, NotificationPush, NotificationAll:
			// Valid
		default:
			return errors.New("invalid notification preference")
		}
	}
	if r.Timezone != nil {
		if *r.Timezone == "" {
			return errors.New("timezone cannot be empty")
		}

		// check if timezone is valid using load location.
		if _, err := time.LoadLocation(*r.Timezone); err != nil {
			return errors.New("invalid timezone")
		}
	}
	if r.FirstName != nil && *r.FirstName == "" {
		return errors.New("first name cannot be empty")
	}
	if r.LastName != nil && *r.LastName == "" {
		return errors.New("last name cannot be empty")
	}
	return nil
}
