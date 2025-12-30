package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsValidTimeFormat tests the time format validation function.
func TestIsValidTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		// Valid formats
		{name: "valid morning time", timeStr: "08:00", expected: true},
		{name: "valid noon time", timeStr: "12:00", expected: true},
		{name: "valid evening time", timeStr: "18:30", expected: true},
		{name: "valid midnight", timeStr: "00:00", expected: true},
		{name: "valid end of day", timeStr: "23:59", expected: true},
		{name: "valid single digit minute", timeStr: "10:05", expected: true},

		// Invalid hours
		{name: "hour too high", timeStr: "24:00", expected: false},
		{name: "hour way too high", timeStr: "25:00", expected: false},
		{name: "hour negative", timeStr: "-01:00", expected: false},

		// Invalid minutes
		{name: "minute too high", timeStr: "12:60", expected: false},
		{name: "minute way too high", timeStr: "12:99", expected: false},
		{name: "minute negative", timeStr: "12:-01", expected: false},

		// Invalid formats
		{name: "missing colon", timeStr: "1200", expected: false},
		{name: "single digit hour", timeStr: "8:00", expected: false},
		{name: "single digit minute", timeStr: "08:0", expected: false},
		{name: "empty string", timeStr: "", expected: false},
		{name: "only hour", timeStr: "12", expected: false},
		{name: "only minute", timeStr: ":30", expected: false},
		{name: "text", timeStr: "noon", expected: false},
		{name: "wrong separator", timeStr: "12-30", expected: false},
		{name: "too many parts", timeStr: "12:30:00", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTimeFormat(tt.timeStr)
			assert.Equal(t, tt.expected, result, "isValidTimeFormat(%q) should be %v", tt.timeStr, tt.expected)
		})
	}
}

// TestValidateArrivalTimeRange tests arrival time range validation.
func TestValidateArrivalTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		timeRange     *TimeRangeDTO
		expectedError bool
		errorFields   []string
	}{
		{
			name: "valid time range",
			timeRange: &TimeRangeDTO{
				Start: "08:00",
				End:   "17:00",
			},
			expectedError: false,
		},
		{
			name: "missing start time",
			timeRange: &TimeRangeDTO{
				End: "17:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.arrivalTimeRange.start"},
		},
		{
			name: "missing end time",
			timeRange: &TimeRangeDTO{
				Start: "08:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.arrivalTimeRange.end"},
		},
		{
			name: "invalid start format",
			timeRange: &TimeRangeDTO{
				Start: "25:00",
				End:   "17:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.arrivalTimeRange.start"},
		},
		{
			name: "invalid end format",
			timeRange: &TimeRangeDTO{
				Start: "08:00",
				End:   "12:99",
			},
			expectedError: true,
			errorFields:   []string{"filters.arrivalTimeRange.end"},
		},
		{
			name: "both times invalid",
			timeRange: &TimeRangeDTO{
				Start: "25:00",
				End:   "30:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.arrivalTimeRange.start", "filters.arrivalTimeRange.end"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SearchFlightsRequest{
				Filters: &FilterDTO{
					ArrivalTimeRange: tt.timeRange,
				},
			}

			errs := &ValidationErrors{}
			req.validateArrivalTimeRange(errs)

			if tt.expectedError {
				assert.True(t, errs.HasErrors(), "Expected validation errors")
				// Check that errors contain expected fields
				errorMap := make(map[string]bool)
				for _, err := range errs.Errors {
					errorMap[err.Field] = true
				}
				for _, field := range tt.errorFields {
					assert.True(t, errorMap[field], "Expected error for field %s", field)
				}
			} else {
				assert.False(t, errs.HasErrors(), "Expected no validation errors")
			}
		})
	}
}

// TestValidateDepartureTimeRange tests departure time range validation.
func TestValidateDepartureTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		timeRange     *TimeRangeDTO
		expectedError bool
		errorFields   []string
	}{
		{
			name: "valid time range",
			timeRange: &TimeRangeDTO{
				Start: "06:00",
				End:   "12:00",
			},
			expectedError: false,
		},
		{
			name: "missing start time",
			timeRange: &TimeRangeDTO{
				End: "12:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.departureTimeRange.start"},
		},
		{
			name: "invalid start hour",
			timeRange: &TimeRangeDTO{
				Start: "24:00",
				End:   "12:00",
			},
			expectedError: true,
			errorFields:   []string{"filters.departureTimeRange.start"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SearchFlightsRequest{
				Filters: &FilterDTO{
					DepartureTimeRange: tt.timeRange,
				},
			}

			errs := &ValidationErrors{}
			req.validateDepartureTimeRange(errs)

			if tt.expectedError {
				assert.True(t, errs.HasErrors(), "Expected validation errors")
			} else {
				assert.False(t, errs.HasErrors(), "Expected no validation errors")
			}
		})
	}
}

// TestValidateDurationRange tests duration range validation.
func TestValidateDurationRange(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name          string
		durationRange *DurationRangeDTO
		expectedError bool
		errorFields   []string
	}{
		{
			name: "valid duration range",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(60),
				MaxMinutes: intPtr(180),
			},
			expectedError: false,
		},
		{
			name: "valid min only",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(60),
			},
			expectedError: false,
		},
		{
			name: "valid max only",
			durationRange: &DurationRangeDTO{
				MaxMinutes: intPtr(180),
			},
			expectedError: false,
		},
		{
			name: "negative min",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(-10),
				MaxMinutes: intPtr(180),
			},
			expectedError: true,
			errorFields:   []string{"filters.durationRange.minMinutes"},
		},
		{
			name: "negative max",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(60),
				MaxMinutes: intPtr(-10),
			},
			expectedError: true,
			errorFields:   []string{"filters.durationRange.maxMinutes"},
		},
		{
			name: "min greater than max",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(300),
				MaxMinutes: intPtr(100),
			},
			expectedError: true,
			errorFields:   []string{"filters.durationRange"},
		},
		{
			name: "zero values valid",
			durationRange: &DurationRangeDTO{
				MinMinutes: intPtr(0),
				MaxMinutes: intPtr(0),
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SearchFlightsRequest{
				Filters: &FilterDTO{
					DurationRange: tt.durationRange,
				},
			}

			errs := &ValidationErrors{}
			req.validateDurationRange(errs)

			if tt.expectedError {
				assert.True(t, errs.HasErrors(), "Expected validation errors")
				// Check that errors contain expected fields
				errorMap := make(map[string]bool)
				for _, err := range errs.Errors {
					errorMap[err.Field] = true
				}
				for _, field := range tt.errorFields {
					assert.True(t, errorMap[field], "Expected error for field %s", field)
				}
			} else {
				assert.False(t, errs.HasErrors(), "Expected no validation errors")
			}
		})
	}
}

// TestValidateFilters tests the main filter validation function.
func TestValidateFilters(t *testing.T) {
	intPtr := func(i int) *int { return &i }
	floatPtr := func(f float64) *float64 { return &f }

	tests := []struct {
		name          string
		filters       *FilterDTO
		expectedError bool
		errorCount    int
	}{
		{
			name:          "nil filters is valid",
			filters:       nil,
			expectedError: false,
		},
		{
			name:          "empty filters is valid",
			filters:       &FilterDTO{},
			expectedError: false,
		},
		{
			name: "all valid filters",
			filters: &FilterDTO{
				MaxPrice: floatPtr(1000000),
				MaxStops: intPtr(1),
				Airlines: []string{"GA", "JT"},
				DepartureTimeRange: &TimeRangeDTO{
					Start: "06:00",
					End:   "12:00",
				},
				ArrivalTimeRange: &TimeRangeDTO{
					Start: "08:00",
					End:   "17:00",
				},
				DurationRange: &DurationRangeDTO{
					MinMinutes: intPtr(60),
					MaxMinutes: intPtr(180),
				},
			},
			expectedError: false,
		},
		{
			name: "negative max price",
			filters: &FilterDTO{
				MaxPrice: floatPtr(-1000),
			},
			expectedError: true,
			errorCount:    1,
		},
		{
			name: "negative max stops",
			filters: &FilterDTO{
				MaxStops: intPtr(-1),
			},
			expectedError: true,
			errorCount:    1,
		},
		{
			name: "invalid airline code - too short",
			filters: &FilterDTO{
				Airlines: []string{"G"},
			},
			expectedError: true,
			errorCount:    1,
		},
		{
			name: "multiple validation errors",
			filters: &FilterDTO{
				MaxPrice: floatPtr(-1000),
				MaxStops: intPtr(-1),
				DepartureTimeRange: &TimeRangeDTO{
					Start: "25:00",
					End:   "12:00",
				},
			},
			expectedError: true,
			errorCount:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SearchFlightsRequest{
				Filters: tt.filters,
			}

			errs := &ValidationErrors{}
			req.validateFilters(errs)

			if tt.expectedError {
				assert.True(t, errs.HasErrors(), "Expected validation errors")
				assert.GreaterOrEqual(t, len(errs.Errors), tt.errorCount, "Expected at least %d errors", tt.errorCount)
			} else {
				assert.False(t, errs.HasErrors(), "Expected no validation errors")
			}
		})
	}
}

// TestValidationErrorsError tests the Error() method.
func TestValidationErrorsError(t *testing.T) {
	errs := &ValidationErrors{}
	errs.Add("field1", "error1")
	errs.Add("field2", "error2")

	errorMsg := errs.Error()
	require.NotEmpty(t, errorMsg)
	// Error() returns the first error's message
	assert.Equal(t, "error1", errorMsg)

	// Test empty errors
	emptyErrs := &ValidationErrors{}
	assert.Equal(t, "validation failed", emptyErrs.Error())
}
