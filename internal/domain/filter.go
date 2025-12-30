package domain

import (
	"strings"
	"time"
)

// SortOption defines the available sorting options for flight results.
type SortOption string

// Available sort options.
const (
	// SortByBestValue sorts by the calculated ranking score (default)
	SortByBestValue SortOption = "best"

	// SortByPrice sorts by price ascending (cheapest first)
	SortByPrice SortOption = "price"

	// SortByDuration sorts by flight duration ascending (shortest first)
	SortByDuration SortOption = "duration"

	// SortByDeparture sorts by departure time ascending (earliest first)
	SortByDeparture SortOption = "departure"
)

// IsValid checks if the sort option is a valid value.
func (s SortOption) IsValid() bool {
	switch s {
	case SortByBestValue, SortByPrice, SortByDuration, SortByDeparture:
		return true
	default:
		return false
	}
}

// FilterOptions defines optional filters to apply to flight results.
type FilterOptions struct {
	// MaxPrice filters out flights with price above this amount (in the search currency)
	MaxPrice *float64 `json:"maxPrice,omitempty"`

	// MaxStops filters out flights with more stops than this value
	// 0 = direct flights only, 1 = max 1 stop, etc.
	MaxStops *int `json:"maxStops,omitempty"`

	// Airlines filters to only include flights from these airline codes
	// Empty slice means no filtering by airline
	Airlines []string `json:"airlines,omitempty"`

	// DepartureTimeRange filters flights departing within this time range
	DepartureTimeRange *TimeRange `json:"departureTimeRange,omitempty"`

	// DurationRange filters flights by total duration in minutes
	DurationRange *DurationRange `json:"durationRange,omitempty"`
}

// TimeRange represents a time window for filtering.
type TimeRange struct {
	// Start is the beginning of the time range (inclusive)
	Start time.Time `json:"start"`

	// End is the end of the time range (inclusive)
	End time.Time `json:"end"`
}

// DurationRange represents a duration range filter for flights.
type DurationRange struct {
	// MinMinutes is the minimum acceptable flight duration in minutes (inclusive)
	MinMinutes *int `json:"minMinutes,omitempty"`

	// MaxMinutes is the maximum acceptable flight duration in minutes (inclusive)
	MaxMinutes *int `json:"maxMinutes,omitempty"`
}

// IsValid checks if the duration range is valid.
// Returns false if min > max, or if any values are negative.
func (dr *DurationRange) IsValid() bool {
	if dr == nil {
		return true
	}

	// Check for negative values
	if dr.MinMinutes != nil && *dr.MinMinutes < 0 {
		return false
	}
	if dr.MaxMinutes != nil && *dr.MaxMinutes < 0 {
		return false
	}

	// Check min <= max if both are provided
	if dr.MinMinutes != nil && dr.MaxMinutes != nil {
		if *dr.MinMinutes > *dr.MaxMinutes {
			return false
		}
	}

	return true
}

// Contains checks if a given duration (in minutes) falls within the range.
func (dr *DurationRange) Contains(durationMinutes int) bool {
	if dr == nil {
		return true
	}

	// Check minimum duration
	if dr.MinMinutes != nil && durationMinutes < *dr.MinMinutes {
		return false
	}

	// Check maximum duration
	if dr.MaxMinutes != nil && durationMinutes > *dr.MaxMinutes {
		return false
	}

	return true
}

// Contains checks if a given time falls within the time range.
func (tr *TimeRange) Contains(t time.Time) bool {
	if tr == nil {
		return true
	}
	// Extract just the time portion for comparison
	tHour, tMin := t.Hour(), t.Minute()
	tMinutes := tHour*60 + tMin

	startHour, startMin := tr.Start.Hour(), tr.Start.Minute()
	startMinutes := startHour*60 + startMin

	endHour, endMin := tr.End.Hour(), tr.End.Minute()
	endMinutes := endHour*60 + endMin

	return tMinutes >= startMinutes && tMinutes <= endMinutes
}

// MatchesFlight checks if a flight matches all the filter criteria.
func (f *FilterOptions) MatchesFlight(flight Flight) bool {
	if f == nil {
		return true
	}

	// Check price filter
	if f.MaxPrice != nil && flight.Price.Amount > *f.MaxPrice {
		return false
	}

	// Check stops filter
	if f.MaxStops != nil && flight.Stops > *f.MaxStops {
		return false
	}

	// Check airline filter (case-insensitive matching)
	if len(f.Airlines) > 0 {
		found := false
		flightCode := strings.ToUpper(flight.Airline.Code)
		for _, code := range f.Airlines {
			if strings.ToUpper(code) == flightCode {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check departure time range filter
	if f.DepartureTimeRange != nil && !f.DepartureTimeRange.Contains(flight.Departure.DateTime) {
	// Check duration range filter
	if f.DurationRange != nil && !f.DurationRange.Contains(flight.Duration.TotalMinutes) {
		return false
	}

		return false
	}

	return true
}

// ParseSortOption converts a string to a SortOption.
// Returns SortByBestValue if the string is empty or invalid.
func ParseSortOption(s string) SortOption {
	option := SortOption(s)
	if option.IsValid() {
		return option
	}
	return SortByBestValue
}
