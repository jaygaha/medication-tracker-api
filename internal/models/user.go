// internal/models/user.go
package models

import (
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
