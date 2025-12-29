package airasia

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapter_Name verifies the adapter returns the correct provider name.
func TestAdapter_Name(t *testing.T) {
	adapter := NewAdapter("mock.json")
	assert.Equal(t, "airasia", adapter.Name())
}

// TestAdapter_ImplementsInterface ensures compile-time interface verification.
func TestAdapter_ImplementsInterface(t *testing.T) {
	var _ domain.FlightProvider = (*Adapter)(nil)
}

// TestAdapter_Search tests the Search method with various scenarios.
func TestAdapter_Search(t *testing.T) {
	tests := []struct {
		name           string
		jsonData       string
		criteria       domain.SearchCriteria
		expectedCount  int
		expectedError  bool
		errorRetryable bool
	}{
		{
			name: "successful parsing with multiple flights",
			jsonData: `{
				"status": "ok",
				"flights": [
					{
						"flight_code": "QZ520",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T04:45:00+07:00",
						"arrive_time": "2025-12-15T07:25:00+08:00",
						"duration_hours": 1.67,
						"direct_flight": true,
						"price_idr": 650000,
						"seats": 67,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only, checked bags additional fee"
					},
					{
						"flight_code": "QZ524",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T10:00:00+07:00",
						"arrive_time": "2025-12-15T12:45:00+08:00",
						"duration_hours": 1.75,
						"direct_flight": true,
						"price_idr": 720000,
						"seats": 54,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only, checked bags additional fee"
					}
				]
			}`,
			criteria:      domain.SearchCriteria{},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "empty flights array returns empty slice",
			jsonData: `{
				"status": "ok",
				"flights": []
			}`,
			criteria:      domain.SearchCriteria{},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "filter by origin",
			jsonData: `{
				"status": "ok",
				"flights": [
					{
						"flight_code": "QZ520",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T04:45:00+07:00",
						"arrive_time": "2025-12-15T07:25:00+08:00",
						"duration_hours": 1.67,
						"direct_flight": true,
						"price_idr": 650000,
						"seats": 67,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only"
					},
					{
						"flight_code": "QZ600",
						"airline": "AirAsia",
						"from_airport": "SUB",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T08:00:00+07:00",
						"arrive_time": "2025-12-15T09:30:00+08:00",
						"duration_hours": 0.5,
						"direct_flight": true,
						"price_idr": 450000,
						"seats": 80,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only"
					}
				]
			}`,
			criteria:      domain.SearchCriteria{Origin: "CGK"},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "filter by departure date",
			jsonData: `{
				"status": "ok",
				"flights": [
					{
						"flight_code": "QZ520",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T04:45:00+07:00",
						"arrive_time": "2025-12-15T07:25:00+08:00",
						"duration_hours": 1.67,
						"direct_flight": true,
						"price_idr": 650000,
						"seats": 67,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only"
					},
					{
						"flight_code": "QZ524",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-16T10:00:00+07:00",
						"arrive_time": "2025-12-16T12:45:00+08:00",
						"duration_hours": 1.75,
						"direct_flight": true,
						"price_idr": 720000,
						"seats": 54,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only"
					}
				]
			}`,
			criteria:      domain.SearchCriteria{DepartureDate: "2025-12-15"},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:           "malformed JSON returns error",
			jsonData:       `{invalid json}`,
			criteria:       domain.SearchCriteria{},
			expectedCount:  0,
			expectedError:  true,
			errorRetryable: false,
		},
		{
			name: "connecting flight with stops",
			jsonData: `{
				"status": "ok",
				"flights": [
					{
						"flight_code": "QZ7250",
						"airline": "AirAsia",
						"from_airport": "CGK",
						"to_airport": "DPS",
						"depart_time": "2025-12-15T15:15:00+07:00",
						"arrive_time": "2025-12-15T20:35:00+08:00",
						"duration_hours": 4.33,
						"direct_flight": false,
						"stops": [{"airport": "SOC", "wait_time_minutes": 95}],
						"price_idr": 485000,
						"seats": 88,
						"cabin_class": "economy",
						"baggage_note": "Cabin baggage only, checked bags additional fee"
					}
				]
			}`,
			criteria:      domain.SearchCriteria{},
			expectedCount: 1,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file with test data
			tmpFile := createTempFile(t, tt.jsonData)
			defer os.Remove(tmpFile)

			adapter := NewAdapter(tmpFile)
			ctx := context.Background()

			flights, err := adapter.Search(ctx, tt.criteria)

			if tt.expectedError {
				require.Error(t, err)
				var providerErr *domain.ProviderError
				if assert.ErrorAs(t, err, &providerErr) {
					assert.Equal(t, "airasia", providerErr.Provider)
					assert.Equal(t, tt.errorRetryable, providerErr.Retryable)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, flights, tt.expectedCount)

				// Verify all flights have provider set
				for _, f := range flights {
					assert.Equal(t, "airasia", f.Provider)
				}
			}
		})
	}
}

// TestAdapter_Search_FileNotFound tests error handling for missing files.
func TestAdapter_Search_FileNotFound(t *testing.T) {
	adapter := NewAdapter("nonexistent_file.json")
	ctx := context.Background()

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	require.Error(t, err)
	assert.Nil(t, flights)

	var providerErr *domain.ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "airasia", providerErr.Provider)
	assert.True(t, providerErr.Retryable, "File read errors should be retryable")
}

// TestAdapter_Search_ContextCancellation tests context cancellation handling.
func TestAdapter_Search_ContextCancellation(t *testing.T) {
	adapter := NewAdapter("mock.json")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	require.Error(t, err)
	assert.Nil(t, flights)

	var providerErr *domain.ProviderError
	require.ErrorAs(t, err, &providerErr)
	assert.Equal(t, "airasia", providerErr.Provider)
	assert.False(t, providerErr.Retryable)
	assert.ErrorIs(t, providerErr.Err, context.Canceled)
}

// TestHoursToMinutes tests the float to minutes conversion.
func TestHoursToMinutes(t *testing.T) {
	tests := []struct {
		name     string
		hours    float64
		expected int
	}{
		{"1.75 hours", 1.75, 105},
		{"2.5 hours", 2.5, 150},
		{"0.5 hours", 0.5, 30},
		{"1.0 hours", 1.0, 60},
		{"0 hours", 0, 0},
		{"1.67 hours", 1.67, 100}, // Rounds to nearest
		{"2.25 hours", 2.25, 135},
		{"3.33 hours", 3.33, 200}, // Rounds properly
		{"4.33 hours", 4.33, 260},
		{"0.25 hours", 0.25, 15},
		{"0.1 hours", 0.1, 6},
		{"5.99 hours", 5.99, 359},
		{"negative hours", -1.5, -90}, // Edge case: negative
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hoursToMinutes(tt.hours)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatDurationFromHours tests duration formatting.
func TestFormatDurationFromHours(t *testing.T) {
	tests := []struct {
		name     string
		hours    float64
		expected string
	}{
		{"1h 45m", 1.75, "1h 45m"},
		{"2h 30m", 2.5, "2h 30m"},
		{"30m only", 0.5, "30m"},
		{"1h only", 1.0, "1h"},
		{"0 minutes", 0, "0m"},
		{"1h 40m", 1.67, "1h 40m"},
		{"2h 15m", 2.25, "2h 15m"},
		{"4h 20m", 4.33, "4h 20m"},
		{"15m only", 0.25, "15m"},
		{"3h only", 3.0, "3h"},
		{"5h 59m", 5.99, "5h 59m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDurationFromHours(tt.hours)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDirectFlightToStops tests the direct flight to stops conversion.
func TestDirectFlightToStops(t *testing.T) {
	tests := []struct {
		name     string
		isDirect bool
		stops    []AirAsiaStop
		expected int
	}{
		{
			name:     "direct flight",
			isDirect: true,
			stops:    nil,
			expected: 0,
		},
		{
			name:     "not direct with no stops array",
			isDirect: false,
			stops:    nil,
			expected: 1,
		},
		{
			name:     "not direct with one stop",
			isDirect: false,
			stops:    []AirAsiaStop{{Airport: "SOC", WaitTimeMinutes: 95}},
			expected: 1,
		},
		{
			name:     "not direct with two stops",
			isDirect: false,
			stops:    []AirAsiaStop{{Airport: "SOC", WaitTimeMinutes: 60}, {Airport: "SUB", WaitTimeMinutes: 45}},
			expected: 2,
		},
		{
			name:     "not direct with empty stops array",
			isDirect: false,
			stops:    []AirAsiaStop{},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := directFlightToStops(tt.isDirect, tt.stops)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseDateTime tests datetime parsing with various formats.
func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name        string
		datetime    string
		expectError bool
		expectedUTC string
	}{
		{
			name:        "RFC3339 with colon in offset",
			datetime:    "2025-12-15T06:00:00+07:00",
			expectError: false,
			expectedUTC: "2025-12-14T23:00:00Z",
		},
		{
			name:        "RFC3339 with UTC",
			datetime:    "2025-12-15T06:00:00Z",
			expectError: false,
			expectedUTC: "2025-12-15T06:00:00Z",
		},
		{
			name:        "with +08:00 offset",
			datetime:    "2025-12-15T08:45:00+08:00",
			expectError: false,
			expectedUTC: "2025-12-15T00:45:00Z",
		},
		{
			name:        "invalid format",
			datetime:    "15-12-2025 06:00:00",
			expectError: true,
		},
		{
			name:        "empty string",
			datetime:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.datetime)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUTC, result.UTC().Format(time.RFC3339))
			}
		})
	}
}

// TestParseBaggageNote tests baggage note parsing.
func TestParseBaggageNote(t *testing.T) {
	tests := []struct {
		name            string
		note            string
		expectedCabin   int
		expectedChecked int
	}{
		{
			name:            "cabin only with additional fee note",
			note:            "Cabin baggage only, checked bags additional fee",
			expectedCabin:   7,
			expectedChecked: 0,
		},
		{
			name:            "includes 20kg checked",
			note:            "Includes 20kg checked baggage",
			expectedCabin:   7,
			expectedChecked: 20,
		},
		{
			name:            "includes 15kg checked",
			note:            "Cabin + 15kg checked baggage included",
			expectedCabin:   7,
			expectedChecked: 15,
		},
		{
			name:            "empty note",
			note:            "",
			expectedCabin:   7,
			expectedChecked: 0,
		},
		{
			name:            "generic note",
			note:            "Standard baggage policy applies",
			expectedCabin:   7,
			expectedChecked: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cabinKg, checkedKg := parseBaggageNote(tt.note)
			assert.Equal(t, tt.expectedCabin, cabinKg)
			assert.Equal(t, tt.expectedChecked, checkedKg)
		})
	}
}

// TestNormalize_FieldMapping tests that all fields are mapped correctly.
func TestNormalize_FieldMapping(t *testing.T) {
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ520",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T04:45:00+07:00",
			ArriveTime:    "2025-12-15T07:25:00+08:00",
			DurationHours: 1.67,
			DirectFlight:  true,
			PriceIDR:      650000,
			Seats:         67,
			CabinClass:    "ECONOMY",
			BaggageNote:   "Cabin baggage only, checked bags additional fee",
		},
	}

	result := normalize(flights)

	require.Len(t, result, 1)
	f := result[0]

	// Verify all field mappings
	assert.Equal(t, "airasia-QZ520-CGK-DPS", f.ID)
	assert.Equal(t, "QZ520", f.FlightNumber)
	assert.Equal(t, "QZ", f.Airline.Code)
	assert.Equal(t, "AirAsia", f.Airline.Name)
	assert.Equal(t, "CGK", f.Departure.AirportCode)
	assert.Equal(t, "DPS", f.Arrival.AirportCode)
	assert.Equal(t, 100, f.Duration.TotalMinutes)
	assert.Equal(t, "1h 40m", f.Duration.Formatted)
	assert.Equal(t, float64(650000), f.Price.Amount)
	assert.Equal(t, "IDR", f.Price.Currency)
	assert.Equal(t, 7, f.Baggage.CabinKg)
	assert.Equal(t, 0, f.Baggage.CheckedKg)
	assert.Equal(t, "economy", f.Class) // Normalized to lowercase
	assert.Equal(t, 0, f.Stops)
	assert.Equal(t, "airasia", f.Provider)
}

// TestNormalize_SkipsInvalidFlights tests that flights with invalid data are skipped.
func TestNormalize_SkipsInvalidFlights(t *testing.T) {
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ520",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "invalid-datetime", // Invalid
			ArriveTime:    "2025-12-15T07:25:00+08:00",
			DurationHours: 1.67,
			DirectFlight:  true,
			PriceIDR:      650000,
			CabinClass:    "economy",
		},
		{
			FlightCode:    "QZ524",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T10:00:00+07:00",
			ArriveTime:    "2025-12-15T12:45:00+08:00", // Valid
			DurationHours: 1.75,
			DirectFlight:  true,
			PriceIDR:      720000,
			CabinClass:    "economy",
		},
	}

	result := normalize(flights)

	// Only the valid flight should be returned
	assert.Len(t, result, 1)
	assert.Equal(t, "QZ524", result[0].FlightNumber)
}

// TestNormalize_ConnectingFlight tests normalization of connecting flights.
func TestNormalize_ConnectingFlight(t *testing.T) {
	flights := []AirAsiaFlight{
		{
			FlightCode:    "QZ7250",
			Airline:       "AirAsia",
			FromAirport:   "CGK",
			ToAirport:     "DPS",
			DepartTime:    "2025-12-15T15:15:00+07:00",
			ArriveTime:    "2025-12-15T20:35:00+08:00",
			DurationHours: 4.33,
			DirectFlight:  false,
			Stops:         []AirAsiaStop{{Airport: "SOC", WaitTimeMinutes: 95}},
			PriceIDR:      485000,
			Seats:         88,
			CabinClass:    "economy",
			BaggageNote:   "Cabin baggage only",
		},
	}

	result := normalize(flights)

	require.Len(t, result, 1)
	f := result[0]

	assert.Equal(t, 1, f.Stops)
	assert.Equal(t, 260, f.Duration.TotalMinutes)
	assert.Equal(t, "4h 20m", f.Duration.Formatted)
}

// TestAdapter_Search_WithRealMockFile tests against the actual mock file.
func TestAdapter_Search_WithRealMockFile(t *testing.T) {
	// Try to find the real mock file
	mockPath := filepath.Join("..", "..", "..", "..", "docs", "response-mock", "airasia_search_response.json")

	// Skip if file doesn't exist
	if _, err := os.Stat(mockPath); os.IsNotExist(err) {
		t.Skip("Mock file not found, skipping integration test")
	}

	adapter := NewAdapter(mockPath)
	ctx := context.Background()

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	require.NoError(t, err)
	require.NotEmpty(t, flights)

	// Verify basic properties of all returned flights
	for _, f := range flights {
		assert.Equal(t, "airasia", f.Provider)
		assert.NotEmpty(t, f.ID)
		assert.NotEmpty(t, f.FlightNumber)
		assert.NotEmpty(t, f.Airline.Code)
		assert.NotEmpty(t, f.Departure.AirportCode)
		assert.NotEmpty(t, f.Arrival.AirportCode)
		assert.Greater(t, f.Duration.TotalMinutes, 0)
		assert.NotEmpty(t, f.Duration.Formatted)
		assert.Greater(t, f.Price.Amount, float64(0))
		assert.Equal(t, "IDR", f.Price.Currency)
	}
}

// createTempFile creates a temporary file with the given content.
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "airasia_test_*.json")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}
