package airasia

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// ProviderName is the unique identifier for the AirAsia provider.
const ProviderName = "airasia"

// AirAsia IATA code
const airlineCode = "QZ"

// normalize converts a slice of AirAsiaFlight to domain.Flight entities.
// It skips flights with invalid data (e.g., unparseable datetime).
func normalize(flights []AirAsiaFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(flights))
	skippedCount := 0

	for _, f := range flights {
		normalized, ok := normalizeSingle(f)
		if !ok {
			// Skip flights with invalid data
			// TODO: Add structured logging when logger is available
			skippedCount++
			continue
		}

		// Validate the normalized flight
		if err := normalized.Validate(); err != nil {
			// Log validation error with flight details
			// TODO: Replace with structured logging (WARN level)
			fmt.Printf("[WARN] [%s] Flight %s validation failed: %v\n",
				ProviderName, normalized.FlightNumber, err)
			skippedCount++
			continue
		}

		result = append(result, normalized)
	}

	// Log summary if any flights were skipped
	if skippedCount > 0 {
		// TODO: Replace with structured logging (INFO level)
		fmt.Printf("[INFO] [%s] Skipped %d invalid flights out of %d total\n",
			ProviderName, skippedCount, len(flights))
	}

	return result
}

// normalizeSingle converts a single AirAsiaFlight to a domain.Flight.
// Returns false if the flight cannot be normalized (e.g., invalid datetime).
func normalizeSingle(f AirAsiaFlight) (domain.Flight, bool) {
	// Parse departure time
	departureTime, err := parseDateTime(f.DepartTime)
	if err != nil {
		return domain.Flight{}, false
	}

	// Parse arrival time
	arrivalTime, err := parseDateTime(f.ArriveTime)
	if err != nil {
		return domain.Flight{}, false
	}

	// Calculate stops count
	stopsCount := directFlightToStops(f.DirectFlight, f.Stops)

	// Generate a unique ID for this flight
	flightID := generateFlightID(f)

	// Parse baggage from note
	cabinKg, checkedKg := parseBaggageNote(f.BaggageNote)

	return domain.Flight{
		ID:           flightID,
		FlightNumber: f.FlightCode,
		Airline: domain.AirlineInfo{
			Code: airlineCode, // AirAsia Indonesia uses QZ
			Name: f.Airline,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.FromAirport,
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.ToAirport,
			DateTime:    arrivalTime,
		},
		Duration: domain.DurationInfo{
			TotalMinutes: hoursToMinutes(f.DurationHours),
			Formatted:    formatDurationFromHours(f.DurationHours),
		},
		Price: domain.PriceInfo{
			Amount:   f.PriceIDR,
			Currency: "IDR", // AirAsia Indonesia prices are in IDR
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		Class:    strings.ToLower(f.CabinClass),
		Stops:    stopsCount,
		Provider: ProviderName,
	}, true
}

// generateFlightID creates a unique identifier for a flight.
func generateFlightID(f AirAsiaFlight) string {
	return fmt.Sprintf("%s-%s-%s-%s", ProviderName, f.FlightCode, f.FromAirport, f.ToAirport)
}

// hoursToMinutes converts float hours to integer minutes with proper rounding.
// Examples: 1.75 → 105, 2.5 → 150, 0.5 → 30
func hoursToMinutes(hours float64) int {
	return int(math.Round(hours * 60))
}

// formatDurationFromHours formats a float duration as "Xh Ym".
// Examples: 1.75 → "1h 45m", 2.5 → "2h 30m", 0.5 → "30m"
func formatDurationFromHours(hours float64) string {
	totalMinutes := hoursToMinutes(hours)
	h := totalMinutes / 60
	m := totalMinutes % 60

	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

// directFlightToStops converts the direct_flight boolean to stops count.
// If direct_flight is true, returns 0.
// If direct_flight is false, returns the actual number of stops or 1 if unknown.
func directFlightToStops(isDirect bool, stops []AirAsiaStop) int {
	if isDirect {
		return 0
	}
	// If stops array is provided, use its length
	if len(stops) > 0 {
		return len(stops)
	}
	// Default to 1 stop if not direct but no stops array
	return 1
}

// parseDateTime parses an ISO 8601 datetime string to time.Time.
// Supports formats with timezone offset (e.g., "2025-12-15T06:00:00+07:00").
func parseDateTime(datetime string) (time.Time, error) {
	// Try standard RFC3339 format first
	t, err := time.Parse(time.RFC3339, datetime)
	if err == nil {
		return t, nil
	}

	// Try without colon in timezone (e.g., +0700)
	t, err = time.Parse("2006-01-02T15:04:05-0700", datetime)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", datetime)
}

// parseBaggageNote extracts baggage weights from a baggage note string.
// AirAsia typically provides a note like "Cabin baggage only, checked bags additional fee"
// which means only cabin baggage is included (default 7kg), no checked baggage.
func parseBaggageNote(note string) (cabinKg, checkedKg int) {
	noteLower := strings.ToLower(note)

	// Default cabin baggage for AirAsia is 7kg
	cabinKg = 7

	// Check if checked baggage is included
	if strings.Contains(noteLower, "checked bag") && strings.Contains(noteLower, "additional fee") {
		// No checked baggage included
		checkedKg = 0
	} else if strings.Contains(noteLower, "20kg") {
		checkedKg = 20
	} else if strings.Contains(noteLower, "15kg") {
		checkedKg = 15
	} else {
		// Default to no checked baggage for low-cost carrier
		checkedKg = 0
	}

	return cabinKg, checkedKg
}
