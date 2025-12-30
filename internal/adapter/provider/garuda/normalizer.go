package garuda

import (
	"fmt"
	"strings"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// ProviderName is the unique identifier for the Garuda Indonesia provider.
const ProviderName = "garuda_indonesia"

// DefaultCabinBaggageKg is the default cabin baggage weight per piece.
const DefaultCabinBaggageKg = 7

// DefaultCheckedBaggageKg is the default checked baggage weight per piece.
const DefaultCheckedBaggageKg = 20

// normalize converts a slice of Garuda flights to domain Flight entities.
func normalize(garudaFlights []GarudaFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(garudaFlights))
	skippedCount := 0

	for _, f := range garudaFlights {
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
			ProviderName, skippedCount, len(garudaFlights))
	}

	return result
}

// normalizeFlight converts a single Garuda flight to a domain Flight entity.
func normalizeFlight(f GarudaFlight) (domain.Flight, error) {
	// Parse departure time
	departureTime, err := parseDateTime(f.Departure.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}

	// Parse arrival time
	arrivalTime, err := parseDateTime(f.Arrival.Time)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	// Calculate stops from segments if available, otherwise use stops field
	stops := f.Stops
	if len(f.Segments) > 1 {
		stops = len(f.Segments) - 1
	}

	return domain.Flight{
		ID:           f.FlightID,
		FlightNumber: f.FlightID, // Use flight_id as flight number since it contains the flight identifier
		Airline: domain.AirlineInfo{
			Code: f.AirlineCode,
			Name: f.Airline,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.Departure.Airport,
			AirportName: formatAirportName(f.Departure.Airport, f.Departure.City),
			Terminal:    f.Departure.Terminal,
			DateTime:    departureTime,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.Arrival.Airport,
			AirportName: formatAirportName(f.Arrival.Airport, f.Arrival.City),
			Terminal:    f.Arrival.Terminal,
			DateTime:    arrivalTime,
		},
		Duration: domain.NewDurationInfo(f.DurationMinutes),
		Price: domain.PriceInfo{
			Amount:   f.Price.Amount,
			Currency: f.Price.Currency,
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   f.Baggage.CarryOn * DefaultCabinBaggageKg,
			CheckedKg: f.Baggage.Checked * DefaultCheckedBaggageKg,
		},
		Class:    normalizeClass(f.FareClass),
		Stops:    stops,
		Provider: ProviderName,
	}, nil
}

// parseDateTime parses an ISO 8601 datetime string to time.Time.
// Supports formats: "2006-01-02T15:04:05Z07:00" and "2006-01-02T15:04:05"
func parseDateTime(dateTime string) (time.Time, error) {
	// Try RFC3339 format first (with timezone)
	t, err := time.Parse(time.RFC3339, dateTime)
	if err == nil {
		return t, nil
	}

	// Try without timezone
	t, err = time.Parse("2006-01-02T15:04:05", dateTime)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime %q", dateTime)
}

// formatAirportName creates a formatted airport name from code and city.
func formatAirportName(code, city string) string {
	if city == "" {
		return code
	}
	return fmt.Sprintf("%s (%s)", city, code)
}

// normalizeClass normalizes the class string to lowercase standard values.
func normalizeClass(class string) string {
	normalized := strings.ToLower(strings.TrimSpace(class))

	switch normalized {
	case "economy", "eco", "y":
		return "economy"
	case "business", "biz", "j", "c":
		return "business"
	case "first", "f":
		return "first"
	default:
		return "economy" // Default to economy if unknown
	}
}
