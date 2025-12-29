// Package http provides the HTTP handler layer for the flight search API.
package http

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	// Error is a machine-readable error code
	Error string `json:"error"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Details contains field-specific error details (for validation errors)
	Details map[string]string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(errorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   errorCode,
		Message: message,
	}
}

// NewValidationErrorResponse creates an error response for validation errors.
func NewValidationErrorResponse(message string, details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Error:   "validation_error",
		Message: message,
		Details: details,
	}
}

// Common error codes used in API responses.
const (
	// ErrCodeInvalidRequest indicates a malformed request body.
	ErrCodeInvalidRequest = "invalid_request"

	// ErrCodeValidationError indicates validation failures.
	ErrCodeValidationError = "validation_error"

	// ErrCodeServiceUnavailable indicates all providers are down.
	ErrCodeServiceUnavailable = "service_unavailable"

	// ErrCodeTimeout indicates a request timeout.
	ErrCodeTimeout = "timeout"

	// ErrCodeInternalError indicates an unexpected server error.
	ErrCodeInternalError = "internal_error"
)

// Common error messages.
const (
	MsgInvalidRequestBody     = "Failed to parse request body"
	MsgValidationFailed       = "Request validation failed"
	MsgServiceUnavailable     = "All flight providers are currently unavailable"
	MsgTimeout                = "Request timed out"
	MsgInternalError          = "An unexpected error occurred"
)

// Pre-defined error responses for common cases.
var (
	// ErrInvalidRequestBody is returned when the request body cannot be parsed.
	ErrInvalidRequestBody = NewErrorResponse(ErrCodeInvalidRequest, MsgInvalidRequestBody)

	// ErrServiceUnavailable is returned when all providers fail.
	ErrServiceUnavailable = NewErrorResponse(ErrCodeServiceUnavailable, MsgServiceUnavailable)

	// ErrRequestTimeout is returned when the request times out.
	ErrRequestTimeout = NewErrorResponse(ErrCodeTimeout, MsgTimeout)

	// ErrInternalServerError is returned for unexpected errors.
	ErrInternalServerError = NewErrorResponse(ErrCodeInternalError, MsgInternalError)
)
