package batikair

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// ProviderName is the unique identifier for the Batik Air provider.
const ProviderName = "batik_air"

// durationRegex matches duration strings like "2h 15m", "1h", "45m"
var durationRegex = regexp.MustCompile(`(?:(\d+)h)?\s*(?:(\d+)m)?`)

// normalize converts a slice of Batik Air flights to domain Flight entities.
func normalize(batikAirFlights []BatikAirFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(batikAirFlights))
	skippedCount := 0

	for _, f := range batikAirFlights {
		normalized, err := normalizeFlight(f)
		if err != nil {
			// Skip flights that cannot be normalized
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
			ProviderName, skippedCount, len(batikAirFlights))
	}

	return result
}

// normalizeFlight converts a single Batik Air flight to a domain Flight entity.
func normalizeFlight(f BatikAirFlight) (domain.Flight, error) {
	// Parse departure time
	departureTime, err := parseDateTime(f.DepartureDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := parseDateTime(f.ArrivalDateTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	// Parse duration from travel time string
	durationMinutes, err := parseDurationString(f.TravelTime)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse travel time: %w", err)
	}

	// Parse baggage info
	cabinKg, checkedKg := parseBaggageInfo(f.BaggageInfo)

	// Use totalPrice if available, otherwise calculate from base + taxes
	totalPrice := f.Fare.TotalPrice
	if totalPrice == 0 {
		totalPrice = f.Fare.BasePrice + f.Fare.Taxes
	}

	return domain.Flight{
		ID:           f.FlightNumber,
		FlightNumber: f.FlightNumber,
		Airline: domain.AirlineInfo{
			Code: f.AirlineIATA,
			Name: f.AirlineName,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.Origin,
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.Destination,
			DateTime:    arrivalTime,
		},
		Duration: domain.DurationInfo{
			TotalMinutes: durationMinutes,
			Formatted:    f.TravelTime,
		},
		Price: domain.PriceInfo{
			Amount:   totalPrice,
			Currency: f.Fare.CurrencyCode,
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		Class:    mapCabinClass(f.Fare.Class),
		Stops:    f.NumberOfStops,
		Provider: ProviderName,
	}, nil
}

// parseDateTime parses an ISO 8601 datetime string to time.Time.
// Supports formats: "2006-01-02T15:04:05+0700" and "2006-01-02T15:04:05Z07:00"
func parseDateTime(datetime string) (time.Time, error) {
	// Try RFC3339 format first (with colon in timezone)
	t, err := time.Parse(time.RFC3339, datetime)
	if err == nil {
		return t, nil
	}

	// Try without colon in timezone offset (e.g., +0700)
	t, err = time.Parse("2006-01-02T15:04:05-0700", datetime)
	if err == nil {
		return t, nil
	}

	// Try without timezone
	t, err = time.Parse("2006-01-02T15:04:05", datetime)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime %q", datetime)
}

// parseDurationString parses a duration string like "2h 15m" to total minutes.
// Handles formats: "2h 15m", "1h", "45m", "0h 30m"
func parseDurationString(duration string) (int, error) {
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	matches := durationRegex.FindStringSubmatch(duration)
	if matches == nil || (matches[1] == "" && matches[2] == "") {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	var hours, minutes int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}

	return hours*60 + minutes, nil
}

// parseBaggageInfo extracts cabin and checked baggage weights from a string.
// Example: "7kg cabin, 20kg checked" -> 7, 20
func parseBaggageInfo(baggageInfo string) (cabinKg, checkedKg int) {
	// Default values
	cabinKg = 7
	checkedKg = 20

	if baggageInfo == "" {
		return
	}

	info := strings.ToLower(baggageInfo)

	// Try to extract cabin baggage
	cabinRegex := regexp.MustCompile(`(\d+)\s*kg\s*cabin`)
	if matches := cabinRegex.FindStringSubmatch(info); len(matches) > 1 {
		cabinKg, _ = strconv.Atoi(matches[1])
	}

	// Try to extract checked baggage
	checkedRegex := regexp.MustCompile(`(\d+)\s*kg\s*checked`)
	if matches := checkedRegex.FindStringSubmatch(info); len(matches) > 1 {
		checkedKg, _ = strconv.Atoi(matches[1])
	}

	return
}

// mapCabinClass maps airline cabin class codes to standard class names.
func mapCabinClass(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))

	classMap := map[string]string{
		"Y": "economy",
		"W": "premium_economy",
		"C": "business",
		"J": "business",
		"F": "first",
	}

	if class, ok := classMap[code]; ok {
		return class
	}
	return "economy" // Default
}
