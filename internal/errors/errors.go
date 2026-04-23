// internal/errors/errors.go
package errors

import "fmt"

// Database error types
type DatabaseError struct {
	Message string
	Err     error // The underlying error (for chaining)
}

// Implement Error interface
func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("database error: %s -> caused by: %s", e.Message, e.Err)
	}
	return fmt.Sprintf("database error: %s", e.Message)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, err error) *DatabaseError {
	return &DatabaseError{
		Message: message,
		Err:     err,
	}
}

// NotFoundError represents a "not found" error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}
