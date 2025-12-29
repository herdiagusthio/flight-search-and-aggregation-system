package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Setup registers all middleware on the Echo instance in the correct order.
// The order is important:
//  1. RequestID - First, to generate/propagate request ID for all subsequent logging
//  2. RequestLogger - Second, logs all requests with request ID
//  3. Recover - Third, catches panics and returns 500 (wraps handlers)
//
// This function should be called before registering routes.
func Setup(e *echo.Echo, log zerolog.Logger) {
	e.Use(RequestID())
	e.Use(RequestLogger(log))
	e.Use(Recover(log))
}

// SetupWithConfig registers middleware with custom recovery configuration.
func SetupWithConfig(e *echo.Echo, log zerolog.Logger, recoveryConfig RecoveryConfig) {
	e.Use(RequestID())
	e.Use(RequestLogger(log))
	e.Use(RecoverWithConfig(log, recoveryConfig))
}

// Chain returns all middleware as a slice for use with route groups.
// Useful when you want to apply middleware to specific route groups only.
func Chain(log zerolog.Logger) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		RequestID(),
		RequestLogger(log),
		Recover(log),
	}
}
