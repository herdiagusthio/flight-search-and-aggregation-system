// Package middleware provides HTTP middleware for cross-cutting concerns.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	// RequestIDHeader is the HTTP header name for request ID.
	RequestIDHeader = "X-Request-ID"
	// requestIDKey is the context key for storing request ID.
	requestIDKey = "request_id"
)

// RequestID returns middleware that generates or propagates request IDs.
// If the incoming request has an X-Request-ID header, it uses that value.
// Otherwise, it generates a new UUID.
// The request ID is stored in the context and added to response headers.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check for existing request ID from incoming header
			reqID := c.Request().Header.Get(RequestIDHeader)
			if reqID == "" {
				reqID = uuid.New().String()
			}

			// Set in context for use by handlers and other middleware
			c.Set(requestIDKey, reqID)

			// Set in response header for client correlation
			c.Response().Header().Set(RequestIDHeader, reqID)

			return next(c)
		}
	}
}

// GetRequestID retrieves the request ID from the echo context.
// Returns an empty string if no request ID is set.
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(requestIDKey).(string); ok {
		return id
	}
	return ""
}
