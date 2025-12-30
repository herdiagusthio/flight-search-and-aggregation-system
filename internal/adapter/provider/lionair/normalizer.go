package lionair

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// ProviderName is the unique identifier for the Lion Air provider.
const ProviderName = "lion_air"

// normalize converts a slice of Lion Air flights to domain Flight entities.
func normalize(lionAirFlights []LionAirFlight) []domain.Flight {
	result := make([]domain.Flight, 0, len(lionAirFlights))
	skippedCount := 0

	for _, f := range lionAirFlights {
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
			ProviderName, skippedCount, len(lionAirFlights))
	}

	return result
}

// normalizeFlight converts a single Lion Air flight to a domain Flight entity.
func normalizeFlight(f LionAirFlight) (domain.Flight, error) {
	// Parse departure time with timezone
	departureTime, err := parseDateTimeWithTimezone(f.Schedule.Departure, f.Schedule.DepartureTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}

	// Parse arrival time with timezone
	arrivalTime, err := parseDateTimeWithTimezone(f.Schedule.Arrival, f.Schedule.ArrivalTimezone)
	if err != nil {
		return domain.Flight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	// Parse baggage allowances
	cabinKg := parseBaggageWeight(f.Services.BaggageAllowance.Cabin)
	checkedKg := parseBaggageWeight(f.Services.BaggageAllowance.Hold)

	// Calculate stops
	stops := 0
	if !f.IsDirect {
		stops = f.StopCount
		if stops == 0 && len(f.Layovers) > 0 {
			stops = len(f.Layovers)
		}
	}

	return domain.Flight{
		ID:           f.ID,
		FlightNumber: f.ID,
		Airline: domain.AirlineInfo{
			Code: f.Carrier.IATA,
			Name: f.Carrier.Name,
		},
		Departure: domain.FlightPoint{
			AirportCode: f.Route.From.Code,
			AirportName: f.Route.From.Name,
			DateTime:    departureTime,
			Timezone:    f.Schedule.DepartureTimezone,
		},
		Arrival: domain.FlightPoint{
			AirportCode: f.Route.To.Code,
			AirportName: f.Route.To.Name,
			DateTime:    arrivalTime,
			Timezone:    f.Schedule.ArrivalTimezone,
		},
		Duration: domain.NewDurationInfo(f.FlightTime),
		Price: domain.PriceInfo{
			Amount:   f.Pricing.Total,
			Currency: f.Pricing.Currency,
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   cabinKg,
			CheckedKg: checkedKg,
		},
		Class:    normalizeClass(f.Pricing.FareType),
		Stops:    stops,
		Provider: ProviderName,
	}, nil
}

// parseDateTimeWithTimezone parses a datetime string with a separate timezone.
// The datetime format is "2006-01-02T15:04:05" (ISO 8601 without offset).
func parseDateTimeWithTimezone(datetime, timezone string) (time.Time, error) {
	// Try parsing with T separator (ISO 8601 format)
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, datetime)
	if err != nil {
		// Try with space separator as fallback
		layout = "2006-01-02 15:04:05"
		t, err = time.Parse(layout, datetime)
		if err != nil {
			return time.Time{}, fmt.Errorf("unable to parse datetime %q", datetime)
		}
	}

	// Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		return t.UTC(), nil
	}

	// Create time in the specified timezone
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond(),
		loc,
	), nil
}

// parseBaggageWeight extracts the weight in kg from a baggage string like "7 kg".
func parseBaggageWeight(baggageStr string) int {
	// Remove "kg" suffix and trim spaces
	cleaned := strings.TrimSpace(strings.ToLower(baggageStr))
	cleaned = strings.TrimSuffix(cleaned, "kg")
	cleaned = strings.TrimSpace(cleaned)

	weight, err := strconv.Atoi(cleaned)
	if err != nil {
		return 0
	}
	return weight
}

// normalizeClass normalizes the class string to lowercase standard values.
func normalizeClass(class string) string {
	normalized := strings.ToLower(strings.TrimSpace(class))

	switch normalized {
	case "economy", "eco", "y", "economy_class":
		return "economy"
	case "business", "biz", "j", "c", "business_class":
		return "business"
	case "first", "f", "first_class":
		return "first"
	default:
		return "economy" // Default to economy if unknown
	}
}
