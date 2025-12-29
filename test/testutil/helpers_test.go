package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustParseTime(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			dateStr: "2025-12-15T08:00:00Z",
			wantErr: false,
		},
		{
			name:    "valid RFC3339 with timezone",
			dateStr: "2025-12-15T08:00:00+07:00",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustParseTime(t, tt.dateStr)
			assert.False(t, result.IsZero())
		})
	}
}

func TestMustParseDate(t *testing.T) {
	result := MustParseDate(t, "2025-12-15")
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.December, result.Month())
	assert.Equal(t, 15, result.Day())
}

func TestPtr(t *testing.T) {
	// Test with int
	intVal := Ptr(42)
	require.NotNil(t, intVal)
	assert.Equal(t, 42, *intVal)

	// Test with string
	strVal := Ptr("hello")
	require.NotNil(t, strVal)
	assert.Equal(t, "hello", *strVal)

	// Test with float64
	floatVal := Ptr(3.14)
	require.NotNil(t, floatVal)
	assert.Equal(t, 3.14, *floatVal)
}

func TestFloatPtr(t *testing.T) {
	ptr := FloatPtr(1000000.0)
	require.NotNil(t, ptr)
	assert.Equal(t, 1000000.0, *ptr)
}

func TestIntPtr(t *testing.T) {
	ptr := IntPtr(5)
	require.NotNil(t, ptr)
	assert.Equal(t, 5, *ptr)
}

func TestStringSlice(t *testing.T) {
	slice := StringSlice("GA", "JT", "ID")
	assert.Len(t, slice, 3)
	assert.Contains(t, slice, "GA")
	assert.Contains(t, slice, "JT")
	assert.Contains(t, slice, "ID")
}

func TestLoadMockJSON(t *testing.T) {
	// Test loading a real mock file
	data := LoadMockJSON(t, "garuda_indonesia_search_response.json")
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "Garuda")
}
