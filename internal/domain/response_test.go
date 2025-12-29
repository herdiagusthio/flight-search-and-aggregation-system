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
				SearchTimeMs:       100,
				ProvidersQueried:   2,
				ProvidersSucceeded: 2,
				ProvidersFailed:    0,
			},
			wantFlightCount:  2,
			wantTotalResults: 2,
			wantNilFlights:   false,
		},
		{
			name:    "handles nil flights",
			flights: nil,
			metadata: SearchMetadata{
				SearchTimeMs:       50,
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
			},
			wantFlightCount:  0,
			wantTotalResults: 0,
			wantNilFlights:   false, // Should be converted to empty slice
		},
		{
			name:    "handles empty flights",
			flights: []Flight{},
			metadata: SearchMetadata{
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
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
				SearchTimeMs:       25,
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
			},
			wantFlightCount:  1,
			wantTotalResults: 1,
			wantNilFlights:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
				Passengers:    1,
				Class:         "economy",
			}
			response := NewSearchResponse(criteria, tt.flights, tt.metadata)

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
