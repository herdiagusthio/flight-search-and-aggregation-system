package integration

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
	"github.com/flight-search/flight-search-and-aggregation-system/test/mock"
)

// TestHandler_SearchFlights_Success tests successful flight search via HTTP.
func TestHandler_SearchFlights_Success(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 3))
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	req := DefaultSearchRequest()

	// Act
	resp := ts.SearchRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)
	assert.Len(t, searchResp.Flights, 3)
	assert.Equal(t, 3, searchResp.Metadata.TotalResults)
	assert.Greater(t, searchResp.Metadata.ProvidersQueried, 0)
}

// TestHandler_ResponseBodyStructure tests that the response body has correct structure.
func TestHandler_ResponseBodyStructure(t *testing.T) {
	// Arrange
	flights := []domain.Flight{
		{
			ID:           "test-1",
			FlightNumber: "GA 100",
			Airline: domain.AirlineInfo{
				Code: "GA",
				Name: "Garuda Indonesia",
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				AirportName: "Soekarno-Hatta",
				DateTime:    time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC),
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				AirportName: "Ngurah Rai",
				DateTime:    time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC),
			},
			Duration: domain.DurationInfo{
				TotalMinutes: 150,
				Formatted:    "2h 30m",
			},
			Price: domain.PriceInfo{
				Amount:   1500000,
				Currency: "IDR",
			},
			Baggage: domain.BaggageInfo{
				CabinKg:   7,
				CheckedKg: 20,
			},
			Class:    "economy",
			Stops:    0,
			Provider: "garuda",
		},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Act
	resp := ts.SearchRequest(DefaultSearchRequest())

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)
	require.Len(t, searchResp.Flights, 1)

	flight := searchResp.Flights[0]
	assert.Equal(t, "test-1", flight.ID)
	assert.Equal(t, "GA 100", flight.FlightNumber)
	assert.Equal(t, "GA", flight.Airline.Code)
	assert.Equal(t, "Garuda Indonesia", flight.Airline.Name)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, 150, flight.Duration.TotalMinutes)
	assert.Equal(t, float64(1500000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 0, flight.Stops)
	assert.Equal(t, "garuda", flight.Provider)
}

// TestHandler_MetadataInResponse tests that metadata is correctly populated.
func TestHandler_MetadataInResponse(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 2))
	provider2 := mock.NewProvider("lion").WithError(errors.New("unavailable"))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2})
	ts := NewTestServer(uc)

	// Act
	resp := ts.SearchRequest(DefaultSearchRequest())

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)

	// Verify metadata
	assert.Equal(t, 2, searchResp.Metadata.TotalResults)
	assert.GreaterOrEqual(t, searchResp.Metadata.SearchTimeMs, int64(0))
	assert.Equal(t, 2, searchResp.Metadata.ProvidersQueried)
	assert.Equal(t, 1, searchResp.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, searchResp.Metadata.ProvidersFailed)
}

// TestHandler_ValidationErrors tests various validation error scenarios.
func TestHandler_ValidationErrors(t *testing.T) {
	futureDate := FutureDate()

	tests := []struct {
		name         string
		body         map[string]interface{}
		wantCode     int
		wantContains string
	}{
		{
			name: "missing origin",
			body: map[string]interface{}{
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "origin",
		},
		{
			name: "empty origin",
			body: map[string]interface{}{
				"origin":        "",
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "origin",
		},
		{
			name: "invalid origin code - too long",
			body: map[string]interface{}{
				"origin":        "CGKK",
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "origin",
		},
		{
			name: "missing destination",
			body: map[string]interface{}{
				"origin":        "CGK",
				"departureDate": futureDate,
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "destination",
		},
		{
			name: "same origin and destination",
			body: map[string]interface{}{
				"origin":        "CGK",
				"destination":   "CGK",
				"departureDate": futureDate,
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "different",
		},
		{
			name: "missing departure date",
			body: map[string]interface{}{
				"origin":      "CGK",
				"destination": "DPS",
				"passengers":  1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "departure",
		},
		{
			name: "invalid date format",
			body: map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": "15-12-2025",
				"passengers":    1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "departure",
		},
		{
			name: "zero passengers",
			body: map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    0,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "passengers",
		},
		{
			name: "negative passengers",
			body: map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    -1,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "passengers",
		},
		{
			name: "too many passengers",
			body: map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": futureDate,
				"passengers":    10,
			},
			wantCode:     http.StatusBadRequest,
			wantContains: "passengers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Mock use case doesn't matter for validation errors
			provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 1))
			uc := CreateUseCase([]domain.FlightProvider{provider})
			ts := NewTestServer(uc)

			// Act
			resp := ts.SearchRequest(tt.body)

			// Assert
			assert.Equal(t, tt.wantCode, resp.Code, "status code mismatch")
			assert.Contains(t, string(resp.Body), tt.wantContains, "expected error message not found")
		})
	}
}

// TestHandler_ServiceUnavailable tests that 503 is returned when all providers fail.
func TestHandler_ServiceUnavailable(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithError(errors.New("unavailable"))
	provider2 := mock.NewProvider("lion").WithError(errors.New("unavailable"))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2})
	ts := NewTestServer(uc)

	// Act
	resp := ts.SearchRequest(DefaultSearchRequest())

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

// TestHandler_Timeout tests that 504 is returned on timeout.
func TestHandler_Timeout(t *testing.T) {
	// Arrange - Provider takes longer than timeout
	slowProvider := mock.NewProvider("slow").
		WithDelay(500 * time.Millisecond).
		WithFlights(mock.SampleFlights("slow", 1))

	config := &usecase.Config{
		GlobalTimeout:   100 * time.Millisecond,
		ProviderTimeout: 50 * time.Millisecond,
	}

	uc := CreateUseCaseWithConfig([]domain.FlightProvider{slowProvider}, config)
	ts := NewTestServer(uc)

	// Act
	resp := ts.SearchRequest(DefaultSearchRequest())

	// Assert - Should return 503 because all providers failed (due to timeout)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

// TestHandler_HealthCheck tests the health endpoint.
func TestHandler_HealthCheck(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 1))
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Act
	resp := ts.HealthRequest()

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)
}

// TestHandler_InvalidJSON tests that invalid JSON returns 400.
func TestHandler_InvalidJSON(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 1))
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Act - Send raw string instead of marshaled JSON
	resp := ts.Do(Request{
		Method:      http.MethodPost,
		Path:        "/api/v1/flights/search",
		Body:        nil,
		ContentType: "application/json",
	})

	// Send invalid JSON as raw bytes
	resp = ts.Do(Request{
		Method:      http.MethodPost,
		Path:        "/api/v1/flights/search",
		ContentType: "application/json",
	})

	// Assert - Empty body should fail validation
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestHandler_MultipleProvidersSuccess tests aggregation via HTTP.
func TestHandler_MultipleProvidersSuccess(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 2))
	provider2 := mock.NewProvider("lion").WithFlights(mock.SampleFlights("lion", 3))
	provider3 := mock.NewProvider("batik").WithFlights(mock.SampleFlights("batik", 1))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2, provider3})
	ts := NewTestServer(uc)

	// Act
	resp := ts.SearchRequest(DefaultSearchRequest())

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)
	assert.Len(t, searchResp.Flights, 6) // 2 + 3 + 1
	assert.Equal(t, 3, searchResp.Metadata.ProvidersQueried)
}

// TestHandler_FiltersApplied tests that filters are applied via HTTP.
func TestHandler_FiltersApplied(t *testing.T) {
	// Arrange - Create flights with different prices
	flights := []domain.Flight{
		{ID: "1", Price: domain.PriceInfo{Amount: 500000, Currency: "IDR"}, Provider: "garuda"},
		{ID: "2", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Provider: "garuda"},
		{ID: "3", Price: domain.PriceInfo{Amount: 2000000, Currency: "IDR"}, Provider: "garuda"},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Request with filter
	req := map[string]interface{}{
		"origin":        "CGK",
		"destination":   "DPS",
		"departureDate": FutureDate(),
		"passengers":    1,
		"filters": map[string]interface{}{
			"maxPrice": 1000000,
		},
	}

	// Act
	resp := ts.SearchRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)
	assert.Len(t, searchResp.Flights, 2) // Only flights <= 1,000,000
}

// TestHandler_SortingApplied tests that sorting is applied via HTTP.
func TestHandler_SortingApplied(t *testing.T) {
	// Arrange - Create flights with different prices
	flights := []domain.Flight{
		{ID: "1", Price: domain.PriceInfo{Amount: 1500000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 150}, Provider: "garuda"},
		{ID: "2", Price: domain.PriceInfo{Amount: 500000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 180}, Provider: "garuda"},
		{ID: "3", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 120}, Provider: "garuda"},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Request with sorting
	req := map[string]interface{}{
		"origin":        "CGK",
		"destination":   "DPS",
		"departureDate": FutureDate(),
		"passengers":    1,
		"sortBy":        "price",
	}

	// Act
	resp := ts.SearchRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)
	require.Len(t, searchResp.Flights, 3)

	// Verify sorted by price
	assert.Equal(t, 500000.0, searchResp.Flights[0].Price.Amount)
	assert.Equal(t, 1000000.0, searchResp.Flights[1].Price.Amount)
	assert.Equal(t, 1500000.0, searchResp.Flights[2].Price.Amount)
}

// TestHandler_DurationRangeFilter tests duration range filtering via HTTP.
func TestHandler_DurationRangeFilter(t *testing.T) {
	// Arrange - Create flights with different durations
	flights := []domain.Flight{
		{ID: "1", FlightNumber: "GA 100", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 90}, Provider: "garuda"},   // 1.5h
		{ID: "2", FlightNumber: "GA 101", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 120}, Provider: "garuda"},  // 2h
		{ID: "3", FlightNumber: "GA 102", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 180}, Provider: "garuda"},  // 3h
		{ID: "4", FlightNumber: "GA 103", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 240}, Provider: "garuda"},  // 4h
		{ID: "5", FlightNumber: "GA 104", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 360}, Provider: "garuda"},  // 6h
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	tests := []struct {
		name           string
		durationFilter map[string]interface{}
		expectedIDs    []string
	}{
		{
			name: "only min duration",
			durationFilter: map[string]interface{}{
				"minMinutes": 120,
			},
			expectedIDs: []string{"2", "3", "4", "5"},
		},
		{
			name: "only max duration",
			durationFilter: map[string]interface{}{
				"maxMinutes": 240,
			},
			expectedIDs: []string{"1", "2", "3", "4"},
		},
		{
			name: "both min and max duration",
			durationFilter: map[string]interface{}{
				"minMinutes": 120,
				"maxMinutes": 240,
			},
			expectedIDs: []string{"2", "3", "4"},
		},
		{
			name: "narrow range",
			durationFilter: map[string]interface{}{
				"minMinutes": 175,
				"maxMinutes": 185,
			},
			expectedIDs: []string{"3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Request with duration filter
			req := map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": FutureDate(),
				"passengers":    1,
				"filters": map[string]interface{}{
					"durationRange": tt.durationFilter,
				},
			}

			// Act
			resp := ts.SearchRequest(req)

			// Assert
			assert.Equal(t, http.StatusOK, resp.Code)

			searchResp, err := resp.ParseSearchResponse()
			require.NoError(t, err)
			assert.Len(t, searchResp.Flights, len(tt.expectedIDs))

			// Extract IDs
			resultIDs := make([]string, len(searchResp.Flights))
			for i, f := range searchResp.Flights {
				resultIDs[i] = f.ID
			}

			assert.ElementsMatch(t, tt.expectedIDs, resultIDs)
		})
	}
}

// TestHandler_CombinedFiltersWithDuration tests combining duration with other filters.
func TestHandler_CombinedFiltersWithDuration(t *testing.T) {
	// Arrange
	flights := []domain.Flight{
		{ID: "1", FlightNumber: "GA 100", Price: domain.PriceInfo{Amount: 800000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 120}, Stops: 0, Provider: "garuda"},
		{ID: "2", FlightNumber: "GA 101", Price: domain.PriceInfo{Amount: 1200000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 180}, Stops: 1, Provider: "garuda"},
		{ID: "3", FlightNumber: "GA 102", Price: domain.PriceInfo{Amount: 900000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 240}, Stops: 0, Provider: "garuda"},
		{ID: "4", FlightNumber: "GA 103", Price: domain.PriceInfo{Amount: 700000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 300}, Stops: 2, Provider: "garuda"},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	// Request with combined filters: duration + price + stops
	req := map[string]interface{}{
		"origin":        "CGK",
		"destination":   "DPS",
		"departureDate": FutureDate(),
		"passengers":    1,
		"filters": map[string]interface{}{
			"durationRange": map[string]interface{}{
				"minMinutes": 100,
				"maxMinutes": 250,
			},
			"maxPrice": 1000000,
			"maxStops": 0,
		},
	}

	// Act
	resp := ts.SearchRequest(req)

	// Assert
	assert.Equal(t, http.StatusOK, resp.Code)

	searchResp, err := resp.ParseSearchResponse()
	require.NoError(t, err)

	// Flights that match all criteria:
	// - Flight ID "1": Duration 120 (within 100-250), Price 800000 (<=1000000), Stops 0 (<=0) ✓
	// - Flight ID "3": Duration 240 (within 100-250), Price 900000 (<=1000000), Stops 0 (<=0) ✓
	require.Len(t, searchResp.Flights, 2)

	resultIDs := make([]string, len(searchResp.Flights))
	for i, f := range searchResp.Flights {
		resultIDs[i] = f.ID
	}
	assert.ElementsMatch(t, []string{"1", "3"}, resultIDs)
}

// TestHandler_InvalidDurationRange tests validation of invalid duration ranges.
func TestHandler_InvalidDurationRange(t *testing.T) {
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 3))
	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	tests := []struct {
		name           string
		durationFilter map[string]interface{}
	}{
		{
			name: "negative minMinutes",
			durationFilter: map[string]interface{}{
				"minMinutes": -10,
			},
		},
		{
			name: "negative maxMinutes",
			durationFilter: map[string]interface{}{
				"maxMinutes": -50,
			},
		},
		{
			name: "min greater than max",
			durationFilter: map[string]interface{}{
				"minMinutes": 300,
				"maxMinutes": 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := map[string]interface{}{
				"origin":        "CGK",
				"destination":   "DPS",
				"departureDate": FutureDate(),
				"passengers":    1,
				"filters": map[string]interface{}{
					"durationRange": tt.durationFilter,
				},
			}

			// Act
			resp := ts.SearchRequest(req)

			// Assert - Should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, resp.Code)
		})
	}
}
