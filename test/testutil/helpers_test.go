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
	tests := []struct {
		name      string
		dateStr   string
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "valid date",
			dateStr:   "2025-12-15",
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   15,
		},
		{
			name:      "january date",
			dateStr:   "2025-01-01",
			wantYear:  2025,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			name:      "leap year date",
			dateStr:   "2024-02-29",
			wantYear:  2024,
			wantMonth: time.February,
			wantDay:   29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustParseDate(t, tt.dateStr)
			assert.Equal(t, tt.wantYear, result.Year())
			assert.Equal(t, tt.wantMonth, result.Month())
			assert.Equal(t, tt.wantDay, result.Day())
		})
	}
}

func TestPtr(t *testing.T) {
	t.Run("int value", func(t *testing.T) {
		intVal := Ptr(42)
		require.NotNil(t, intVal)
		assert.Equal(t, 42, *intVal)
	})

	t.Run("string value", func(t *testing.T) {
		strVal := Ptr("hello")
		require.NotNil(t, strVal)
		assert.Equal(t, "hello", *strVal)
	})

	t.Run("float64 value", func(t *testing.T) {
		floatVal := Ptr(3.14)
		require.NotNil(t, floatVal)
		assert.Equal(t, 3.14, *floatVal)
	})

	t.Run("bool value", func(t *testing.T) {
		boolVal := Ptr(true)
		require.NotNil(t, boolVal)
		assert.Equal(t, true, *boolVal)
	})
}

func TestFloatPtr(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{name: "large value", value: 1000000.0},
		{name: "small value", value: 0.01},
		{name: "zero", value: 0.0},
		{name: "negative", value: -500.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := FloatPtr(tt.value)
			require.NotNil(t, ptr)
			assert.Equal(t, tt.value, *ptr)
		})
	}
}

func TestIntPtr(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{name: "positive", value: 5},
		{name: "zero", value: 0},
		{name: "negative", value: -10},
		{name: "large", value: 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := IntPtr(tt.value)
			require.NotNil(t, ptr)
			assert.Equal(t, tt.value, *ptr)
		})
	}
}

func TestStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		wantLen  int
		contains []string
	}{
		{
			name:     "three values",
			values:   []string{"GA", "JT", "ID"},
			wantLen:  3,
			contains: []string{"GA", "JT", "ID"},
		},
		{
			name:     "single value",
			values:   []string{"GA"},
			wantLen:  1,
			contains: []string{"GA"},
		},
		{
			name:     "empty",
			values:   []string{},
			wantLen:  0,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := StringSlice(tt.values...)
			assert.Len(t, slice, tt.wantLen)
			for _, val := range tt.contains {
				assert.Contains(t, slice, val)
			}
		})
	}
}

func TestLoadMockJSON(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		shouldContain string
	}{
		{
			name:          "garuda response",
			filename:      "garuda_indonesia_search_response.json",
			shouldContain: "Garuda",
		},
		{
			name:          "lion air response",
			filename:      "lion_air_search_response.json",
			shouldContain: "Lion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := LoadMockJSON(t, tt.filename)
			assert.NotEmpty(t, data)
			assert.Contains(t, string(data), tt.shouldContain)
		})
	}
}
