package domain

import (
	"fmt"
	"regexp"
	"time"
)

// SearchCriteria defines the parameters for a flight search request.
type SearchCriteria struct {
	// Origin is the IATA code of the departure airport (e.g., "CGK")
	Origin string `json:"origin"`

	// Destination is the IATA code of the arrival airport (e.g., "DPS")
	Destination string `json:"destination"`

	// DepartureDate is the desired departure date in YYYY-MM-DD format
	DepartureDate string `json:"departureDate"`

	// Passengers is the number of passengers (default: 1)
	Passengers int `json:"passengers"`

	// Class is the travel class: economy, business, or first (default: economy)
	Class string `json:"class,omitempty"`
}

// airportCodeRegex matches valid IATA airport codes (3 uppercase letters).
var airportCodeRegex = regexp.MustCompile(`^[A-Z]{3}$`)

// dateRegex matches dates in YYYY-MM-DD format.
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// validClasses defines the allowed travel classes.
var validClasses = map[string]bool{
	"economy":  true,
	"business": true,
	"first":    true,
}

// Validate checks if the search criteria is valid.
// Returns a wrapped ErrInvalidRequest error if validation fails.
func (s *SearchCriteria) Validate() error {
	// Validate origin
	if s.Origin == "" {
		return fmt.Errorf("%w: origin is required", ErrInvalidRequest)
	}
	if !airportCodeRegex.MatchString(s.Origin) {
		return fmt.Errorf("%w: origin must be a valid 3-letter IATA code, got %q", ErrInvalidRequest, s.Origin)
	}

	// Validate destination
	if s.Destination == "" {
		return fmt.Errorf("%w: destination is required", ErrInvalidRequest)
	}
	if !airportCodeRegex.MatchString(s.Destination) {
		return fmt.Errorf("%w: destination must be a valid 3-letter IATA code, got %q", ErrInvalidRequest, s.Destination)
	}

	// Origin and destination must be different
	if s.Origin == s.Destination {
		return fmt.Errorf("%w: origin and destination must be different", ErrInvalidRequest)
	}

	// Validate departure date
	if s.DepartureDate == "" {
		return fmt.Errorf("%w: departureDate is required", ErrInvalidRequest)
	}
	if !dateRegex.MatchString(s.DepartureDate) {
		return fmt.Errorf("%w: departureDate must be in YYYY-MM-DD format, got %q", ErrInvalidRequest, s.DepartureDate)
	}

	// Parse and validate the date is valid
	_, err := time.Parse("2006-01-02", s.DepartureDate)
	if err != nil {
		return fmt.Errorf("%w: departureDate is not a valid date: %s", ErrInvalidRequest, s.DepartureDate)
	}

	// Validate passengers
	if s.Passengers < 1 {
		return fmt.Errorf("%w: passengers must be at least 1", ErrInvalidRequest)
	}
	if s.Passengers > 9 {
		return fmt.Errorf("%w: passengers cannot exceed 9", ErrInvalidRequest)
	}

	// Validate class (if provided)
	if s.Class != "" && !validClasses[s.Class] {
		return fmt.Errorf("%w: class must be one of: economy, business, first; got %q", ErrInvalidRequest, s.Class)
	}

	return nil
}

// SetDefaults applies default values to empty optional fields.
func (s *SearchCriteria) SetDefaults() {
	if s.Passengers == 0 {
		s.Passengers = 1
	}
	if s.Class == "" {
		s.Class = "economy"
	}
}
