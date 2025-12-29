package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCriteria_Validate(t *testing.T) {
	// Helper to create a valid base criteria
	validCriteria := func() *SearchCriteria {
		return &SearchCriteria{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: time.Now().AddDate(0, 0, 7).Format("2006-01-02"), // 7 days from now
			Passengers:    1,
			Class:         "economy",
		}
	}

	tests := []struct {
		name         string
		modify       func(*SearchCriteria)
		wantErr      bool
		errContains  string
		isInvalidReq bool
	}{
		{
			name:    "valid criteria passes",
			modify:  func(c *SearchCriteria) {},
			wantErr: false,
		},
		{
			name:         "empty origin fails",
			modify:       func(c *SearchCriteria) { c.Origin = "" },
			wantErr:      true,
			errContains:  "origin is required",
			isInvalidReq: true,
		},
		{
			name:         "invalid origin format fails",
			modify:       func(c *SearchCriteria) { c.Origin = "JFK1" },
			wantErr:      true,
			errContains:  "IATA code",
			isInvalidReq: true,
		},
		{
			name:         "lowercase origin fails",
			modify:       func(c *SearchCriteria) { c.Origin = "cgk" },
			wantErr:      true,
			isInvalidReq: true,
		},
		{
			name:         "empty destination fails",
			modify:       func(c *SearchCriteria) { c.Destination = "" },
			wantErr:      true,
			errContains:  "destination is required",
			isInvalidReq: true,
		},
		{
			name:         "same origin and destination fails",
			modify:       func(c *SearchCriteria) { c.Destination = c.Origin },
			wantErr:      true,
			errContains:  "must be different",
			isInvalidReq: true,
		},
		{
			name:         "empty departure date fails",
			modify:       func(c *SearchCriteria) { c.DepartureDate = "" },
			wantErr:      true,
			errContains:  "departureDate is required",
			isInvalidReq: true,
		},
		{
			name:         "invalid date format fails",
			modify:       func(c *SearchCriteria) { c.DepartureDate = "01-15-2025" },
			wantErr:      true,
			errContains:  "YYYY-MM-DD format",
			isInvalidReq: true,
		},
		{
			name:    "past date now allowed",
			modify:  func(c *SearchCriteria) { c.DepartureDate = "2020-01-01" },
			wantErr: false,
		},
		{
			name:    "today's date passes",
			modify:  func(c *SearchCriteria) { c.DepartureDate = time.Now().Format("2006-01-02") },
			wantErr: false,
		},
		{
			name:         "zero passengers fails",
			modify:       func(c *SearchCriteria) { c.Passengers = 0 },
			wantErr:      true,
			errContains:  "at least 1",
			isInvalidReq: true,
		},
		{
			name:         "too many passengers fails",
			modify:       func(c *SearchCriteria) { c.Passengers = 10 },
			wantErr:      true,
			errContains:  "cannot exceed 9",
			isInvalidReq: true,
		},
		{
			name:         "invalid class fails",
			modify:       func(c *SearchCriteria) { c.Class = "premium" },
			wantErr:      true,
			errContains:  "economy, business, first",
			isInvalidReq: true,
		},
		{
			name:    "empty class passes",
			modify:  func(c *SearchCriteria) { c.Class = "" },
			wantErr: false,
		},
		{
			name:    "business class passes",
			modify:  func(c *SearchCriteria) { c.Class = "business" },
			wantErr: false,
		},
		{
			name:    "first class passes",
			modify:  func(c *SearchCriteria) { c.Class = "first" },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := validCriteria()
			tt.modify(criteria)

			err := criteria.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				if tt.isInvalidReq {
					assert.True(t, errors.Is(err, ErrInvalidRequest))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchCriteria_SetDefaults(t *testing.T) {
	tests := []struct {
		name           string
		initial        *SearchCriteria
		wantPassengers int
		wantClass      string
	}{
		{
			name: "sets default passengers when zero",
			initial: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-06-01",
				Passengers:    0,
				Class:         "",
			},
			wantPassengers: 1,
			wantClass:      "economy",
		},
		{
			name: "sets default class when empty",
			initial: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-06-01",
				Passengers:    2,
				Class:         "",
			},
			wantPassengers: 2,
			wantClass:      "economy",
		},
		{
			name: "does not override existing passengers",
			initial: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-06-01",
				Passengers:    3,
				Class:         "business",
			},
			wantPassengers: 3,
			wantClass:      "business",
		},
		{
			name: "preserves first class",
			initial: &SearchCriteria{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-06-01",
				Passengers:    1,
				Class:         "first",
			},
			wantPassengers: 1,
			wantClass:      "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.SetDefaults()

			assert.Equal(t, tt.wantPassengers, tt.initial.Passengers)
			assert.Equal(t, tt.wantClass, tt.initial.Class)
		})
	}
}
