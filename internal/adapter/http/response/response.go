// Package response provides standardized HTTP response builders for the flight search API.
// It centralizes response formatting to ensure consistency across all endpoints.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standardized API response envelope.
type Response struct {
	// Success indicates whether the request was successful
	Success bool `json:"success"`

	// Data contains the response payload (for successful responses)
	Data interface{} `json:"data,omitempty"`

	// Error contains error details (for error responses)
	Error *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail contains structured error information.
type ErrorDetail struct {
	// Code is a machine-readable error code
	Code string `json:"code"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Details contains field-specific error details (for validation errors)
	Details map[string]string `json:"details,omitempty"`
}

// Error codes used in API responses.
const (
	CodeInvalidRequest    = "invalid_request"
	CodeValidationError   = "validation_error"
	CodeServiceUnavailable = "service_unavailable"
	CodeTimeout           = "timeout"
	CodeInternalError     = "internal_error"
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

// JSON writes a JSON response with the given status code and data.
func JSON(c echo.Context, statusCode int, data interface{}) error {
	return c.JSON(statusCode, data)
}

// Success creates a successful response envelope.
func Success(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

// Failure creates a failed response envelope.
func Failure(code, message string, details map[string]string) *Response {
	return &Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// OK writes a 200 OK response with the given data.
func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// Created writes a 201 Created response with the given data.
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
