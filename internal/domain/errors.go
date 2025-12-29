package domain

import (
	"errors"
	"fmt"
)

// Domain errors are sentinel errors that can be used with errors.Is() for comparison.
// They represent common failure scenarios in the flight search domain.
var (
	// ErrInvalidRequest indicates the request parameters are invalid (HTTP 400).
	// This error should be wrapped with specific details about what is invalid.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrAllProvidersFailed indicates all flight providers failed to respond (HTTP 503).
	// This typically means the service is temporarily unavailable.
	ErrAllProvidersFailed = errors.New("all providers failed")

	// ErrProviderTimeout indicates a specific provider timed out.
	// This is an internal error used during aggregation.
	ErrProviderTimeout = errors.New("provider timeout")

	// ErrProviderUnavailable indicates a provider is not reachable.
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrNoFlightsFound indicates no flights matched the search criteria.
	// This is not necessarily an error but useful for explicit handling.
	ErrNoFlightsFound = errors.New("no flights found")
)

// ProviderError wraps an error with provider context.
// It includes information about whether the error is retryable,
// which helps the use case layer decide on retry strategies.
type ProviderError struct {
	// Provider is the name/identifier of the provider that failed
	Provider string

	// Err is the underlying error
	Err error

	// Retryable indicates whether this error is transient and the operation
	// might succeed if retried. Examples of retryable errors:
	//   - Temporary network issues
	//   - Rate limiting (429)
	//   - Service temporarily unavailable (503)
	// Examples of non-retryable errors:
	//   - Invalid request parameters (400)
	//   - Authentication failures (401)
	//   - Resource not found (404)
	Retryable bool
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider %s: %v", e.Provider, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError.
// By default, errors are considered non-retryable.
func NewProviderError(provider string, err error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Err:       err,
		Retryable: false,
	}
}

// NewRetryableProviderError creates a new ProviderError marked as retryable.
func NewRetryableProviderError(provider string, err error) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Err:       err,
		Retryable: true,
	}
}

// NewProviderTimeoutError creates a timeout error for a specific provider.
func NewProviderTimeoutError(provider string) *ProviderError {
	return NewProviderError(provider, ErrProviderTimeout)
}

// NewProviderUnavailableError creates an unavailable error for a specific provider.
func NewProviderUnavailableError(provider string) *ProviderError {
	return NewProviderError(provider, ErrProviderUnavailable)
}

// ValidationError represents a validation error with field details.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error for a specific field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// WrapInvalidRequest wraps an error as an invalid request error.
func WrapInvalidRequest(format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}

// IsInvalidRequest checks if an error is an invalid request error.
func IsInvalidRequest(err error) bool {
	return errors.Is(err, ErrInvalidRequest)
}

// IsAllProvidersFailed checks if an error indicates all providers failed.
func IsAllProvidersFailed(err error) bool {
	return errors.Is(err, ErrAllProvidersFailed)
}

// IsProviderTimeout checks if an error is a provider timeout error.
func IsProviderTimeout(err error) bool {
	return errors.Is(err, ErrProviderTimeout)
}
