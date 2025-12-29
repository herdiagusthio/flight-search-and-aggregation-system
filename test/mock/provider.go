// Package mock provides test doubles for the flight search system.
// These mocks are designed for integration testing where we need
// configurable behavior (delays, errors, specific responses).
package mock

import (
	"context"
	"sync"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// Provider is a configurable mock implementation of domain.FlightProvider.
// It supports configurable delays, errors, and responses for testing
// various scenarios including timeouts and partial failures.
type Provider struct {
	name      string
	flights   []domain.Flight
	err       error
	delay     time.Duration
	callCount int
	mu        sync.Mutex
}

// NewProvider creates a new mock provider with the given name.
// The provider is configured using the builder pattern methods.
func NewProvider(name string) *Provider {
	return &Provider{
		name:    name,
		flights: nil,
		err:     nil,
		delay:   0,
	}
}

// WithFlights configures the provider to return the given flights.
func (p *Provider) WithFlights(flights []domain.Flight) *Provider {
	p.flights = flights
	return p
}

// WithError configures the provider to return the given error.
func (p *Provider) WithError(err error) *Provider {
	p.err = err
	return p
}

// WithDelay configures the provider to wait the given duration before responding.
// This is useful for testing timeout behavior.
func (p *Provider) WithDelay(d time.Duration) *Provider {
	p.delay = d
	return p
}

// Name returns the provider's unique identifier.
func (p *Provider) Name() string {
	return p.name
}

// Search implements domain.FlightProvider.Search.
// It respects context cancellation, applies configured delay,
// and returns configured flights or error.
func (p *Provider) Search(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
	p.mu.Lock()
	p.callCount++
	p.mu.Unlock()

	// Apply delay if configured
	if p.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(p.delay):
		}
	}

	// Check context after delay
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Return configured error if set
	if p.err != nil {
		return nil, p.err
	}

	// Return configured flights
	return p.flights, nil
}

// CallCount returns the number of times Search was called.
// This is useful for verifying provider interactions.
func (p *Provider) CallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.callCount
}

// Reset resets the call count to zero.
func (p *Provider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.callCount = 0
}

// Ensure Provider implements domain.FlightProvider at compile time.
var _ domain.FlightProvider = (*Provider)(nil)

// SampleFlights returns a slice of sample flights for testing.
// The flights have all required fields populated with realistic values.
func SampleFlights(provider string, count int) []domain.Flight {
	flights := make([]domain.Flight, count)

	baseTime := time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		departureTime := baseTime.Add(time.Duration(i*2) * time.Hour)
		arrivalTime := departureTime.Add(2*time.Hour + 30*time.Minute)

		flights[i] = domain.Flight{
			ID:           generateFlightID(provider, i),
			FlightNumber: generateFlightNumber(provider, i),
			Airline: domain.AirlineInfo{
				Code: providerToAirlineCode(provider),
				Name: providerToAirlineName(provider),
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				AirportName: "Soekarno-Hatta International Airport",
				Terminal:    "3",
				DateTime:    departureTime,
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				AirportName: "Ngurah Rai International Airport",
				Terminal:    "D",
				DateTime:    arrivalTime,
			},
			Duration: domain.NewDurationInfo(150), // 2h 30m
			Price: domain.PriceInfo{
				Amount:   1000000 + float64(i*100000),
				Currency: "IDR",
			},
			Baggage: domain.BaggageInfo{
				CabinKg:   7,
				CheckedKg: 20,
			},
			Class:    "economy",
			Stops:    0,
			Provider: provider,
		}
	}

	return flights
}

// generateFlightID creates a unique flight ID.
func generateFlightID(provider string, index int) string {
	return provider + "-" + intToString(index+1)
}

// generateFlightNumber creates a realistic flight number.
func generateFlightNumber(provider string, index int) string {
	code := providerToAirlineCode(provider)
	return code + " " + intToString(100+index)
}

// providerToAirlineCode maps provider names to IATA codes.
func providerToAirlineCode(provider string) string {
	codes := map[string]string{
		"garuda":  "GA",
		"lion":    "JT",
		"batik":   "ID",
		"airasia": "QZ",
	}
	if code, ok := codes[provider]; ok {
		return code
	}
	return "XX"
}

// providerToAirlineName maps provider names to full names.
func providerToAirlineName(provider string) string {
	names := map[string]string{
		"garuda":  "Garuda Indonesia",
		"lion":    "Lion Air",
		"batik":   "Batik Air",
		"airasia": "AirAsia",
	}
	if name, ok := names[provider]; ok {
		return name
	}
	return "Unknown Airline"
}

// intToString converts an integer to string without importing strconv.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
