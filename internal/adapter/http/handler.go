// Package http provides the HTTP handler layer for the flight search API.
// It handles request parsing, validation, response formatting, and error mapping.
package http

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/http/response"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
)

// FlightHandler handles HTTP requests for flight-related endpoints.
type FlightHandler struct {
	useCase usecase.FlightSearchUseCase
}

// NewFlightHandler creates a new FlightHandler with the given use case.
func NewFlightHandler(uc usecase.FlightSearchUseCase) *FlightHandler {
	return &FlightHandler{
		useCase: uc,
	}
}

// SearchFlights handles POST /api/v1/flights/search
//
// @Summary Search for flights
// @Description Search for available flights across multiple providers
// @Tags flights
// @Accept json
// @Produce json
// @Param request body SearchFlightsRequest true "Search criteria"
// @Success 200 {object} domain.SearchResponse
// @Failure 400 {object} response.ErrorDetail "Validation error"
// @Failure 503 {object} response.ErrorDetail "Service unavailable"
// @Failure 504 {object} response.ErrorDetail "Gateway timeout"
// @Router /api/v1/flights/search [post]
func (h *FlightHandler) SearchFlights(c echo.Context) error {
	var req SearchFlightsRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		return response.InvalidRequestBody(c)
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return h.handleValidationError(c, err)
	}

	// Convert to domain types
	criteria := ToDomainCriteria(&req)
	opts := ToSearchOptions(&req)

	// Call use case with request context
	result, err := h.useCase.Search(c.Request().Context(), criteria, opts)
	if err != nil {
		return h.handleError(c, err)
	}

	// Return successful response
	return response.SearchResults(c, result)
}

// handleValidationError handles validation errors and returns a 400 response.
func (h *FlightHandler) handleValidationError(c echo.Context, err error) error {
	var validationErrs *ValidationErrors
	if errors.As(err, &validationErrs) {
		return response.ValidationError(c, validationErrs.ToMap())
	}

	// Fallback for non-structured validation errors
	return response.ValidationErrorWithMessage(c, err.Error())
}

// handleError maps domain errors to appropriate HTTP responses.
func (h *FlightHandler) handleError(c echo.Context, err error) error {
	// Check for all providers failed
	if errors.Is(err, domain.ErrAllProvidersFailed) {
		return response.ServiceUnavailable(c)
	}

	// Check for context deadline exceeded (timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return response.GatewayTimeout(c)
	}

	// Check for context cancelled
	if errors.Is(err, context.Canceled) {
		return response.RequestCancelled(c)
	}

	// Check for invalid request (domain validation)
	if errors.Is(err, domain.ErrInvalidRequest) {
		return response.ValidationErrorWithMessage(c, err.Error())
	}

	// Default to internal server error
	return response.InternalServerError(c)
}

// Health handles GET /health
// Simple health check endpoint.
func (h *FlightHandler) Health(c echo.Context) error {
	return response.Health(c)
}
