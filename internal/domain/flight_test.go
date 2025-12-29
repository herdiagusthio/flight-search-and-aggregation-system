package domain

import (
	"testing"

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
