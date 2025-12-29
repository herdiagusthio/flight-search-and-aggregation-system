package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// RequestLogger returns middleware that logs HTTP requests.
// It logs on request completion with method, path, status, duration, and client info.
// The logger should be the zerolog.Logger instance from the logger package.
func RequestLogger(log zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request through handler chain
			err := next(c)
			if err != nil {
				// Let Echo's error handler process the error
				c.Error(err)
			}

			// Calculate duration after request completes
			duration := time.Since(start)

			// Get request ID from context (set by RequestID middleware)
			reqID := GetRequestID(c)

			// Get request details
			req := c.Request()
			res := c.Response()

			// Determine log level based on status code
			var event *zerolog.Event
			status := res.Status
			switch {
			case status >= 500:
				event = log.Error()
			case status >= 400:
				event = log.Warn()
			default:
				event = log.Info()
			}

			// Log the request with all relevant fields
			event.
				Str("request_id", reqID).
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("query", req.URL.RawQuery).
				Int("status", status).
				Int64("duration_ms", duration.Milliseconds()).
				Int64("bytes_out", res.Size).
				Str("client_ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Msg("HTTP request")

			// Return nil since we already handled the error via c.Error()
			return nil
		}
	}
}
