// Package usecase provides the business logic for flight search operations.
package usecase

import (
	"strings"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// ApplyFilters applies the given filter options to a list of flights.
// It returns a new slice containing only flights that match all filter criteria.
//
// Behavior:
//   - Returns the original slice if opts is nil (no filtering)
//   - Filters are applied in sequence (price -> stops -> airlines -> departure time)
//   - Nil/empty filter values are skipped (no filtering on that criterion)
//   - Does NOT mutate the original flights slice
//   - Performance is O(n) where n = number of flights
//
// Example usage:
//
//	maxPrice := float64(1000000)
//	opts := &domain.FilterOptions{MaxPrice: &maxPrice}
//	filtered := ApplyFilters(flights, opts)
func ApplyFilters(flights []domain.Flight, opts *domain.FilterOptions) []domain.Flight {
	if opts == nil {
		return flights
	}

	// Pre-build airline set for O(1) lookup if airlines filter is provided
	var airlineSet map[string]struct{}
	if len(opts.Airlines) > 0 {
		airlineSet = buildAirlineSet(opts.Airlines)
	}

	// Pre-allocate with estimated capacity
	result := make([]domain.Flight, 0, len(flights))

	for _, f := range flights {
		if passesAllFilters(f, opts, airlineSet) {
			result = append(result, f)
		}
	}

	return result
}

// passesAllFilters checks if a flight passes all filter criteria.
// This is an internal function that performs the actual filter checks.
func passesAllFilters(f domain.Flight, opts *domain.FilterOptions, airlineSet map[string]struct{}) bool {
	// Price filter: include flights where price <= maxPrice
	if opts.MaxPrice != nil && f.Price.Amount > *opts.MaxPrice {
		return false
	}

	// Stops filter: include flights where stops <= maxStops
	if opts.MaxStops != nil && f.Stops > *opts.MaxStops {
		return false
	}

	// Airlines filter: include flights where airline code is in whitelist
	if len(opts.Airlines) > 0 && !isAirlineInSet(f.Airline.Code, airlineSet) {
		return false
	}

	// Departure time range filter: include flights departing within the range
	if opts.DepartureTimeRange != nil && !opts.DepartureTimeRange.Contains(f.Departure.DateTime) {
		return false
	}

	// Duration range filter: include flights with duration within the range
	if opts.DurationRange != nil && !opts.DurationRange.Contains(f.Duration.TotalMinutes) {
		return false
	}

	return true
}

// buildAirlineSet creates a case-insensitive lookup set from a list of airline codes.
// This provides O(1) lookup performance for the airlines filter.
func buildAirlineSet(airlines []string) map[string]struct{} {
	set := make(map[string]struct{}, len(airlines))
	for _, code := range airlines {
		// Store uppercase for case-insensitive matching
		set[strings.ToUpper(code)] = struct{}{}
	}
	return set
}

// isAirlineInSet checks if an airline code is in the allowed set (case-insensitive).
func isAirlineInSet(code string, set map[string]struct{}) bool {
	_, exists := set[strings.ToUpper(code)]
	return exists
}

// FilterByMaxPrice filters flights by maximum price.
// Returns all flights if maxPrice is nil.
// This is an exported helper for direct filter application.
func FilterByMaxPrice(flights []domain.Flight, maxPrice *float64) []domain.Flight {
	if maxPrice == nil {
		return flights
	}

	result := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if f.Price.Amount <= *maxPrice {
			result = append(result, f)
		}
	}
	return result
}

// FilterByMaxStops filters flights by maximum number of stops.
// Returns all flights if maxStops is nil.
// Common values: 0 (direct only), 1 (max 1 stop), 2 (max 2 stops)
func FilterByMaxStops(flights []domain.Flight, maxStops *int) []domain.Flight {
	if maxStops == nil {
		return flights
	}

	result := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if f.Stops <= *maxStops {
			result = append(result, f)
		}
	}
	return result
}

// FilterByAirlines filters flights to only include those from specified airlines.
// Returns all flights if airlines slice is nil or empty.
// Matching is case-insensitive.
func FilterByAirlines(flights []domain.Flight, airlines []string) []domain.Flight {
	if len(airlines) == 0 {
		return flights
	}

	airlineSet := buildAirlineSet(airlines)
	result := make([]domain.Flight, 0, len(flights))

	for _, f := range flights {
		if isAirlineInSet(f.Airline.Code, airlineSet) {
			result = append(result, f)
		}
	}
	return result
}

// FilterByDepartureTime filters flights by departure time range.
// Returns all flights if timeRange is nil.
func FilterByDepartureTime(flights []domain.Flight, timeRange *domain.TimeRange) []domain.Flight {
	if timeRange == nil {
		return flights
	}

	result := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if timeRange.Contains(f.Departure.DateTime) {
			result = append(result, f)
		}
	}
	return result
}

// FilterByDuration filters flights by duration range in minutes.
// Returns all flights if durationRange is nil.
func FilterByDuration(flights []domain.Flight, durationRange *domain.DurationRange) []domain.Flight {
	if durationRange == nil {
		return flights
	}

	result := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if durationRange.Contains(f.Duration.TotalMinutes) {
			result = append(result, f)
		}
	}
	return result
}
