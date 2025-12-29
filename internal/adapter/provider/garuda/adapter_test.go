package garuda

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

// TestAdapter_Name tests the Name method.
func TestAdapter_Name(t *testing.T) {
	adapter := NewAdapter("")
	assert.Equal(t, "garuda_indonesia", adapter.Name())
}

// TestAdapter_ImplementsInterface ensures Adapter implements FlightProvider.
func TestAdapter_ImplementsInterface(t *testing.T) {
	var _ domain.FlightProvider = (*Adapter)(nil)
}

// TestAdapter_Search tests the Search method with various scenarios.
func TestAdapter_Search(t *testing.T) {
	// Create a temporary directory for test fixtures
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		jsonContent    string
		criteria       domain.SearchCriteria
		wantFlights    int
		wantErr        bool
		wantRetryable  bool
		checkFirstFlight func(*testing.T, domain.Flight)
	}{
		{
			name: "successful parsing with valid flights",
			jsonContent: `{
				"status": "success",
				"flights": [
					{
						"flight_id": "GA400",
						"airline": "Garuda Indonesia",
						"airline_code": "GA",
						"departure": {
							"airport": "CGK",
							"city": "Jakarta",
							"time": "2025-12-15T06:00:00+07:00",
							"terminal": "3"
						},
						"arrival": {
							"airport": "DPS",
							"city": "Denpasar",
							"time": "2025-12-15T08:50:00+08:00",
							"terminal": "I"
						},
						"duration_minutes": 110,
						"stops": 0,
						"aircraft": "Boeing 737-800",
						"price": {
							"amount": 1250000,
							"currency": "IDR"
						},
						"available_seats": 28,
						"fare_class": "economy",
						"baggage": {
							"carry_on": 1,
							"checked": 2
						}
					}
				]
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, "GA400", f.ID)
				assert.Equal(t, "GA400", f.FlightNumber)
				assert.Equal(t, "GA", f.Airline.Code)
				assert.Equal(t, "Garuda Indonesia", f.Airline.Name)
				assert.Equal(t, "CGK", f.Departure.AirportCode)
				assert.Equal(t, "3", f.Departure.Terminal)
				assert.Equal(t, "DPS", f.Arrival.AirportCode)
				assert.Equal(t, 110, f.Duration.TotalMinutes)
				assert.Equal(t, "1h 50m", f.Duration.Formatted)
				assert.Equal(t, float64(1250000), f.Price.Amount)
				assert.Equal(t, "IDR", f.Price.Currency)
				assert.Equal(t, 7, f.Baggage.CabinKg)
				assert.Equal(t, 40, f.Baggage.CheckedKg)
				assert.Equal(t, "economy", f.Class)
				assert.Equal(t, 0, f.Stops)
				assert.Equal(t, "garuda_indonesia", f.Provider)
			},
		},
		{
			name: "empty flights array returns empty slice",
			jsonContent: `{
				"status": "success",
				"flights": []
			}`,
			criteria:    domain.SearchCriteria{},
			wantFlights: 0,
			wantErr:     false,
		},
		{
			name: "multi-segment flight calculates stops correctly",
			jsonContent: `{
				"status": "success",
				"flights": [
					{
						"flight_id": "GA315",
						"airline": "Garuda Indonesia",
						"airline_code": "GA",
						"departure": {
							"airport": "CGK",
							"city": "Jakarta",
							"time": "2025-12-15T14:00:00+07:00",
							"terminal": "3"
						},
						"arrival": {
							"airport": "DPS",
							"city": "Denpasar",
							"time": "2025-12-15T18:45:00+08:00",
							"terminal": "I"
						},
						"duration_minutes": 285,
						"stops": 0,
						"price": {"amount": 1850000, "currency": "IDR"},
						"fare_class": "economy",
						"baggage": {"carry_on": 1, "checked": 2},
						"segments": [
							{
								"flight_number": "GA315",
								"departure": {"airport": "CGK", "time": "2025-12-15T14:00:00+07:00"},
								"arrival": {"airport": "SUB", "time": "2025-12-15T15:30:00+07:00"},
								"duration_minutes": 90
							},
							{
								"flight_number": "GA332",
								"departure": {"airport": "SUB", "time": "2025-12-15T17:15:00+07:00"},
								"arrival": {"airport": "DPS", "time": "2025-12-15T18:45:00+08:00"},
								"duration_minutes": 90,
								"layover_minutes": 105
							}
						]
					}
				]
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, 1, f.Stops, "Should calculate stops from segments")
			},
		},
		{
			name: "filters by origin and destination",
			jsonContent: `{
				"status": "success",
				"flights": [
					{
						"flight_id": "GA400",
						"airline": "Garuda Indonesia",
						"airline_code": "GA",
						"departure": {"airport": "CGK", "city": "Jakarta", "time": "2025-12-15T06:00:00+07:00"},
						"arrival": {"airport": "DPS", "city": "Denpasar", "time": "2025-12-15T08:50:00+08:00"},
						"duration_minutes": 110,
						"stops": 0,
						"price": {"amount": 1250000, "currency": "IDR"},
						"fare_class": "economy",
						"baggage": {"carry_on": 1, "checked": 2}
					},
					{
						"flight_id": "GA500",
						"airline": "Garuda Indonesia",
						"airline_code": "GA",
						"departure": {"airport": "SUB", "city": "Surabaya", "time": "2025-12-15T10:00:00+07:00"},
						"arrival": {"airport": "DPS", "city": "Denpasar", "time": "2025-12-15T11:30:00+08:00"},
						"duration_minutes": 90,
						"stops": 0,
						"price": {"amount": 900000, "currency": "IDR"},
						"fare_class": "economy",
						"baggage": {"carry_on": 1, "checked": 2}
					}
				]
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, "GA400", f.ID)
			},
		},
		{
			name:        "malformed JSON returns error",
			jsonContent: `{ invalid json }`,
			criteria:    domain.SearchCriteria{},
			wantFlights: 0,
			wantErr:     true,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test fixture
			mockPath := filepath.Join(tempDir, tt.name+".json")
			err := os.WriteFile(mockPath, []byte(tt.jsonContent), 0644)
			require.NoError(t, err)

			adapter := NewAdapter(mockPath)
			flights, err := adapter.Search(context.Background(), tt.criteria)

			if tt.wantErr {
				require.Error(t, err)
				providerErr, ok := err.(*domain.ProviderError)
				require.True(t, ok, "Error should be ProviderError")
				assert.Equal(t, ProviderName, providerErr.Provider)
				assert.Equal(t, tt.wantRetryable, providerErr.Retryable)
			} else {
				require.NoError(t, err)
				assert.Len(t, flights, tt.wantFlights)
				if tt.checkFirstFlight != nil && len(flights) > 0 {
					tt.checkFirstFlight(t, flights[0])
				}
			}
		})
	}
}

// TestAdapter_Search_FileNotFound tests error handling for missing files.
func TestAdapter_Search_FileNotFound(t *testing.T) {
	adapter := NewAdapter("/nonexistent/path/to/file.json")
	flights, err := adapter.Search(context.Background(), domain.SearchCriteria{})

	require.Error(t, err)
	assert.Empty(t, flights)

	providerErr, ok := err.(*domain.ProviderError)
	require.True(t, ok, "Error should be ProviderError")
	assert.Equal(t, ProviderName, providerErr.Provider)
	assert.True(t, providerErr.Retryable, "File read errors should be retryable")
}

// TestAdapter_Search_ContextCancellation tests context cancellation handling.
func TestAdapter_Search_ContextCancellation(t *testing.T) {
	adapter := NewAdapter("")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	require.Error(t, err)
	assert.Empty(t, flights)

	providerErr, ok := err.(*domain.ProviderError)
	require.True(t, ok, "Error should be ProviderError")
	assert.Equal(t, ProviderName, providerErr.Provider)
	assert.Equal(t, context.Canceled, providerErr.Err)
	assert.False(t, providerErr.Retryable, "Context cancellation should not be retryable")
}

// TestNormalizeClass tests class normalization.
func TestNormalizeClass(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"economy", "economy"},
		{"Economy", "economy"},
		{"ECONOMY", "economy"},
		{"eco", "economy"},
		{"y", "economy"},
		{"business", "business"},
		{"Business", "business"},
		{"biz", "business"},
		{"j", "business"},
		{"c", "business"},
		{"first", "first"},
		{"First", "first"},
		{"f", "first"},
		{"unknown", "economy"},
		{"", "economy"},
		{"  economy  ", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeClass(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseDateTime tests datetime parsing.
func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkTime func(*testing.T, time.Time)
	}{
		{
			name:    "RFC3339 with positive timezone",
			input:   "2025-12-15T06:00:00+07:00",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 2025, tm.Year())
				assert.Equal(t, time.December, tm.Month())
				assert.Equal(t, 15, tm.Day())
				assert.Equal(t, 6, tm.Hour())
				assert.Equal(t, 0, tm.Minute())
			},
		},
		{
			name:    "RFC3339 with Z timezone",
			input:   "2025-12-15T10:30:00Z",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 10, tm.Hour())
				assert.Equal(t, 30, tm.Minute())
			},
		},
		{
			name:    "datetime without timezone",
			input:   "2025-12-15T14:00:00",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 14, tm.Hour())
			},
		},
		{
			name:    "invalid datetime format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkTime != nil {
					tt.checkTime(t, result)
				}
			}
		})
	}
}

// TestFormatAirportName tests airport name formatting.
func TestFormatAirportName(t *testing.T) {
	tests := []struct {
		code     string
		city     string
		expected string
	}{
		{"CGK", "Jakarta", "Jakarta (CGK)"},
		{"DPS", "Denpasar", "Denpasar (DPS)"},
		{"CGK", "", "CGK"},
		{"", "Jakarta", "Jakarta ()"},
	}

	for _, tt := range tests {
		t.Run(tt.code+"_"+tt.city, func(t *testing.T) {
			result := formatAirportName(tt.code, tt.city)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalize_SkipsInvalidFlights tests that invalid flights are skipped.
func TestNormalize_SkipsInvalidFlights(t *testing.T) {
	flights := []GarudaFlight{
		{
			FlightID:    "GA400",
			Airline:     "Garuda Indonesia",
			AirlineCode: "GA",
			Departure:   GarudaEndpoint{Airport: "CGK", City: "Jakarta", Time: "2025-12-15T06:00:00+07:00"},
			Arrival:     GarudaEndpoint{Airport: "DPS", City: "Denpasar", Time: "2025-12-15T08:50:00+08:00"},
			DurationMinutes: 110,
			Price:       GarudaPrice{Amount: 1250000, Currency: "IDR"},
			FareClass:   "economy",
			Baggage:     GarudaBaggage{CarryOn: 1, Checked: 2},
		},
		{
			FlightID:    "GA401",
			Airline:     "Garuda Indonesia",
			AirlineCode: "GA",
			Departure:   GarudaEndpoint{Airport: "CGK", City: "Jakarta", Time: "invalid-date"},
			Arrival:     GarudaEndpoint{Airport: "DPS", City: "Denpasar", Time: "2025-12-15T10:00:00+08:00"},
			DurationMinutes: 120,
			Price:       GarudaPrice{Amount: 1300000, Currency: "IDR"},
			FareClass:   "economy",
			Baggage:     GarudaBaggage{CarryOn: 1, Checked: 2},
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "GA400", result[0].ID)
}

// TestAdapter_Search_WithRealMockFile tests with the actual mock file.
func TestAdapter_Search_WithRealMockFile(t *testing.T) {
	// Path to the actual mock file
	mockPath := "../../../../docs/response-mock/garuda_indonesia_search_response.json"
	
	// Skip if file doesn't exist (e.g., running in isolation)
	if _, err := os.Stat(mockPath); os.IsNotExist(err) {
		t.Skip("Mock file not found, skipping integration test")
	}

	adapter := NewAdapter(mockPath)
	flights, err := adapter.Search(context.Background(), domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, flights)

	// Verify all flights have the correct provider
	for _, f := range flights {
		assert.Equal(t, "garuda_indonesia", f.Provider)
		assert.NotEmpty(t, f.ID)
		assert.NotEmpty(t, f.Airline.Code)
		assert.NotEmpty(t, f.Departure.AirportCode)
		assert.NotEmpty(t, f.Arrival.AirportCode)
		assert.Greater(t, f.Duration.TotalMinutes, 0)
		assert.Greater(t, f.Price.Amount, float64(0))
	}
}
