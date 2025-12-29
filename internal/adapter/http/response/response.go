// Package response provides standardized HTTP response builders for the flight search API.
package response

// ErrorDetail contains structured error information.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Error codes used in API responses.
const (
	CodeInvalidRequest     = "invalid_request"
	CodeValidationError    = "validation_error"
	CodeServiceUnavailable = "service_unavailable"
	CodeTimeout            = "timeout"
	CodeInternalError      = "internal_error"
)

// Error messages used in API responses.
const (
	MsgInvalidRequestBody = "Failed to parse request body"
	MsgValidationFailed   = "Request validation failed"
	MsgServiceUnavailable = "All flight providers are currently unavailable"
	MsgTimeout            = "Request timed out"
	MsgRequestCancelled   = "Request was cancelled"
	MsgInternalError      = "An unexpected error occurred"
)
