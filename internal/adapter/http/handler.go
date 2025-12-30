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
//	@Summary		Search for flights
//	@Description	Search for available flights across multiple airline providers (Garuda, Lion Air, Batik Air, AirAsia). Supports filtering by price, stops, airlines, departure time, arrival time, and flight duration.
//	@Tags			flights
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SearchFlightsRequest	true	"Search criteria with optional filters (maxPrice, maxStops, airlines, departureTimeRange, arrivalTimeRange, durationRange)"
//	@Success		200		{object}	SwaggerSearchResponse	"Successful search with flight results"
//	@Failure		400		{object}	SwaggerErrorResponse	"Validation error - invalid request parameters"
//	@Failure		503		{object}	SwaggerErrorResponse	"Service unavailable - all providers failed"
//	@Failure		504		{object}	SwaggerErrorResponse	"Gateway timeout - request took too long"
//	@Router			/flights/search [post]
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

	// Convert to DTO format matching expected output
	dto := ToSearchResponseDTO(result)

	// Return successful response
	return response.SearchResults(c, dto)
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
