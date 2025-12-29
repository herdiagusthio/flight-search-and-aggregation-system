// Package http provides the HTTP handler layer for the flight search API.
// It handles request parsing, validation, and response formatting.
package http

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SearchFlightsRequest represents the request body for flight search.
type SearchFlightsRequest struct {
	// Origin is the IATA code of the departure airport (e.g., "CGK")
	Origin string `json:"origin"`

	// Destination is the IATA code of the arrival airport (e.g., "DPS")
	Destination string `json:"destination"`

	// DepartureDate is the desired departure date in YYYY-MM-DD format
	DepartureDate string `json:"departureDate"`

	// Passengers is the number of passengers (1-9)
	Passengers int `json:"passengers"`

	// Class is the travel class: economy, business, or first (optional)
	Class string `json:"class,omitempty"`

	// Filters contains optional filtering criteria
	Filters *FilterDTO `json:"filters,omitempty"`

	// SortBy specifies how to sort results: best_value, price, duration, departure
	SortBy string `json:"sortBy,omitempty"`
}

// FilterDTO represents optional filters for flight search.
type FilterDTO struct {
	// MaxPrice filters flights with price above this amount
	MaxPrice *float64 `json:"maxPrice,omitempty"`

	// MaxStops filters flights with more stops than this value (0 = direct only)
	MaxStops *int `json:"maxStops,omitempty"`

	// Airlines filters to only include flights from these airline codes
	Airlines []string `json:"airlines,omitempty"`

	// DepartureTimeRange filters flights departing within a time window
	DepartureTimeRange *TimeRangeDTO `json:"departureTimeRange,omitempty"`
}

// TimeRangeDTO represents a time window for filtering.
type TimeRangeDTO struct {
	// Start is the beginning of the time range (HH:MM format, e.g., "06:00")
	Start string `json:"start"`

	// End is the end of the time range (HH:MM format, e.g., "12:00")
	End string `json:"end"`
}

// Validation regex patterns.
var (
	airportCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)
	datePattern        = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timePattern        = regexp.MustCompile(`^\d{2}:\d{2}$`)
)

// Valid travel classes.
var validClasses = map[string]bool{
	"economy":  true,
	"business": true,
	"first":    true,
	"":         true, // Empty is valid (defaults to economy)
}

// Valid sort options.
var validSortOptions = map[string]bool{
	"best":      true, // Alias for best_value
	"price":     true,
	"duration":  true,
	"departure": true,
	"":          true, // Empty is valid (defaults to best_value)
}

// ValidationError represents a field-level validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors holds multiple validation errors.
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface.
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	return v.Errors[0].Message
}

// Add adds a validation error.
func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors.
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// ToMap converts validation errors to a map for API response.
func (v *ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string, len(v.Errors))
	for _, e := range v.Errors {
		result[e.Field] = e.Message
	}
	return result
}

// Validate validates the search request and returns any validation errors.
func (r *SearchFlightsRequest) Validate() error {
	errs := &ValidationErrors{}

	// Validate origin
	r.validateOrigin(errs)

	// Validate destination
	r.validateDestination(errs)

	// Check origin != destination
	r.validateOriginDestinationDifferent(errs)

	// Validate departure date
	r.validateDepartureDate(errs)

	// Validate passengers
	r.validatePassengers(errs)

	// Validate class
	r.validateClass(errs)

	// Validate sort option
	r.validateSortBy(errs)

	// Validate filters
	r.validateFilters(errs)

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func (r *SearchFlightsRequest) validateOrigin(errs *ValidationErrors) {
	if r.Origin == "" {
		errs.Add("origin", "origin is required")
		return
	}

	origin := strings.ToUpper(r.Origin)
	if !airportCodePattern.MatchString(origin) {
		errs.Add("origin", "origin must be a valid 3-letter IATA airport code")
		return
	}
	r.Origin = origin // Normalize to uppercase
}

func (r *SearchFlightsRequest) validateDestination(errs *ValidationErrors) {
	if r.Destination == "" {
		errs.Add("destination", "destination is required")
		return
	}

	dest := strings.ToUpper(r.Destination)
	if !airportCodePattern.MatchString(dest) {
		errs.Add("destination", "destination must be a valid 3-letter IATA airport code")
		return
	}
	r.Destination = dest // Normalize to uppercase
}

func (r *SearchFlightsRequest) validateOriginDestinationDifferent(errs *ValidationErrors) {
	if r.Origin != "" && r.Destination != "" &&
		strings.EqualFold(r.Origin, r.Destination) {
		errs.Add("destination", "origin and destination must be different")
	}
}

func (r *SearchFlightsRequest) validateDepartureDate(errs *ValidationErrors) {
	if r.DepartureDate == "" {
		errs.Add("departureDate", "departureDate is required")
		return
	}

	if !datePattern.MatchString(r.DepartureDate) {
		errs.Add("departureDate", "departureDate must be in YYYY-MM-DD format")
		return
	}

	_, err := time.Parse("2006-01-02", r.DepartureDate)
	if err != nil {
		errs.Add("departureDate", "departureDate is not a valid date")
		return
	}
}

func (r *SearchFlightsRequest) validatePassengers(errs *ValidationErrors) {
	if r.Passengers < 1 {
		errs.Add("passengers", "passengers must be at least 1")
		return
	}
	if r.Passengers > 9 {
		errs.Add("passengers", "passengers cannot exceed 9")
	}
}

func (r *SearchFlightsRequest) validateClass(errs *ValidationErrors) {
	if !validClasses[strings.ToLower(r.Class)] {
		errs.Add("class", "class must be one of: economy, business, first")
	}
}

func (r *SearchFlightsRequest) validateSortBy(errs *ValidationErrors) {
	if !validSortOptions[strings.ToLower(r.SortBy)] {
		errs.Add("sortBy", "sortBy must be one of: best, price, duration, departure")
	}
}

func (r *SearchFlightsRequest) validateFilters(errs *ValidationErrors) {
	if r.Filters == nil {
		return
	}

	// Validate maxPrice
	if r.Filters.MaxPrice != nil && *r.Filters.MaxPrice < 0 {
		errs.Add("filters.maxPrice", "maxPrice must be a positive number")
	}

	// Validate maxStops
	if r.Filters.MaxStops != nil && *r.Filters.MaxStops < 0 {
		errs.Add("filters.maxStops", "maxStops must be a non-negative number")
	}

	// Validate airline codes
	for i, airline := range r.Filters.Airlines {
		normalized := strings.ToUpper(airline)
		if len(normalized) < 2 || len(normalized) > 3 {
			errs.Add(fmt.Sprintf("filters.airlines[%d]", i),
				"airline code must be 2 or 3 characters")
		}
		r.Filters.Airlines[i] = normalized
	}

	// Validate departure time range
	if r.Filters.DepartureTimeRange != nil {
		r.validateTimeRange(errs)
	}
}

func (r *SearchFlightsRequest) validateTimeRange(errs *ValidationErrors) {
	tr := r.Filters.DepartureTimeRange

	if tr.Start == "" {
		errs.Add("filters.departureTimeRange.start", "start time is required when departureTimeRange is specified")
	} else if !timePattern.MatchString(tr.Start) {
		errs.Add("filters.departureTimeRange.start", "start must be in HH:MM format")
	}

	if tr.End == "" {
		errs.Add("filters.departureTimeRange.end", "end time is required when departureTimeRange is specified")
	} else if !timePattern.MatchString(tr.End) {
		errs.Add("filters.departureTimeRange.end", "end must be in HH:MM format")
	}
}
