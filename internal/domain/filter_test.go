package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSortOption_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		option SortOption
		want   bool
	}{
		{name: "best value is valid", option: SortByBestValue, want: true},
		{name: "price is valid", option: SortByPrice, want: true},
		{name: "duration is valid", option: SortByDuration, want: true},
		{name: "departure is valid", option: SortByDeparture, want: true},
		{name: "invalid option", option: SortOption("invalid"), want: false},
		{name: "empty option", option: SortOption(""), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.option.IsValid())
		})
	}
}

func TestParseSortOption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SortOption
	}{
		{name: "parse best", input: "best", expected: SortByBestValue},
		{name: "parse price", input: "price", expected: SortByPrice},
		{name: "parse duration", input: "duration", expected: SortByDuration},
		{name: "parse departure", input: "departure", expected: SortByDeparture},
		{name: "invalid defaults to best", input: "invalid", expected: SortByBestValue},
		{name: "empty defaults to best", input: "", expected: SortByBestValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseSortOption(tt.input))
		})
	}
}

func TestTimeRange_Contains(t *testing.T) {
	// Create a time range from 08:00 to 12:00
	startTime := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	tr := &TimeRange{Start: startTime, End: endTime}

	tests := []struct {
		name      string
		timeRange *TimeRange
		testTime  time.Time
		want      bool
	}{
		{
			name:      "time within range",
			timeRange: tr,
			testTime:  time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "time at start boundary",
			timeRange: tr,
			testTime:  time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "time at end boundary",
			timeRange: tr,
			testTime:  time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "time before range",
			timeRange: tr,
			testTime:  time.Date(2025, 6, 15, 7, 30, 0, 0, time.UTC),
			want:      false,
		},
		{
			name:      "time after range",
			timeRange: tr,
			testTime:  time.Date(2025, 6, 15, 14, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name:      "nil time range always contains",
			timeRange: nil,
			testTime:  time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.timeRange.Contains(tt.testTime))
		})
	}
}

func TestFilterOptions_MatchesFlight(t *testing.T) {
	baseFlight := Flight{
		ID:           "test-1",
		FlightNumber: "GA-123",
		Airline:      AirlineInfo{Code: "GA", Name: "Garuda Indonesia"},
		Price:        PriceInfo{Amount: 1500000, Currency: "IDR"},
		Stops:        1,
		Departure: FlightPoint{
			AirportCode: "CGK",
			DateTime:    time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		},
	}

	// Helper functions for creating pointers
	floatPtr := func(f float64) *float64 { return &f }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name   string
		filter *FilterOptions
		flight Flight
		want   bool
	}{
		{
			name:   "nil filter matches all",
			filter: nil,
			flight: baseFlight,
			want:   true,
		},
		{
			name:   "empty filter matches all",
			filter: &FilterOptions{},
			flight: baseFlight,
			want:   true,
		},
		{
			name:   "max price filter passes when under limit",
			filter: &FilterOptions{MaxPrice: floatPtr(2000000.0)},
			flight: baseFlight,
			want:   true,
		},
		{
			name:   "max price filter fails when over limit",
			filter: &FilterOptions{MaxPrice: floatPtr(1000000.0)},
			flight: baseFlight,
			want:   false,
		},
		{
			name:   "max stops filter passes when equal",
			filter: &FilterOptions{MaxStops: intPtr(1)},
			flight: baseFlight,
			want:   true,
		},
		{
			name:   "max stops filter fails when exceeded",
			filter: &FilterOptions{MaxStops: intPtr(0)},
			flight: baseFlight,
			want:   false,
		},
		{
			name:   "airline filter passes when matched",
			filter: &FilterOptions{Airlines: []string{"GA", "JT"}},
			flight: baseFlight,
			want:   true,
		},
		{
			name:   "airline filter fails when not matched",
			filter: &FilterOptions{Airlines: []string{"JT", "ID"}},
			flight: baseFlight,
			want:   false,
		},
		{
			name: "departure time filter passes when in range",
			filter: &FilterOptions{
				DepartureTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			flight: baseFlight,
			want:   true,
		},
		{
			name: "departure time filter fails when out of range",
			filter: &FilterOptions{
				DepartureTimeRange: &TimeRange{
					Start: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC),
					End:   time.Date(2025, 1, 1, 18, 0, 0, 0, time.UTC),
				},
			},
			flight: baseFlight,
			want:   false,
		},
		{
			name: "multiple filters all pass",
			filter: &FilterOptions{
				MaxPrice: floatPtr(2000000.0),
				MaxStops: intPtr(2),
				Airlines: []string{"GA"},
			},
			flight: baseFlight,
			want:   true,
		},
		{
			name: "multiple filters one fails",
			filter: &FilterOptions{
				MaxPrice: floatPtr(2000000.0),
				MaxStops: intPtr(0), // Flight has 1 stop
			},
			flight: baseFlight,
			want:   false,
		},
		{
			name:   "airline filter case insensitive - lowercase matches uppercase",
			filter: &FilterOptions{Airlines: []string{"ga"}},
			flight: baseFlight,
			want:   true,
		},
		{
			name: "airline filter case insensitive - mixed case in list",
			filter: &FilterOptions{Airlines: []string{"Ga", "jT"}},
			flight: baseFlight,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.MatchesFlight(tt.flight))
		})
	}
}

func TestDurationRange_IsValid(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name string
		dr   *DurationRange
		want bool
	}{
		{
			name: "nil duration range is valid",
			dr:   nil,
			want: true,
		},
		{
			name: "valid range with both min and max",
			dr:   &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			want: true,
		},
		{
			name: "valid range with only min",
			dr:   &DurationRange{MinMinutes: intPtr(60)},
			want: true,
		},
		{
			name: "valid range with only max",
			dr:   &DurationRange{MaxMinutes: intPtr(180)},
			want: true,
		},
		{
			name: "valid range with zero min",
			dr:   &DurationRange{MinMinutes: intPtr(0), MaxMinutes: intPtr(180)},
			want: true,
		},
		{
			name: "invalid range - negative min",
			dr:   &DurationRange{MinMinutes: intPtr(-10), MaxMinutes: intPtr(180)},
			want: false,
		},
		{
			name: "invalid range - negative max",
			dr:   &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(-10)},
			want: false,
		},
		{
			name: "invalid range - min greater than max",
			dr:   &DurationRange{MinMinutes: intPtr(200), MaxMinutes: intPtr(100)},
			want: false,
		},
		{
			name: "valid range - min equals max",
			dr:   &DurationRange{MinMinutes: intPtr(120), MaxMinutes: intPtr(120)},
			want: true,
		},
		{
			name: "empty duration range (both nil) is valid",
			dr:   &DurationRange{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.dr.IsValid())
		})
	}
}

func TestDurationRange_Contains(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name            string
		dr              *DurationRange
		durationMinutes int
		want            bool
	}{
		{
			name:            "nil duration range contains all",
			dr:              nil,
			durationMinutes: 100,
			want:            true,
		},
		{
			name:            "duration within range (both bounds)",
			dr:              &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			durationMinutes: 120,
			want:            true,
		},
		{
			name:            "duration at min boundary",
			dr:              &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			durationMinutes: 60,
			want:            true,
		},
		{
			name:            "duration at max boundary",
			dr:              &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			durationMinutes: 180,
			want:            true,
		},
		{
			name:            "duration below min",
			dr:              &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			durationMinutes: 30,
			want:            false,
		},
		{
			name:            "duration above max",
			dr:              &DurationRange{MinMinutes: intPtr(60), MaxMinutes: intPtr(180)},
			durationMinutes: 200,
			want:            false,
		},
		{
			name:            "only min specified - duration above min",
			dr:              &DurationRange{MinMinutes: intPtr(60)},
			durationMinutes: 120,
			want:            true,
		},
		{
			name:            "only min specified - duration below min",
			dr:              &DurationRange{MinMinutes: intPtr(60)},
			durationMinutes: 30,
			want:            false,
		},
		{
			name:            "only max specified - duration below max",
			dr:              &DurationRange{MaxMinutes: intPtr(180)},
			durationMinutes: 120,
			want:            true,
		},
		{
			name:            "only max specified - duration above max",
			dr:              &DurationRange{MaxMinutes: intPtr(180)},
			durationMinutes: 200,
			want:            false,
		},
		{
			name:            "zero duration with zero min",
			dr:              &DurationRange{MinMinutes: intPtr(0), MaxMinutes: intPtr(180)},
			durationMinutes: 0,
			want:            true,
		},
		{
			name:            "empty duration range (no bounds) contains all",
			dr:              &DurationRange{},
			durationMinutes: 500,
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.dr.Contains(tt.durationMinutes))
		})
	}
}
