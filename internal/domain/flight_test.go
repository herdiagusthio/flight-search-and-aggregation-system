package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDurationInfo(t *testing.T) {
	tests := []struct {
		name          string
		totalMinutes  int
		wantFormatted string
	}{
		{
			name:          "hours and minutes",
			totalMinutes:  150, // 2h 30m
			wantFormatted: "2h 30m",
		},
		{
			name:          "only hours",
			totalMinutes:  120, // 2h
			wantFormatted: "2h",
		},
		{
			name:          "only minutes",
			totalMinutes:  45,
			wantFormatted: "45m",
		},
		{
			name:          "zero minutes",
			totalMinutes:  0,
			wantFormatted: "0m",
		},
		{
			name:          "single digit minutes",
			totalMinutes:  65, // 1h 5m
			wantFormatted: "1h 5m",
		},
		{
			name:          "large duration",
			totalMinutes:  725, // 12h 5m
			wantFormatted: "12h 5m",
		},
		{
			name:          "exactly one hour",
			totalMinutes:  60,
			wantFormatted: "1h",
		},
		{
			name:          "one minute",
			totalMinutes:  1,
			wantFormatted: "1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDurationInfo(tt.totalMinutes)
			assert.Equal(t, tt.totalMinutes, result.TotalMinutes)
			assert.Equal(t, tt.wantFormatted, result.Formatted)
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "single digit", input: 1, want: "1"},
		{name: "two digits", input: 10, want: "10"},
		{name: "three digits", input: 123, want: "123"},
		{name: "four digits", input: 9999, want: "9999"},
		{name: "large number", input: 12345, want: "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, intToString(tt.input))
		})
	}
}

func TestFlight_Validate(t *testing.T) {
	// Base times for testing
	departureTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	arrivalTime := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		flight  Flight
		wantErr error
	}{
		{
			name: "valid flight with all required fields",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    arrivalTime,
				},
			},
			wantErr: nil,
		},
		{
			name: "arrival time equals departure time - should fail",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    departureTime, // Same as departure
				},
			},
			wantErr: ErrInvalidFlightTimes,
		},
		{
			name: "arrival time before departure time - should fail",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    departureTime.Add(-1 * time.Hour), // Before departure
				},
			},
			wantErr: ErrInvalidFlightTimes,
		},
		{
			name: "missing flight number - should fail",
			flight: Flight{
				FlightNumber: "", // Empty
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    arrivalTime,
				},
			},
			wantErr: ErrMissingRequiredField,
		},
		{
			name: "missing airline code - should fail",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "", // Empty
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    arrivalTime,
				},
			},
			wantErr: ErrMissingRequiredField,
		},
		{
			name: "missing departure airport code - should fail",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "", // Empty
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    arrivalTime,
				},
			},
			wantErr: ErrMissingRequiredField,
		},
		{
			name: "missing arrival airport code - should fail",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "", // Empty
					DateTime:    arrivalTime,
				},
			},
			wantErr: ErrMissingRequiredField,
		},
		{
			name: "duration mismatch - should pass (only logged as warning)",
			flight: Flight{
				FlightNumber: "GA-123",
				Airline: AirlineInfo{
					Code: "GA",
					Name: "Garuda Indonesia",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "DPS",
					DateTime:    arrivalTime, // Actual: 2h 30m
				},
				Duration: DurationInfo{
					TotalMinutes: 180, // Claimed: 3h (mismatch)
					Formatted:    "3h",
				},
			},
			wantErr: nil, // Duration mismatch doesn't fail validation
		},
		{
			name: "valid international flight with all optional fields",
			flight: Flight{
				ID:           "test-123",
				FlightNumber: "SQ-123",
				Airline: AirlineInfo{
					Code: "SQ",
					Name: "Singapore Airlines",
					Logo: "https://example.com/sq.png",
				},
				Departure: FlightPoint{
					AirportCode: "SIN",
					AirportName: "Singapore Changi Airport",
					Terminal:    "3",
					DateTime:    departureTime,
					Timezone:    "Asia/Singapore",
				},
				Arrival: FlightPoint{
					AirportCode: "JFK",
					AirportName: "John F. Kennedy International Airport",
					Terminal:    "4",
					DateTime:    arrivalTime,
					Timezone:    "America/New_York",
				},
				Duration: DurationInfo{
					TotalMinutes: 150,
					Formatted:    "2h 30m",
				},
				Price: PriceInfo{
					Amount:    1500000,
					Currency:  "IDR",
					Formatted: "IDR 1,500,000",
				},
				Baggage: BaggageInfo{
					CabinKg:   7,
					CheckedKg: 20,
				},
				Class:        "economy",
				Stops:        0,
				Provider:     "test-provider",
				RankingScore: 85.5,
			},
			wantErr: nil,
		},
		{
			name: "minimal valid flight - only required fields",
			flight: Flight{
				FlightNumber: "JT-001",
				Airline: AirlineInfo{
					Code: "JT",
				},
				Departure: FlightPoint{
					AirportCode: "CGK",
					DateTime:    departureTime,
				},
				Arrival: FlightPoint{
					AirportCode: "SUB",
					DateTime:    arrivalTime,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flight.Validate()

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr),
					"expected error to wrap %v, got %v", tt.wantErr, err)
			}
		})
	}
}
