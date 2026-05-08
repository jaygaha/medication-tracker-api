// internal/models/notification.go
package models

import (
	"errors"
	"time"
)

type DevicePlatform string

const (
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformWeb     DevicePlatform = "web"
)

type DeviceToken struct {
	ID        string         `json:"id" db:"id"`
	UserID    string         `json:"-" db:"user_id"`
	Token     string         `json:"token" db:"token"`
	Platform  DevicePlatform `json:"platform" db:"platform"`
	LastUsed  *time.Time     `json:"last_used" db:"last_used"`
	IsActive  bool           `json:"is_active" db:"is_active"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time     `json:"-" db:"deleted_at"`
}

type RegisterTokenRequest struct {
	Token    string         `json:"token"    binding:"required"`
	Platform DevicePlatform `json:"platform" binding:"required"`
}

func (r *RegisterTokenRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}
	if r.Platform == "" {
		return errors.New("platform is required")
	}

	switch r.Platform {
	case DevicePlatformAndroid, DevicePlatformIOS, DevicePlatformWeb:
		// Valid platform
	default:
		return errors.New("invalid platform")
	}
	return nil
}
