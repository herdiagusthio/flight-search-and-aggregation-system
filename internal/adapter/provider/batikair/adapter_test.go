package batikair

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
	assert.Equal(t, "batik_air", adapter.Name())
}

// TestAdapter_ImplementsInterface ensures Adapter implements FlightProvider.
func TestAdapter_ImplementsInterface(t *testing.T) {
	var _ domain.FlightProvider = (*Adapter)(nil)
}

// TestAdapter_Search tests the Search method with various scenarios.
func TestAdapter_Search(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name             string
		jsonContent      string
		criteria         domain.SearchCriteria
		wantFlights      int
		wantErr          bool
		wantRetryable    bool
		checkFirstFlight func(*testing.T, domain.Flight)
	}{
		{
			name: "successful parsing with valid flights",
			jsonContent: `{
				"code": 200,
				"message": "OK",
				"results": [
					{
						"flightNumber": "ID6514",
						"airlineName": "Batik Air",
						"airlineIATA": "ID",
						"origin": "CGK",
						"destination": "DPS",
						"departureDateTime": "2025-12-15T07:15:00+0700",
						"arrivalDateTime": "2025-12-15T10:00:00+0800",
						"travelTime": "1h 45m",
						"numberOfStops": 0,
						"fare": {
							"basePrice": 980000,
							"taxes": 120000,
							"totalPrice": 1100000,
							"currencyCode": "IDR",
							"class": "Y"
						},
						"seatsAvailable": 32,
						"aircraftModel": "Airbus A320",
						"baggageInfo": "7kg cabin, 20kg checked"
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
				assert.Equal(t, "ID6514", f.ID)
				assert.Equal(t, "ID6514", f.FlightNumber)
				assert.Equal(t, "ID", f.Airline.Code)
				assert.Equal(t, "Batik Air", f.Airline.Name)
				assert.Equal(t, "CGK", f.Departure.AirportCode)
				assert.Equal(t, "DPS", f.Arrival.AirportCode)
				assert.Equal(t, 105, f.Duration.TotalMinutes)
				assert.Equal(t, "1h 45m", f.Duration.Formatted)
				assert.Equal(t, float64(1100000), f.Price.Amount)
				assert.Equal(t, "IDR", f.Price.Currency)
				assert.Equal(t, 7, f.Baggage.CabinKg)
				assert.Equal(t, 20, f.Baggage.CheckedKg)
				assert.Equal(t, "economy", f.Class)
				assert.Equal(t, 0, f.Stops)
				assert.Equal(t, "batik_air", f.Provider)
			},
		},
		{
			name: "empty flights array returns empty slice",
			jsonContent: `{
				"code": 200,
				"message": "OK",
				"results": []
			}`,
			criteria:    domain.SearchCriteria{},
			wantFlights: 0,
			wantErr:     false,
		},
		{
			name: "connecting flight with stops",
			jsonContent: `{
				"code": 200,
				"message": "OK",
				"results": [
					{
						"flightNumber": "ID7042",
						"airlineName": "Batik Air",
						"airlineIATA": "ID",
						"origin": "CGK",
						"destination": "DPS",
						"departureDateTime": "2025-12-15T18:45:00+0700",
						"arrivalDateTime": "2025-12-15T23:50:00+0800",
						"travelTime": "3h 5m",
						"numberOfStops": 1,
						"connections": [{"stopAirport": "UPG", "stopDuration": "55m"}],
						"fare": {
							"basePrice": 850000,
							"taxes": 100000,
							"totalPrice": 950000,
							"currencyCode": "IDR",
							"class": "Y"
						},
						"seatsAvailable": 41,
						"baggageInfo": "7kg cabin, 20kg checked"
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
				assert.Equal(t, 1, f.Stops)
				assert.Equal(t, 185, f.Duration.TotalMinutes) // 3h 5m = 185 minutes
				assert.Equal(t, "3h 5m", f.Duration.Formatted)
			},
		},
		{
			name: "price calculated from base + taxes when totalPrice is 0",
			jsonContent: `{
				"code": 200,
				"message": "OK",
				"results": [
					{
						"flightNumber": "ID6514",
						"airlineName": "Batik Air",
						"airlineIATA": "ID",
						"origin": "CGK",
						"destination": "DPS",
						"departureDateTime": "2025-12-15T07:15:00+0700",
						"arrivalDateTime": "2025-12-15T10:00:00+0800",
						"travelTime": "1h 45m",
						"numberOfStops": 0,
						"fare": {
							"basePrice": 980000,
							"taxes": 120000,
							"totalPrice": 0,
							"currencyCode": "IDR",
							"class": "Y"
						},
						"baggageInfo": "7kg cabin, 20kg checked"
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
				assert.Equal(t, float64(1100000), f.Price.Amount) // 980000 + 120000
			},
		},
		{
			name: "filters by origin and destination",
			jsonContent: `{
				"code": 200,
				"message": "OK",
				"results": [
					{
						"flightNumber": "ID6514",
						"airlineName": "Batik Air",
						"airlineIATA": "ID",
						"origin": "CGK",
						"destination": "DPS",
						"departureDateTime": "2025-12-15T07:15:00+0700",
						"arrivalDateTime": "2025-12-15T10:00:00+0800",
						"travelTime": "1h 45m",
						"numberOfStops": 0,
						"fare": {"totalPrice": 1100000, "currencyCode": "IDR", "class": "Y"},
						"baggageInfo": "7kg cabin, 20kg checked"
					},
					{
						"flightNumber": "ID8000",
						"airlineName": "Batik Air",
						"airlineIATA": "ID",
						"origin": "SUB",
						"destination": "DPS",
						"departureDateTime": "2025-12-15T10:00:00+0700",
						"arrivalDateTime": "2025-12-15T11:30:00+0800",
						"travelTime": "1h 30m",
						"numberOfStops": 0,
						"fare": {"totalPrice": 800000, "currencyCode": "IDR", "class": "Y"},
						"baggageInfo": "7kg cabin, 20kg checked"
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
				assert.Equal(t, "ID6514", f.ID)
			},
		},
		{
			name:          "malformed JSON returns error",
			jsonContent:   `{ invalid json }`,
			criteria:      domain.SearchCriteria{},
			wantFlights:   0,
			wantErr:       true,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	cancel()

	flights, err := adapter.Search(ctx, domain.SearchCriteria{})

	require.Error(t, err)
	assert.Empty(t, flights)

	providerErr, ok := err.(*domain.ProviderError)
	require.True(t, ok, "Error should be ProviderError")
	assert.Equal(t, ProviderName, providerErr.Provider)
	assert.Equal(t, context.Canceled, providerErr.Err)
	assert.False(t, providerErr.Retryable)
}

// TestParseDurationString tests duration string parsing.
func TestParseDurationString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMinutes int
		wantErr     bool
	}{
		{"2h 15m", "2h 15m", 135, false},
		{"1h 45m", "1h 45m", 105, false},
		{"3h 5m", "3h 5m", 185, false},
		{"1h only", "1h", 60, false},
		{"1h 0m", "1h 0m", 60, false},
		{"45m only", "45m", 45, false},
		{"0h 30m", "0h 30m", 30, false},
		{"with spaces", "  2h  15m  ", 135, false},
		{"12h 30m", "12h 30m", 750, false},
		{"0h 0m", "0h 0m", 0, false},
		{"empty string", "", 0, true},
		{"invalid format", "invalid", 0, true},
		{"no duration", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDurationString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantMinutes, result)
			}
		})
	}
}

// TestMapCabinClass tests cabin class code mapping.
func TestMapCabinClass(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Y", "economy"},
		{"y", "economy"},
		{"W", "premium_economy"},
		{"w", "premium_economy"},
		{"C", "business"},
		{"c", "business"},
		{"J", "business"},
		{"j", "business"},
		{"F", "first"},
		{"f", "first"},
		{"unknown", "economy"},
		{"", "economy"},
		{"  Y  ", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapCabinClass(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseBaggageInfo tests baggage info string parsing.
func TestParseBaggageInfo(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCabin   int
		wantChecked int
	}{
		{"standard format", "7kg cabin, 20kg checked", 7, 20},
		{"with spaces", "7 kg cabin, 20 kg checked", 7, 20},
		{"different values", "10kg cabin, 30kg checked", 10, 30},
		{"uppercase", "7KG CABIN, 20KG CHECKED", 7, 20},
		{"empty string returns defaults", "", 7, 20},
		{"only cabin", "10kg cabin", 10, 20},
		{"only checked", "25kg checked", 7, 25},
		{"no match returns defaults", "some other format", 7, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cabin, checked := parseBaggageInfo(tt.input)
			assert.Equal(t, tt.wantCabin, cabin)
			assert.Equal(t, tt.wantChecked, checked)
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
			name:    "timezone without colon",
			input:   "2025-12-15T07:15:00+0700",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 2025, tm.Year())
				assert.Equal(t, time.December, tm.Month())
				assert.Equal(t, 15, tm.Day())
				assert.Equal(t, 7, tm.Hour())
				assert.Equal(t, 15, tm.Minute())
			},
		},
		{
			name:    "RFC3339 format",
			input:   "2025-12-15T10:00:00+08:00",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 10, tm.Hour())
			},
		},
		{
			name:    "Z timezone",
			input:   "2025-12-15T12:00:00Z",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 12, tm.Hour())
			},
		},
		{
			name:    "without timezone",
			input:   "2025-12-15T14:30:00",
			wantErr: false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 14, tm.Hour())
				assert.Equal(t, 30, tm.Minute())
			},
		},
		{
			name:    "invalid format",
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

// TestNormalize_SkipsInvalidFlights tests that invalid flights are skipped.
func TestNormalize_SkipsInvalidFlights(t *testing.T) {
	flights := []BatikAirFlight{
		{
			FlightNumber:      "ID6514",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "DPS",
			DepartureDateTime: "2025-12-15T07:15:00+0700",
			ArrivalDateTime:   "2025-12-15T10:00:00+0800",
			TravelTime:        "1h 45m",
			NumberOfStops:     0,
			Fare:              BatikAirFare{TotalPrice: 1100000, CurrencyCode: "IDR", Class: "Y"},
			BaggageInfo:       "7kg cabin, 20kg checked",
		},
		{
			FlightNumber:      "ID6515",
			AirlineName:       "Batik Air",
			AirlineIATA:       "ID",
			Origin:            "CGK",
			Destination:       "DPS",
			DepartureDateTime: "invalid-date",
			ArrivalDateTime:   "2025-12-15T12:00:00+0800",
			TravelTime:        "1h 45m",
			Fare:              BatikAirFare{TotalPrice: 1200000, CurrencyCode: "IDR", Class: "Y"},
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "ID6514", result[0].ID)
}

// TestAdapter_Search_WithRealMockFile tests with the actual mock file.
func TestAdapter_Search_WithRealMockFile(t *testing.T) {
	mockPath := "../../../../docs/response-mock/batik_air_search_response.json"

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

	for _, f := range flights {
		assert.Equal(t, "batik_air", f.Provider)
		assert.NotEmpty(t, f.ID)
		assert.NotEmpty(t, f.Airline.Code)
		assert.NotEmpty(t, f.Departure.AirportCode)
		assert.NotEmpty(t, f.Arrival.AirportCode)
		assert.Greater(t, f.Duration.TotalMinutes, 0)
		assert.Greater(t, f.Price.Amount, float64(0))
	}
}
