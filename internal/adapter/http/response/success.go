// Package response provides standardized HTTP response builders for the flight search API.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health writes a health check response.
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, &HealthResponse{
		Status: "ok",
	})
}

// Accepted writes a 202 Accepted response with the given data.
func Accepted(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusAccepted, data)
}

// List writes a 200 OK response with a list of items and optional pagination metadata.
func List(c echo.Context, items interface{}, meta interface{}) error {
	if meta != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"items": items,
			"meta":  meta,
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

// SearchResults writes a 200 OK response with search results.
// This is a convenience wrapper that returns the data directly without wrapping.
func SearchResults(c echo.Context, results interface{}) error {
	return c.JSON(http.StatusOK, results)
}
