// Package response provides standardized HTTP response builders for the flight search API.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// BadRequest writes a 400 Bad Request response with the given error message.
func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeInvalidRequest,
		Message: message,
	})
}

// InvalidRequestBody writes a 400 Bad Request response for malformed request bodies.
func InvalidRequestBody(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeInvalidRequest,
		Message: MsgInvalidRequestBody,
	})
}

// ValidationError writes a 400 Bad Request response with validation error details.
func ValidationError(c echo.Context, details map[string]string) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeValidationError,
		Message: MsgValidationFailed,
		Details: details,
	})
}

// ValidationErrorWithMessage writes a 400 Bad Request response with a custom message.
func ValidationErrorWithMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, &ErrorDetail{
		Code:    CodeValidationError,
		Message: message,
	})
}

// ServiceUnavailable writes a 503 Service Unavailable response.
func ServiceUnavailable(c echo.Context) error {
	return c.JSON(http.StatusServiceUnavailable, &ErrorDetail{
		Code:    CodeServiceUnavailable,
		Message: MsgServiceUnavailable,
	})
}

// ServiceUnavailableWithMessage writes a 503 Service Unavailable response with a custom message.
func ServiceUnavailableWithMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusServiceUnavailable, &ErrorDetail{
		Code:    CodeServiceUnavailable,
		Message: message,
	})
}

// GatewayTimeout writes a 504 Gateway Timeout response.
func GatewayTimeout(c echo.Context) error {
	return c.JSON(http.StatusGatewayTimeout, &ErrorDetail{
		Code:    CodeTimeout,
		Message: MsgTimeout,
	})
}

// RequestCancelled writes a 504 Gateway Timeout response for cancelled requests.
func RequestCancelled(c echo.Context) error {
	return c.JSON(http.StatusGatewayTimeout, &ErrorDetail{
		Code:    CodeTimeout,
		Message: MsgRequestCancelled,
	})
}

// InternalServerError writes a 500 Internal Server Error response.
func InternalServerError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, &ErrorDetail{
		Code:    CodeInternalError,
		Message: MsgInternalError,
	})
}

// InternalServerErrorWithMessage writes a 500 Internal Server Error response with a custom message.
func InternalServerErrorWithMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, &ErrorDetail{
		Code:    CodeInternalError,
		Message: message,
	})
}
