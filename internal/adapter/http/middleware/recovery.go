package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// DefaultRecoveryConfig returns the default recovery configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
}

// Recover returns middleware that recovers from panics in the handler chain.
// It logs the panic with stack trace and returns a 500 Internal Server Error.
// The server continues to handle subsequent requests.
func Recover(log zerolog.Logger) echo.MiddlewareFunc {
	return RecoverWithConfig(log, DefaultRecoveryConfig())
}

// RecoverWithConfig returns recovery middleware with custom configuration.
func RecoverWithConfig(log zerolog.Logger, config RecoveryConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Get request ID for correlation
					reqID := GetRequestID(c)

					// Build error message from panic value
					var panicMsg string
					if err, ok := r.(error); ok {
						panicMsg = err.Error()
					} else {
						panicMsg = fmt.Sprintf("%v", r)
					}

					// Log panic with stack trace
					event := log.Error().
						Str("request_id", reqID).
						Str("panic", panicMsg)

					if !config.DisablePrintStack {
						event = event.Str("stack", string(debug.Stack()))
					}

					event.Msg("Panic recovered")

					// Return 500 Internal Server Error
					// Use a generic error response to avoid leaking internal details
					if !c.Response().Committed {
						c.JSON(http.StatusInternalServerError, map[string]interface{}{
							"success": false,
							"error": map[string]string{
								"code":    "internal_error",
								"message": "An unexpected error occurred",
							},
						})
					}
				}
			}()

			return next(c)
		}
	}
}
