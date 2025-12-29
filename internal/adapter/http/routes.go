// Package http provides the HTTP handler layer for the flight search API.
package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all flight search API routes.
// It creates a versioned API group and attaches the handler methods.
func RegisterRoutes(e *echo.Echo, h *FlightHandler) {
	// Health check endpoint (no version prefix)
	e.GET("/health", h.Health)

	// API v1 group
	api := e.Group("/api/v1")

	// Flights group
	flights := api.Group("/flights")
	flights.POST("/search", h.SearchFlights)
}

// RegisterRoutesWithMiddleware registers routes with custom middleware.
// This allows for endpoint-specific middleware configuration.
func RegisterRoutesWithMiddleware(e *echo.Echo, h *FlightHandler, middleware ...echo.MiddlewareFunc) {
	// Health check endpoint (no version prefix, no middleware)
	e.GET("/health", h.Health)

	// API v1 group with middleware
	api := e.Group("/api/v1", middleware...)

	// Flights group
	flights := api.Group("/flights")
	flights.POST("/search", h.SearchFlights)
}
