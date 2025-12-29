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

// SearchResults writes a 200 OK response with search results.
func SearchResults(c echo.Context, results interface{}) error {
	return c.JSON(http.StatusOK, results)
}
