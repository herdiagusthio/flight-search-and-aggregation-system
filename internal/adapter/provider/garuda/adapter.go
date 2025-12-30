package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// Adapter implements the domain.FlightProvider interface for Garuda Indonesia.
// It reads from mock JSON data and normalizes it to the unified Flight domain model.
type Adapter struct {
	// mockDataPath is the path to the mock JSON data file.
	mockDataPath string
	// skipSimulation disables delay simulation for deterministic testing.
	skipSimulation bool
}

// NewAdapter creates a new Garuda Indonesia adapter.
// The mockDataPath parameter specifies the path to the mock JSON data file.
func NewAdapter(mockDataPath string) *Adapter {
	return &Adapter{
		mockDataPath:   mockDataPath,
		skipSimulation: true, // Default to skipping simulation for tests
	}
}

// NewAdapterWithSimulation creates a new Garuda Indonesia adapter with real-world simulation enabled.
// Use this for production to simulate realistic API behavior.
func NewAdapterWithSimulation(mockDataPath string) *Adapter {
	return &Adapter{
		mockDataPath:   mockDataPath,
		skipSimulation: false,
	}
}

// Name returns the unique identifier for this provider.
// Implements domain.FlightProvider.
func (a *Adapter) Name() string {
	return ProviderName
}

// Search queries the provider for available flights matching the criteria.
// It reads from mock JSON data and returns normalized flight entities.
// Simulates real-world conditions: Fast response (50-100ms delay).
// Implements domain.FlightProvider.
func (a *Adapter) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	// Only simulate if not in test mode
	if !a.skipSimulation {
		// Simulate network latency: 50-100ms
		delay := time.Duration(50+rand.Intn(51)) * time.Millisecond
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Continue after delay
		case <-ctx.Done():
			return nil, &domain.ProviderError{
				Provider:  ProviderName,
				Err:       ctx.Err(),
				Retryable: false,
			}
		}
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       ctx.Err(),
			Retryable: false,
		}
	default:
	}

	// Read mock data file
	data, err := os.ReadFile(a.mockDataPath)
	if err != nil {
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       fmt.Errorf("failed to read mock data: %w", err),
			Retryable: true, // File read errors might be temporary
		}
	}

	// Parse JSON
	var response GarudaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, &domain.ProviderError{
			Provider:  ProviderName,
			Err:       fmt.Errorf("failed to parse JSON: %w", err),
			Retryable: false, // Parse errors are not retryable
		}
	}

	// Check for empty flights array
	if len(response.Flights) == 0 {
		return []domain.Flight{}, nil
	}

	// Normalize flights to domain model
	flights := normalize(response.Flights)

	// Filter flights based on criteria (origin, destination, date)
	filtered := filterFlights(flights, criteria)

	return filtered, nil
}

// filterFlights filters normalized flights based on the search criteria.
func filterFlights(flights []domain.Flight, criteria domain.SearchCriteria) []domain.Flight {
	result := make([]domain.Flight, 0, len(flights))

	for _, f := range flights {
		// Filter by origin if specified
		if criteria.Origin != "" && f.Departure.AirportCode != criteria.Origin {
			continue
		}

		// Filter by destination if specified
		if criteria.Destination != "" && f.Arrival.AirportCode != criteria.Destination {
			continue
		}

		// Filter by departure date if specified
		if criteria.DepartureDate != "" {
			flightDate := f.Departure.DateTime.Format("2006-01-02")
			if flightDate != criteria.DepartureDate {
				continue
			}
		}

		// Filter by class if specified
		if criteria.Class != "" && f.Class != criteria.Class {
			continue
		}

		result = append(result, f)
	}

	return result
}

// Ensure Adapter implements FlightProvider at compile time.
var _ domain.FlightProvider = (*Adapter)(nil)
