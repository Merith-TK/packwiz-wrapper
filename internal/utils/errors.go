package utils

import "fmt"

// Error templates for consistent error formatting
func NewFailedToError(operation string, err error) error {
	return fmt.Errorf("failed to %s: %w", operation, err)
}

func NewFailedToErrorf(operation string, err error, msg string, args ...interface{}) error {
	return fmt.Errorf("failed to %s: %s: %w", operation, fmt.Sprintf(msg, args...), err)
}
