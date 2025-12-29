package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSearchResponse(t *testing.T) {
	tests := []struct {
		name             string
		flights          []Flight
		metadata         SearchMetadata
		wantFlightCount  int
		wantTotalResults int
		wantNilFlights   bool
	}{
		{
			name: "creates response with flights",
			flights: []Flight{
				{ID: "1", FlightNumber: "GA-123"},
				{ID: "2", FlightNumber: "JT-456"},
			},
			metadata: SearchMetadata{
				SearchDurationMs: 100,
				ProvidersQueried: []string{"garuda", "lionair"},
			},
			wantFlightCount:  2,
			wantTotalResults: 2,
			wantNilFlights:   false,
		},
		{
			name:    "handles nil flights",
			flights: nil,
			metadata: SearchMetadata{
				SearchDurationMs: 50,
			},
			wantFlightCount:  0,
			wantTotalResults: 0,
			wantNilFlights:   false, // Should be converted to empty slice
		},
		{
			name:    "handles empty flights",
			flights: []Flight{},
			metadata: SearchMetadata{
				ProvidersQueried: []string{"garuda"},
			},
			wantFlightCount:  0,
			wantTotalResults: 0,
			wantNilFlights:   false,
		},
		{
			name: "single flight",
			flights: []Flight{
				{ID: "1", FlightNumber: "GA-100"},
			},
			metadata: SearchMetadata{
				SearchDurationMs: 25,
				ProvidersQueried: []string{"garuda"},
			},
			wantFlightCount:  1,
			wantTotalResults: 1,
			wantNilFlights:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewSearchResponse(tt.flights, tt.metadata)

			assert.NotNil(t, response.Flights)
			assert.Len(t, response.Flights, tt.wantFlightCount)
			assert.Equal(t, tt.wantTotalResults, response.Metadata.TotalResults)
		})
	}
}

func TestProviderResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		result     *ProviderResult
		wantSuccess bool
	}{
		{
			name: "success when no error",
			result: &ProviderResult{
				Provider: "garuda",
				Flights:  []Flight{{ID: "1"}},
				Error:    nil,
			},
			wantSuccess: true,
		},
		{
			name: "failure when error present",
			result: &ProviderResult{
				Provider: "garuda",
				Error:    ErrProviderTimeout,
			},
			wantSuccess: false,
		},
		{
			name: "success with empty flights",
			result: &ProviderResult{
				Provider: "lionair",
				Flights:  []Flight{},
				Error:    nil,
			},
			wantSuccess: true,
		},
		{
			name: "failure with unavailable error",
			result: &ProviderResult{
				Provider: "airasia",
				Error:    ErrProviderUnavailable,
			},
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantSuccess, tt.result.IsSuccess())
		})
	}
}
