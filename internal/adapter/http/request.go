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
// Example: {"maxPrice": 1000000, "maxStops": 0, "departureTimeRange": {"start": "06:00", "end": "12:00"}, "arrivalTimeRange": {"start": "08:00", "end": "17:00"}}
type FilterDTO struct {
	// MaxPrice filters flights with price above this amount
	MaxPrice *float64 `json:"maxPrice,omitempty" example:"1000000"`

	// MaxStops filters flights with more stops than this value (0 = direct only)
	MaxStops *int `json:"maxStops,omitempty" example:"0"`

	// Airlines filters to only include flights from these airline codes
	Airlines []string `json:"airlines,omitempty" example:"GA,JT"`

	// DepartureTimeRange filters flights departing within a time window
	DepartureTimeRange *TimeRangeDTO `json:"departureTimeRange,omitempty"`

	// ArrivalTimeRange filters flights arriving within a time window
	ArrivalTimeRange *TimeRangeDTO `json:"arrivalTimeRange,omitempty"`

	// DurationRange filters flights by total duration in minutes
	DurationRange *DurationRangeDTO `json:"durationRange,omitempty"`
}

// TimeRangeDTO represents a time window for filtering.
type TimeRangeDTO struct {
	// Start is the beginning of the time range (HH:MM format, e.g., "06:00")
	Start string `json:"start"`

	// End is the end of the time range (HH:MM format, e.g., "12:00")
	End string `json:"end"`
}

// DurationRangeDTO represents a duration range filter in minutes.
// Example: {"minMinutes": 60, "maxMinutes": 180} filters flights between 1-3 hours.
type DurationRangeDTO struct {
	// MinMinutes is the minimum acceptable flight duration in minutes
	// Example: 60 (1 hour)
	MinMinutes *int `json:"minMinutes,omitempty" example:"60"`

	// MaxMinutes is the maximum acceptable flight duration in minutes
	// Example: 180 (3 hours)
	MaxMinutes *int `json:"maxMinutes,omitempty" example:"180"`
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
		r.validateDepartureTimeRange(errs)
	}

	// Validate arrival time range
	if r.Filters.ArrivalTimeRange != nil {
		r.validateArrivalTimeRange(errs)
	}

	// Validate duration range
	if r.Filters.DurationRange != nil {
		r.validateDurationRange(errs)
	}
}

func (r *SearchFlightsRequest) validateDepartureTimeRange(errs *ValidationErrors) {
	tr := r.Filters.DepartureTimeRange

	if tr.Start == "" {
		errs.Add("filters.departureTimeRange.start", "start time is required when departureTimeRange is specified")
	} else if !isValidTimeFormat(tr.Start) {
		errs.Add("filters.departureTimeRange.start", "start must be in HH:MM format with valid hours (00-23) and minutes (00-59)")
	}

	if tr.End == "" {
		errs.Add("filters.departureTimeRange.end", "end time is required when departureTimeRange is specified")
	} else if !isValidTimeFormat(tr.End) {
		errs.Add("filters.departureTimeRange.end", "end must be in HH:MM format with valid hours (00-23) and minutes (00-59)")
	}
}

func (r *SearchFlightsRequest) validateArrivalTimeRange(errs *ValidationErrors) {
	tr := r.Filters.ArrivalTimeRange

	if tr.Start == "" {
		errs.Add("filters.arrivalTimeRange.start", "start time is required when arrivalTimeRange is specified")
	} else if !isValidTimeFormat(tr.Start) {
		errs.Add("filters.arrivalTimeRange.start", "start must be in HH:MM format with valid hours (00-23) and minutes (00-59)")
	}

	if tr.End == "" {
		errs.Add("filters.arrivalTimeRange.end", "end time is required when arrivalTimeRange is specified")
	} else if !isValidTimeFormat(tr.End) {
		errs.Add("filters.arrivalTimeRange.end", "end must be in HH:MM format with valid hours (00-23) and minutes (00-59)")
	}
}

func (r *SearchFlightsRequest) validateDurationRange(errs *ValidationErrors) {
	dr := r.Filters.DurationRange

	// Validate minimum duration
	if dr.MinMinutes != nil {
		if *dr.MinMinutes < 0 {
			errs.Add("filters.durationRange.minMinutes", "minMinutes must be a non-negative number")
		}
	}

	// Validate maximum duration
	if dr.MaxMinutes != nil {
		if *dr.MaxMinutes < 0 {
			errs.Add("filters.durationRange.maxMinutes", "maxMinutes must be a non-negative number")
		}
	}

	// Validate that min <= max if both are provided
	if dr.MinMinutes != nil && dr.MaxMinutes != nil {
		if *dr.MinMinutes > *dr.MaxMinutes {
			errs.Add("filters.durationRange", "minMinutes must be less than or equal to maxMinutes")
		}
	}
}

// isValidTimeFormat validates that a time string is in HH:MM format with valid values.
// Hours must be 00-23, minutes must be 00-59.
func isValidTimeFormat(timeStr string) bool {
	// Check basic format
	if !timePattern.MatchString(timeStr) {
		return false
	}

	// Parse and validate hour and minute values
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%02d:%02d", &hour, &minute)
	if err != nil {
		return false
	}

	// Validate ranges
	if hour < 0 || hour > 23 {
		return false
	}
	if minute < 0 || minute > 59 {
		return false
	}

	return true
}
