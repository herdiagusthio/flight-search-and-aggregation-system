package lionair

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
	assert.Equal(t, "lion_air", adapter.Name())
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
				"success": true,
				"data": {
					"available_flights": [
						{
							"id": "JT740",
							"carrier": {"name": "Lion Air", "iata": "JT"},
							"route": {
								"from": {"code": "CGK", "name": "Soekarno-Hatta", "city": "Jakarta"},
								"to": {"code": "DPS", "name": "Ngurah Rai", "city": "Denpasar"}
							},
							"schedule": {
								"departure": "2025-12-15T05:30:00",
								"departure_timezone": "Asia/Jakarta",
								"arrival": "2025-12-15T08:15:00",
								"arrival_timezone": "Asia/Makassar"
							},
							"flight_time": 105,
							"is_direct": true,
							"pricing": {"total": 950000, "currency": "IDR", "fare_type": "ECONOMY"},
							"seats_left": 45,
							"plane_type": "Boeing 737-900ER",
							"services": {
								"wifi_available": false,
								"meals_included": false,
								"baggage_allowance": {"cabin": "7 kg", "hold": "20 kg"}
							}
						}
					]
				}
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, "JT740", f.ID)
				assert.Equal(t, "JT740", f.FlightNumber)
				assert.Equal(t, "JT", f.Airline.Code)
				assert.Equal(t, "Lion Air", f.Airline.Name)
				assert.Equal(t, "CGK", f.Departure.AirportCode)
				assert.Equal(t, "Soekarno-Hatta", f.Departure.AirportName)
				assert.Equal(t, "Asia/Jakarta", f.Departure.Timezone)
				assert.Equal(t, "DPS", f.Arrival.AirportCode)
				assert.Equal(t, "Asia/Makassar", f.Arrival.Timezone)
				assert.Equal(t, 105, f.Duration.TotalMinutes)
				assert.Equal(t, "1h 45m", f.Duration.Formatted)
				assert.Equal(t, float64(950000), f.Price.Amount)
				assert.Equal(t, "IDR", f.Price.Currency)
				assert.Equal(t, 7, f.Baggage.CabinKg)
				assert.Equal(t, 20, f.Baggage.CheckedKg)
				assert.Equal(t, "economy", f.Class)
				assert.Equal(t, 0, f.Stops)
				assert.Equal(t, "lion_air", f.Provider)
			},
		},
		{
			name: "empty flights array returns empty slice",
			jsonContent: `{
				"success": true,
				"data": {
					"available_flights": []
				}
			}`,
			criteria:    domain.SearchCriteria{},
			wantFlights: 0,
			wantErr:     false,
		},
		{
			name: "connecting flight with layover calculates stops correctly",
			jsonContent: `{
				"success": true,
				"data": {
					"available_flights": [
						{
							"id": "JT650",
							"carrier": {"name": "Lion Air", "iata": "JT"},
							"route": {
								"from": {"code": "CGK", "name": "Soekarno-Hatta", "city": "Jakarta"},
								"to": {"code": "DPS", "name": "Ngurah Rai", "city": "Denpasar"}
							},
							"schedule": {
								"departure": "2025-12-15T16:20:00",
								"departure_timezone": "Asia/Jakarta",
								"arrival": "2025-12-15T21:10:00",
								"arrival_timezone": "Asia/Makassar"
							},
							"flight_time": 230,
							"is_direct": false,
							"stop_count": 1,
							"layovers": [{"airport": "SUB", "duration_minutes": 75}],
							"pricing": {"total": 780000, "currency": "IDR", "fare_type": "ECONOMY"},
							"seats_left": 52,
							"plane_type": "Boeing 737-800",
							"services": {
								"wifi_available": false,
								"meals_included": false,
								"baggage_allowance": {"cabin": "7 kg", "hold": "20 kg"}
							}
						}
					]
				}
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, 1, f.Stops, "Should have 1 stop")
				assert.Equal(t, "3h 50m", f.Duration.Formatted)
			},
		},
		{
			name: "filters by origin and destination",
			jsonContent: `{
				"success": true,
				"data": {
					"available_flights": [
						{
							"id": "JT740",
							"carrier": {"name": "Lion Air", "iata": "JT"},
							"route": {
								"from": {"code": "CGK", "name": "Soekarno-Hatta", "city": "Jakarta"},
								"to": {"code": "DPS", "name": "Ngurah Rai", "city": "Denpasar"}
							},
							"schedule": {
								"departure": "2025-12-15T05:30:00",
								"departure_timezone": "Asia/Jakarta",
								"arrival": "2025-12-15T08:15:00",
								"arrival_timezone": "Asia/Makassar"
							},
							"flight_time": 105,
							"is_direct": true,
							"pricing": {"total": 950000, "currency": "IDR", "fare_type": "ECONOMY"},
							"seats_left": 45,
							"plane_type": "Boeing 737-900ER",
							"services": {"baggage_allowance": {"cabin": "7 kg", "hold": "20 kg"}}
						},
						{
							"id": "JT800",
							"carrier": {"name": "Lion Air", "iata": "JT"},
							"route": {
								"from": {"code": "SUB", "name": "Juanda", "city": "Surabaya"},
								"to": {"code": "DPS", "name": "Ngurah Rai", "city": "Denpasar"}
							},
							"schedule": {
								"departure": "2025-12-15T10:00:00",
								"departure_timezone": "Asia/Jakarta",
								"arrival": "2025-12-15T11:30:00",
								"arrival_timezone": "Asia/Makassar"
							},
							"flight_time": 90,
							"is_direct": true,
							"pricing": {"total": 600000, "currency": "IDR", "fare_type": "ECONOMY"},
							"seats_left": 30,
							"plane_type": "Boeing 737-800",
							"services": {"baggage_allowance": {"cabin": "7 kg", "hold": "20 kg"}}
						}
					]
				}
			}`,
			criteria: domain.SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			wantFlights: 1,
			wantErr:     false,
			checkFirstFlight: func(t *testing.T, f domain.Flight) {
				assert.Equal(t, "JT740", f.ID)
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

// TestParseDateTimeWithTimezone tests datetime parsing with various timezone formats.
func TestParseDateTimeWithTimezone(t *testing.T) {
	tests := []struct {
		name      string
		datetime  string
		timezone  string
		wantErr   bool
		checkTime func(*testing.T, time.Time)
	}{
		{
			name:     "valid datetime with Asia/Jakarta timezone",
			datetime: "2025-12-15T05:30:00",
			timezone: "Asia/Jakarta",
			wantErr:  false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 2025, tm.Year())
				assert.Equal(t, time.December, tm.Month())
				assert.Equal(t, 15, tm.Day())
				assert.Equal(t, 5, tm.Hour())
				assert.Equal(t, 30, tm.Minute())
				assert.Equal(t, "Asia/Jakarta", tm.Location().String())
			},
		},
		{
			name:     "datetime with space separator",
			datetime: "2025-12-15 10:00:00",
			timezone: "Asia/Makassar",
			wantErr:  false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 10, tm.Hour())
				assert.Equal(t, "Asia/Makassar", tm.Location().String())
			},
		},
		{
			name:     "invalid timezone falls back to UTC",
			datetime: "2025-12-15T12:00:00",
			timezone: "Invalid/Timezone",
			wantErr:  false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, 12, tm.Hour())
				assert.Equal(t, "UTC", tm.Location().String())
			},
		},
		{
			name:     "empty timezone falls back to UTC",
			datetime: "2025-12-15T12:00:00",
			timezone: "",
			wantErr:  false,
			checkTime: func(t *testing.T, tm time.Time) {
				assert.Equal(t, "UTC", tm.Location().String())
			},
		},
		{
			name:     "invalid datetime format",
			datetime: "not-a-date",
			timezone: "Asia/Jakarta",
			wantErr:  true,
		},
		{
			name:     "empty datetime",
			datetime: "",
			timezone: "Asia/Jakarta",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDateTimeWithTimezone(tt.datetime, tt.timezone)
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

// TestParseBaggageWeight tests baggage weight parsing.
func TestParseBaggageWeight(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7 kg", 7},
		{"20 kg", 20},
		{"7kg", 7},
		{"20KG", 20},
		{"  10  kg  ", 10},
		{"0 kg", 0},
		{"invalid", 0},
		{"", 0},
		{"kg", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBaggageWeight(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeClass tests class normalization.
func TestNormalizeClass(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ECONOMY", "economy"},
		{"economy", "economy"},
		{"Economy", "economy"},
		{"eco", "economy"},
		{"y", "economy"},
		{"BUSINESS", "business"},
		{"business", "business"},
		{"biz", "business"},
		{"j", "business"},
		{"c", "business"},
		{"FIRST", "first"},
		{"first", "first"},
		{"f", "first"},
		{"unknown", "economy"},
		{"", "economy"},
		{"  ECONOMY  ", "economy"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeClass(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalize_SkipsInvalidFlights tests that invalid flights are skipped.
func TestNormalize_SkipsInvalidFlights(t *testing.T) {
	flights := []LionAirFlight{
		{
			ID:      "JT740",
			Carrier: LionAirCarrier{Name: "Lion Air", IATA: "JT"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK", Name: "Soekarno-Hatta"},
				To:   LionAirAirport{Code: "DPS", Name: "Ngurah Rai"},
			},
			Schedule: LionAirSchedule{
				Departure:         "2025-12-15T05:30:00",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T08:15:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			FlightTime: 105,
			IsDirect:   true,
			Pricing:    LionAirPricing{Total: 950000, Currency: "IDR", FareType: "ECONOMY"},
			Services: LionAirServices{
				BaggageAllowance: LionAirBaggageAllowance{Cabin: "7 kg", Hold: "20 kg"},
			},
		},
		{
			ID:      "JT741",
			Carrier: LionAirCarrier{Name: "Lion Air", IATA: "JT"},
			Route: LionAirRoute{
				From: LionAirAirport{Code: "CGK"},
				To:   LionAirAirport{Code: "DPS"},
			},
			Schedule: LionAirSchedule{
				Departure:         "invalid-date",
				DepartureTimezone: "Asia/Jakarta",
				Arrival:           "2025-12-15T10:00:00",
				ArrivalTimezone:   "Asia/Makassar",
			},
			FlightTime: 100,
			Pricing:    LionAirPricing{Total: 850000, Currency: "IDR"},
		},
	}

	result := normalize(flights)

	assert.Len(t, result, 1)
	assert.Equal(t, "JT740", result[0].ID)
}

// TestAdapter_Search_WithRealMockFile tests with the actual mock file.
func TestAdapter_Search_WithRealMockFile(t *testing.T) {
	mockPath := "../../../../docs/response-mock/lion_air_search_response.json"

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
		assert.Equal(t, "lion_air", f.Provider)
		assert.NotEmpty(t, f.ID)
		assert.NotEmpty(t, f.Airline.Code)
		assert.NotEmpty(t, f.Departure.AirportCode)
		assert.NotEmpty(t, f.Arrival.AirportCode)
		assert.Greater(t, f.Duration.TotalMinutes, 0)
		assert.Greater(t, f.Price.Amount, float64(0))
	}
}

// TestDurationFormatting tests various duration edge cases.
func TestDurationFormatting(t *testing.T) {
	tests := []struct {
		minutes       int
		wantFormatted string
	}{
		{0, "0m"},
		{30, "30m"},
		{60, "1h"},
		{90, "1h 30m"},
		{120, "2h"},
		{135, "2h 15m"},
		{180, "3h"},
		{230, "3h 50m"},
	}

	for _, tt := range tests {
		t.Run(tt.wantFormatted, func(t *testing.T) {
			duration := domain.NewDurationInfo(tt.minutes)
			assert.Equal(t, tt.minutes, duration.TotalMinutes)
			assert.Equal(t, tt.wantFormatted, duration.Formatted)
		})
	}
}
