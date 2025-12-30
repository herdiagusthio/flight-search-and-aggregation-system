package usecase

import (
	"testing"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createFilterTestFlight creates a flight for filter testing.
func createFilterTestFlight(id string, price float64, stops int, airlineCode string, departureHour int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: "FL-" + id,
		Airline: domain.AirlineInfo{
			Code: airlineCode,
			Name: "Test Airline",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			DateTime:    time.Date(2025, 12, 15, departureHour, 0, 0, 0, time.UTC),
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			DateTime:    time.Date(2025, 12, 15, departureHour+2, 0, 0, 0, time.UTC),
		},
		Duration: domain.DurationInfo{
			TotalMinutes: 120,
			Formatted:    "2h 0m",
		},
		Price: domain.PriceInfo{
			Amount:   price,
			Currency: "IDR",
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   7,
			CheckedKg: 20,
		},
		Class:    "economy",
		Stops:    stops,
		Provider: "test",
	}
}

// createFlightWithArrival creates a flight with a specific arrival time for testing.
func createFlightWithArrival(id string, price float64, stops int, airlineCode string, arrivalHour int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: "FL-" + id,
		Airline: domain.AirlineInfo{
			Code: airlineCode,
			Name: "Test Airline",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			DateTime:    time.Date(2025, 12, 15, arrivalHour-2, 0, 0, 0, time.UTC),
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			DateTime:    time.Date(2025, 12, 15, arrivalHour, 0, 0, 0, time.UTC),
		},
		Duration: domain.DurationInfo{
			TotalMinutes: 120,
			Formatted:    "2h 0m",
		},
		Price: domain.PriceInfo{
			Amount:   price,
			Currency: "IDR",
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   7,
			CheckedKg: 20,
		},
		Class:    "economy",
		Stops:    stops,
		Provider: "test",
	}
}

// Helper functions for creating pointer values
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int           { return &i }

// TestApplyFilters_NilOptions tests that nil options return all flights.
func TestApplyFilters_NilOptions(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 1500000, 1, "JT", 10),
		createFilterTestFlight("3", 800000, 2, "ID", 14),
	}

	result := ApplyFilters(flights, nil)

	assert.Len(t, result, 3)
	assert.Equal(t, flights, result)
}

// TestApplyFilters_EmptyOptions tests that empty options return all flights.
func TestApplyFilters_EmptyOptions(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 1500000, 1, "JT", 10),
	}

	result := ApplyFilters(flights, &domain.FilterOptions{})

	assert.Len(t, result, 2)
}

// TestApplyFilters_EmptyFlightList tests filtering an empty list.
func TestApplyFilters_EmptyFlightList(t *testing.T) {
	flights := []domain.Flight{}
	maxPrice := float64(1000000)
	opts := &domain.FilterOptions{MaxPrice: &maxPrice}

	result := ApplyFilters(flights, opts)

	assert.Empty(t, result)
	assert.NotNil(t, result)
}

// TestApplyFilters_NoMatches tests when no flights match the criteria.
func TestApplyFilters_NoMatches(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 2000000, 2, "GA", 8),
		createFilterTestFlight("2", 2500000, 3, "JT", 10),
	}

	maxPrice := float64(1000000)
	opts := &domain.FilterOptions{MaxPrice: &maxPrice}

	result := ApplyFilters(flights, opts)

	assert.Empty(t, result)
}

// TestApplyFilters_DoesNotMutateOriginal tests that the original slice is not modified.
func TestApplyFilters_DoesNotMutateOriginal(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 500000, 0, "GA", 8),
		createFilterTestFlight("2", 1500000, 1, "JT", 10),
		createFilterTestFlight("3", 800000, 2, "ID", 14),
	}

	originalLen := len(flights)
	originalFirst := flights[0].ID

	maxPrice := float64(1000000)
	opts := &domain.FilterOptions{MaxPrice: &maxPrice}

	result := ApplyFilters(flights, opts)

	// Original slice should be unchanged
	assert.Len(t, flights, originalLen)
	assert.Equal(t, originalFirst, flights[0].ID)

	// Result should be a new slice
	assert.Len(t, result, 2)
}

// TestApplyFilters_MaxPrice tests price filtering with various thresholds.
func TestApplyFilters_MaxPrice(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 500000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 0, "JT", 10),
		createFilterTestFlight("3", 1500000, 0, "ID", 14),
		createFilterTestFlight("4", 2000000, 0, "AK", 16),
	}

	tests := []struct {
		name     string
		maxPrice float64
		expected int
	}{
		{"very low threshold", 100000, 0},
		{"low threshold", 500000, 1},
		{"medium threshold", 1000000, 2},
		{"high threshold", 1500000, 3},
		{"very high threshold", 3000000, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &domain.FilterOptions{MaxPrice: &tt.maxPrice}
			result := ApplyFilters(flights, opts)
			assert.Len(t, result, tt.expected)
		})
	}
}

// TestApplyFilters_MaxPriceExact tests exact price boundary.
func TestApplyFilters_MaxPriceExact(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
	}

	maxPrice := float64(1000000) // Exact match
	opts := &domain.FilterOptions{MaxPrice: &maxPrice}

	result := ApplyFilters(flights, opts)

	assert.Len(t, result, 1, "Flight at exact price boundary should be included")
}

// TestApplyFilters_MaxStops tests stops filtering.
func TestApplyFilters_MaxStops(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("direct", 1000000, 0, "GA", 8),
		createFilterTestFlight("one-stop", 900000, 1, "JT", 10),
		createFilterTestFlight("two-stops", 800000, 2, "ID", 14),
		createFilterTestFlight("three-stops", 700000, 3, "AK", 16),
	}

	tests := []struct {
		name     string
		maxStops int
		expected int
		desc     string
	}{
		{"direct only", 0, 1, "Only direct flights"},
		{"max 1 stop", 1, 2, "Direct + 1 stop"},
		{"max 2 stops", 2, 3, "Direct + 1 + 2 stops"},
		{"max 3 stops", 3, 4, "All flights"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &domain.FilterOptions{MaxStops: &tt.maxStops}
			result := ApplyFilters(flights, opts)
			assert.Len(t, result, tt.expected, tt.desc)
		})
	}
}

// TestApplyFilters_AirlinesSingle tests filtering by a single airline.
func TestApplyFilters_AirlinesSingle(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 900000, 0, "JT", 10),
		createFilterTestFlight("3", 800000, 0, "ID", 14),
	}

	opts := &domain.FilterOptions{Airlines: []string{"GA"}}
	result := ApplyFilters(flights, opts)

	require.Len(t, result, 1)
	assert.Equal(t, "GA", result[0].Airline.Code)
}

// TestApplyFilters_AirlinesMultiple tests filtering by multiple airlines.
func TestApplyFilters_AirlinesMultiple(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 900000, 0, "JT", 10),
		createFilterTestFlight("3", 800000, 0, "ID", 14),
		createFilterTestFlight("4", 700000, 0, "AK", 16),
	}

	opts := &domain.FilterOptions{Airlines: []string{"GA", "JT", "ID"}}
	result := ApplyFilters(flights, opts)

	assert.Len(t, result, 3)
}

// TestApplyFilters_AirlinesCaseInsensitive tests case-insensitive airline matching.
func TestApplyFilters_AirlinesCaseInsensitive(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 900000, 0, "jt", 10),
		createFilterTestFlight("3", 800000, 0, "Id", 14),
	}

	tests := []struct {
		name     string
		airlines []string
		expected int
	}{
		{"lowercase filter matches uppercase", []string{"ga"}, 1},
		{"uppercase filter matches lowercase", []string{"JT"}, 1},
		{"mixed case filter matches mixed", []string{"iD"}, 1},
		{"mixed case multiple", []string{"ga", "JT", "iD"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &domain.FilterOptions{Airlines: tt.airlines}
			result := ApplyFilters(flights, opts)
			assert.Len(t, result, tt.expected)
		})
	}
}

// TestApplyFilters_AirlinesEmptyList tests that empty airline list returns all.
func TestApplyFilters_AirlinesEmptyList(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 900000, 0, "JT", 10),
	}

	opts := &domain.FilterOptions{Airlines: []string{}}
	result := ApplyFilters(flights, opts)

	assert.Len(t, result, 2)
}

// TestApplyFilters_DepartureTimeRange tests time range filtering.
func TestApplyFilters_DepartureTimeRange(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("morning", 1000000, 0, "GA", 6),
		createFilterTestFlight("mid-morning", 900000, 0, "JT", 9),
		createFilterTestFlight("noon", 800000, 0, "ID", 12),
		createFilterTestFlight("evening", 700000, 0, "AK", 18),
	}

	tests := []struct {
		name      string
		startHour int
		endHour   int
		expected  int
	}{
		{"morning only", 6, 9, 2},
		{"afternoon", 12, 18, 2},
		{"all day", 0, 23, 4},
		{"narrow window", 9, 9, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &domain.FilterOptions{
				DepartureTimeRange: &domain.TimeRange{
					Start: time.Date(2025, 1, 1, tt.startHour, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, tt.endHour, 0, 0, 0, time.UTC),
				},
			}
			result := ApplyFilters(flights, opts)
			assert.Len(t, result, tt.expected)
		})
	}
}

// TestApplyFilters_CombinedFilters tests multiple filters applied together.
func TestApplyFilters_CombinedFilters(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 800000, 0, "GA", 8),   // Passes all
		createFilterTestFlight("2", 1200000, 0, "GA", 10), // Fails price
		createFilterTestFlight("3", 900000, 2, "GA", 12),  // Fails stops
		createFilterTestFlight("4", 850000, 0, "JT", 9),   // Fails airline
		createFilterTestFlight("5", 750000, 0, "GA", 6),   // Fails time (6 AM)
		createFilterTestFlight("6", 950000, 1, "GA", 11),  // Passes all
	}

	maxPrice := float64(1000000)
	maxStops := 1
	opts := &domain.FilterOptions{
		MaxPrice: &maxPrice,
		MaxStops: &maxStops,
		Airlines: []string{"GA"},
		DepartureTimeRange: &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
		},
	}

	result := ApplyFilters(flights, opts)

	require.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "6", result[1].ID)
}

// TestApplyFilters_ArrivalTimeRange tests arrival time range filtering.
func TestApplyFilters_ArrivalTimeRange(t *testing.T) {
	// Create flights with varying arrival times
	flights := []domain.Flight{
		createFlightWithArrival("morning", 1000000, 0, "GA", 8),  // Arrives 8:00
		createFlightWithArrival("noon", 900000, 0, "JT", 12),      // Arrives 12:00
		createFlightWithArrival("afternoon", 800000, 0, "ID", 16), // Arrives 16:00
		createFlightWithArrival("evening", 700000, 0, "AK", 20),   // Arrives 20:00
	}

	tests := []struct {
		name        string
		startHour   int
		endHour     int
		expectedIDs []string
	}{
		{"business hours arrivals", 8, 17, []string{"morning", "noon", "afternoon"}},
		{"afternoon only", 14, 18, []string{"afternoon"}},
		{"evening arrivals", 18, 23, []string{"evening"}},
		{"early arrivals", 6, 10, []string{"morning"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &domain.FilterOptions{
				ArrivalTimeRange: &domain.TimeRange{
					Start: time.Date(2025, 1, 1, tt.startHour, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, tt.endHour, 0, 0, 0, time.UTC),
				},
			}

			result := ApplyFilters(flights, opts)
			require.Len(t, result, len(tt.expectedIDs))

			resultIDs := make([]string, len(result))
			for i, f := range result {
				resultIDs[i] = f.ID
			}
			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}

// TestApplyFilters_DepartureAndArrivalTimeRange tests combined departure and arrival time filtering.
func TestApplyFilters_DepartureAndArrivalTimeRange(t *testing.T) {
	// Create flights with specific departure and arrival times
	// Flight departs at departHour and arrives 2 hours later
	flights := []domain.Flight{
		{
			ID:           "early-morning",
			FlightNumber: "FL-001",
			Airline:      domain.AirlineInfo{Code: "GA", Name: "Garuda"},
			Departure:    domain.FlightPoint{AirportCode: "CGK", DateTime: time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC)},
			Arrival:      domain.FlightPoint{AirportCode: "DPS", DateTime: time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)},
			Duration:     domain.DurationInfo{TotalMinutes: 120},
			Price:        domain.PriceInfo{Amount: 800000, Currency: "IDR"},
			Stops:        0,
		},
		{
			ID:           "morning-arrival",
			FlightNumber: "FL-002",
			Airline:      domain.AirlineInfo{Code: "JT", Name: "Lion"},
			Departure:    domain.FlightPoint{AirportCode: "CGK", DateTime: time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)},
			Arrival:      domain.FlightPoint{AirportCode: "DPS", DateTime: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)},
			Duration:     domain.DurationInfo{TotalMinutes: 120},
			Price:        domain.PriceInfo{Amount: 750000, Currency: "IDR"},
			Stops:        0,
		},
		{
			ID:           "afternoon",
			FlightNumber: "FL-003",
			Airline:      domain.AirlineInfo{Code: "ID", Name: "Batik"},
			Departure:    domain.FlightPoint{AirportCode: "CGK", DateTime: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)},
			Arrival:      domain.FlightPoint{AirportCode: "DPS", DateTime: time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)},
			Duration:     domain.DurationInfo{TotalMinutes: 120},
			Price:        domain.PriceInfo{Amount: 900000, Currency: "IDR"},
			Stops:        0,
		},
		{
			ID:           "evening",
			FlightNumber: "FL-004",
			Airline:      domain.AirlineInfo{Code: "AK", Name: "AirAsia"},
			Departure:    domain.FlightPoint{AirportCode: "CGK", DateTime: time.Date(2025, 12, 15, 16, 0, 0, 0, time.UTC)},
			Arrival:      domain.FlightPoint{AirportCode: "DPS", DateTime: time.Date(2025, 12, 15, 18, 0, 0, 0, time.UTC)},
			Duration:     domain.DurationInfo{TotalMinutes: 120},
			Price:        domain.PriceInfo{Amount: 850000, Currency: "IDR"},
			Stops:        0,
		},
	}

	t.Run("morning departure with business hours arrival", func(t *testing.T) {
		opts := &domain.FilterOptions{
			DepartureTimeRange: &domain.TimeRange{
				Start: time.Date(2025, 1, 1, 6, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			ArrivalTimeRange: &domain.TimeRange{
				Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		}

		result := ApplyFilters(flights, opts)
		require.Len(t, result, 3)
		
		resultIDs := make([]string, len(result))
		for i, f := range result {
			resultIDs[i] = f.ID
		}
		assert.ElementsMatch(t, []string{"early-morning", "morning-arrival", "afternoon"}, resultIDs)
	})

	t.Run("strict business hours both ways", func(t *testing.T) {
		opts := &domain.FilterOptions{
			DepartureTimeRange: &domain.TimeRange{
				Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			ArrivalTimeRange: &domain.TimeRange{
				Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
			},
		}

		result := ApplyFilters(flights, opts)
		require.Len(t, result, 2)
		assert.ElementsMatch(t, []string{"morning-arrival", "afternoon"}, 
			[]string{result[0].ID, result[1].ID})
	})
}

// TestApplyFilters_AllFiltersOneMatch tests combined filters with single match.
func TestApplyFilters_AllFiltersOneMatch(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 2000000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 2, "GA", 10),
		createFilterTestFlight("3", 1000000, 0, "JT", 12),
		createFilterTestFlight("perfect", 900000, 0, "GA", 10), // Only this matches
	}

	maxPrice := float64(1000000)
	maxStops := 1
	opts := &domain.FilterOptions{
		MaxPrice: &maxPrice,
		MaxStops: &maxStops,
		Airlines: []string{"GA"},
	}

	result := ApplyFilters(flights, opts)

	require.Len(t, result, 1)
	assert.Equal(t, "perfect", result[0].ID)
}

// TestFilterByMaxPrice tests the individual price filter function.
func TestFilterByMaxPrice(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 500000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 0, "JT", 10),
		createFilterTestFlight("3", 1500000, 0, "ID", 14),
	}

	t.Run("nil maxPrice returns all", func(t *testing.T) {
		result := FilterByMaxPrice(flights, nil)
		assert.Len(t, result, 3)
	})

	t.Run("filters by max price", func(t *testing.T) {
		maxPrice := float64(1000000)
		result := FilterByMaxPrice(flights, &maxPrice)
		assert.Len(t, result, 2)
	})
}

// TestFilterByMaxStops tests the individual stops filter function.
func TestFilterByMaxStops(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 1, "JT", 10),
		createFilterTestFlight("3", 1000000, 2, "ID", 14),
	}

	t.Run("nil maxStops returns all", func(t *testing.T) {
		result := FilterByMaxStops(flights, nil)
		assert.Len(t, result, 3)
	})

	t.Run("filters direct only", func(t *testing.T) {
		maxStops := 0
		result := FilterByMaxStops(flights, &maxStops)
		require.Len(t, result, 1)
		assert.Equal(t, 0, result[0].Stops)
	})
}

// TestFilterByAirlines tests the individual airlines filter function.
func TestFilterByAirlines(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 0, "JT", 10),
		createFilterTestFlight("3", 1000000, 0, "ID", 14),
	}

	t.Run("nil airlines returns all", func(t *testing.T) {
		result := FilterByAirlines(flights, nil)
		assert.Len(t, result, 3)
	})

	t.Run("empty airlines returns all", func(t *testing.T) {
		result := FilterByAirlines(flights, []string{})
		assert.Len(t, result, 3)
	})

	t.Run("filters by single airline", func(t *testing.T) {
		result := FilterByAirlines(flights, []string{"GA"})
		require.Len(t, result, 1)
		assert.Equal(t, "GA", result[0].Airline.Code)
	})

	t.Run("case insensitive", func(t *testing.T) {
		result := FilterByAirlines(flights, []string{"ga", "jt"})
		assert.Len(t, result, 2)
	})
}

// TestFilterByDepartureTime tests the individual departure time filter function.
func TestFilterByDepartureTime(t *testing.T) {
	flights := []domain.Flight{
		createFilterTestFlight("1", 1000000, 0, "GA", 8),
		createFilterTestFlight("2", 1000000, 0, "JT", 12),
		createFilterTestFlight("3", 1000000, 0, "ID", 18),
	}

	t.Run("nil timeRange returns all", func(t *testing.T) {
		result := FilterByDepartureTime(flights, nil)
		assert.Len(t, result, 3)
	})

	t.Run("filters by time range", func(t *testing.T) {
		timeRange := &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
		}
		result := FilterByDepartureTime(flights, timeRange)
		require.Len(t, result, 1)
		assert.Equal(t, 12, result[0].Departure.DateTime.Hour())
	})
}

// TestFilterByArrivalTime tests the individual arrival time filter function.
func TestFilterByArrivalTime(t *testing.T) {
	// Create flights with different arrival times
	flights := []domain.Flight{
		createFlightWithArrival("1", 1000000, 0, "GA", 10), // Arrives at 10:00
		createFlightWithArrival("2", 1000000, 0, "JT", 14), // Arrives at 14:00
		createFlightWithArrival("3", 1000000, 0, "ID", 20), // Arrives at 20:00
	}

	t.Run("nil timeRange returns all", func(t *testing.T) {
		result := FilterByArrivalTime(flights, nil)
		assert.Len(t, result, 3)
	})

	t.Run("filters by arrival time range", func(t *testing.T) {
		timeRange := &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 16, 0, 0, 0, time.UTC),
		}
		result := FilterByArrivalTime(flights, timeRange)
		require.Len(t, result, 1)
		assert.Equal(t, "2", result[0].ID)
		assert.Equal(t, 14, result[0].Arrival.DateTime.Hour())
	})

	t.Run("filters early morning arrivals", func(t *testing.T) {
		timeRange := &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		}
		result := FilterByArrivalTime(flights, timeRange)
		require.Len(t, result, 1)
		assert.Equal(t, "1", result[0].ID)
	})

	t.Run("filters evening arrivals", func(t *testing.T) {
		timeRange := &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 18, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC),
		}
		result := FilterByArrivalTime(flights, timeRange)
		require.Len(t, result, 1)
		assert.Equal(t, "3", result[0].ID)
		assert.Equal(t, 20, result[0].Arrival.DateTime.Hour())
	})

	t.Run("no matches outside range", func(t *testing.T) {
		timeRange := &domain.TimeRange{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 6, 0, 0, 0, time.UTC),
		}
		result := FilterByArrivalTime(flights, timeRange)
		assert.Len(t, result, 0)
	})
}

// TestBuildAirlineSet tests the airline set builder.
func TestBuildAirlineSet(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		set := buildAirlineSet([]string{})
		assert.Empty(t, set)
	})

	t.Run("single airline", func(t *testing.T) {
		set := buildAirlineSet([]string{"GA"})
		assert.Len(t, set, 1)
		_, exists := set["GA"]
		assert.True(t, exists)
	})

	t.Run("multiple airlines uppercase", func(t *testing.T) {
		set := buildAirlineSet([]string{"ga", "jt", "id"})
		assert.Len(t, set, 3)
		_, exists := set["GA"]
		assert.True(t, exists)
		_, exists = set["JT"]
		assert.True(t, exists)
		_, exists = set["ID"]
		assert.True(t, exists)
	})

	t.Run("duplicates handled", func(t *testing.T) {
		set := buildAirlineSet([]string{"GA", "ga", "Ga"})
		assert.Len(t, set, 1)
	})
}

// TestIsAirlineInSet tests the airline set lookup.
func TestIsAirlineInSet(t *testing.T) {
	set := buildAirlineSet([]string{"GA", "JT"})

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{"exact match uppercase", "GA", true},
		{"lowercase matches", "ga", true},
		{"mixed case matches", "Ga", true},
		{"not in set", "ID", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isAirlineInSet(tt.code, set))
		})
	}
}

// TestFilterByDuration tests the duration filter function.
func TestFilterByDuration(t *testing.T) {
	// Create flights with different durations
	createFlightWithDuration := func(id string, durationMinutes int) domain.Flight {
		f := createFilterTestFlight(id, 1000000, 0, "GA", 8)
		f.Duration.TotalMinutes = durationMinutes
		return f
	}

	flights := []domain.Flight{
		createFlightWithDuration("1", 60),   // 1 hour
		createFlightWithDuration("2", 120),  // 2 hours
		createFlightWithDuration("3", 180),  // 3 hours
		createFlightWithDuration("4", 240),  // 4 hours
		createFlightWithDuration("5", 360),  // 6 hours
	}

	tests := []struct {
		name          string
		durationRange *domain.DurationRange
		expectedIDs   []string
	}{
		{
			name:          "nil duration range returns all flights",
			durationRange: nil,
			expectedIDs:   []string{"1", "2", "3", "4", "5"},
		},
		{
			name:          "only min duration - filters flights below minimum",
			durationRange: &domain.DurationRange{MinMinutes: intPtr(120)},
			expectedIDs:   []string{"2", "3", "4", "5"},
		},
		{
			name:          "only max duration - filters flights above maximum",
			durationRange: &domain.DurationRange{MaxMinutes: intPtr(240)},
			expectedIDs:   []string{"1", "2", "3", "4"},
		},
		{
			name:          "both min and max - filters flights outside range",
			durationRange: &domain.DurationRange{MinMinutes: intPtr(120), MaxMinutes: intPtr(240)},
			expectedIDs:   []string{"2", "3", "4"},
		},
		{
			name:          "exact match on boundaries included",
			durationRange: &domain.DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name:          "narrow range with one match",
			durationRange: &domain.DurationRange{MinMinutes: intPtr(175), MaxMinutes: intPtr(185)},
			expectedIDs:   []string{"3"},
		},
		{
			name:          "no flights match the range",
			durationRange: &domain.DurationRange{MinMinutes: intPtr(400), MaxMinutes: intPtr(500)},
			expectedIDs:   []string{},
		},
		{
			name:          "empty range (no bounds) returns all",
			durationRange: &domain.DurationRange{},
			expectedIDs:   []string{"1", "2", "3", "4", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterByDuration(flights, tt.durationRange)

			assert.Len(t, result, len(tt.expectedIDs))

			resultIDs := make([]string, len(result))
			for i, f := range result {
				resultIDs[i] = f.ID
			}

			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}

// TestApplyFilters_DurationRange tests duration range filtering through ApplyFilters.
func TestApplyFilters_DurationRange(t *testing.T) {
	createFlightWithDuration := func(id string, price float64, durationMinutes int) domain.Flight {
		f := createFilterTestFlight(id, price, 0, "GA", 8)
		f.Duration.TotalMinutes = durationMinutes
		return f
	}

	flights := []domain.Flight{
		createFlightWithDuration("1", 1000000, 90),   // 1.5 hours, cheap
		createFlightWithDuration("2", 1500000, 120),  // 2 hours, medium
		createFlightWithDuration("3", 2000000, 180),  // 3 hours, expensive
		createFlightWithDuration("4", 800000, 240),   // 4 hours, very cheap
	}

	tests := []struct {
		name        string
		opts        *domain.FilterOptions
		expectedIDs []string
	}{
		{
			name: "duration range only",
			opts: &domain.FilterOptions{
				DurationRange: &domain.DurationRange{
					MinMinutes: intPtr(90),
					MaxMinutes: intPtr(180),
				},
			},
			expectedIDs: []string{"1", "2", "3"},
		},
		{
			name: "duration + price filter combined",
			opts: &domain.FilterOptions{
				DurationRange: &domain.DurationRange{
					MinMinutes: intPtr(90),
					MaxMinutes: intPtr(180),
				},
				MaxPrice: floatPtr(1500000),
			},
			expectedIDs: []string{"1", "2"},
		},
		{
			name: "all filters combined including duration",
			opts: &domain.FilterOptions{
				DurationRange: &domain.DurationRange{
					MinMinutes: intPtr(100),
					MaxMinutes: intPtr(200),
				},
				MaxPrice: floatPtr(2000000),
				MaxStops: intPtr(1),
			},
			expectedIDs: []string{"2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(flights, tt.opts)

			resultIDs := make([]string, len(result))
			for i, f := range result {
				resultIDs[i] = f.ID
			}

			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}

// TestApplyFilters_Performance verifies O(n) performance characteristic.
func TestApplyFilters_Performance(t *testing.T) {
	// Create a large list of flights
	n := 10000
	flights := make([]domain.Flight, n)
	for i := 0; i < n; i++ {
		flights[i] = createFilterTestFlight(
			"id-"+intToString(i),
			float64(500000+i*100),
			i%4,
			[]string{"GA", "JT", "ID", "AK"}[i%4],
			8+i%12,
		)
	}

	maxPrice := float64(700000)
	maxStops := 1
	opts := &domain.FilterOptions{
		MaxPrice: &maxPrice,
		MaxStops: &maxStops,
		Airlines: []string{"GA", "JT"},
	}

	// This should complete in reasonable time (O(n))
	start := time.Now()
	result := ApplyFilters(flights, opts)
	elapsed := time.Since(start)

	// Should complete in under 100ms for 10k flights
	assert.Less(t, elapsed, 100*time.Millisecond, "Filter should be O(n)")
	assert.Greater(t, len(result), 0, "Should have some matches")
}

// intToString is a simple helper to convert int to string without fmt.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
