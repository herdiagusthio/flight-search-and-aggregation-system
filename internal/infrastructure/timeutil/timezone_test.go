package timeutil

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLocation_UTC(t *testing.T) {
	ClearLocationCache()

	loc, err := GetLocation("UTC")
	require.NoError(t, err)
	assert.NotNil(t, loc)
	assert.Equal(t, "UTC", loc.String())
}

func TestGetLocation_Jakarta(t *testing.T) {
	ClearLocationCache()

	loc, err := GetLocation("Asia/Jakarta")
	require.NoError(t, err)
	assert.NotNil(t, loc)
	assert.Equal(t, "Asia/Jakarta", loc.String())
}

func TestGetLocation_Invalid(t *testing.T) {
	ClearLocationCache()

	loc, err := GetLocation("Invalid/Timezone")
	assert.Error(t, err)
	assert.Nil(t, loc)
	assert.Contains(t, err.Error(), "failed to load timezone")
}

func TestGetLocation_Caching(t *testing.T) {
	ClearLocationCache()

	// First call should load the location
	loc1, err := GetLocation("Asia/Tokyo")
	require.NoError(t, err)

	// Second call should return cached location
	loc2, err := GetLocation("Asia/Tokyo")
	require.NoError(t, err)

	// Should be the exact same pointer
	assert.Same(t, loc1, loc2)
}

func TestGetLocation_ConcurrentAccess(t *testing.T) {
	ClearLocationCache()

	var wg sync.WaitGroup
	locations := []string{"UTC", "Asia/Jakarta", "Asia/Tokyo", "America/New_York", "Europe/London"}

	// Spawn goroutines to access locations concurrently
	for i := 0; i < 100; i++ {
		for _, tz := range locations {
			wg.Add(1)
			go func(timezone string) {
				defer wg.Done()
				loc, err := GetLocation(timezone)
				assert.NoError(t, err)
				assert.NotNil(t, loc)
			}(tz)
		}
	}

	wg.Wait()
}

func TestMustGetLocation_Valid(t *testing.T) {
	ClearLocationCache()

	// Should not panic
	loc := MustGetLocation("UTC")
	assert.NotNil(t, loc)
}

func TestMustGetLocation_Invalid(t *testing.T) {
	ClearLocationCache()

	// Should panic
	assert.Panics(t, func() {
		MustGetLocation("Invalid/Timezone")
	})
}

func TestInTimezone(t *testing.T) {
	utcTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	jakartaTime, err := InTimezone(utcTime, "Asia/Jakarta")
	require.NoError(t, err)

	// Jakarta is UTC+7
	assert.Equal(t, 17, jakartaTime.Hour())
	assert.Equal(t, "Asia/Jakarta", jakartaTime.Location().String())
}

func TestInTimezone_InvalidTimezone(t *testing.T) {
	utcTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)

	_, err := InTimezone(utcTime, "Invalid/Timezone")
	assert.Error(t, err)
}

func TestNowIn(t *testing.T) {
	ClearLocationCache()

	jakartaTime, err := NowIn("Asia/Jakarta")
	require.NoError(t, err)

	assert.Equal(t, "Asia/Jakarta", jakartaTime.Location().String())
}

func TestNowInJakarta(t *testing.T) {
	jakartaTime := NowInJakarta()

	assert.Equal(t, "Asia/Jakarta", jakartaTime.Location().String())
}

func TestNowInUTC(t *testing.T) {
	utcTime := NowInUTC()

	assert.Equal(t, "UTC", utcTime.Location().String())
}

func TestParseInTimezone(t *testing.T) {
	ClearLocationCache()

	parsed, err := ParseInTimezone("2006-01-02 15:04", "2025-12-15 10:30", "Asia/Jakarta")
	require.NoError(t, err)

	assert.Equal(t, 2025, parsed.Year())
	assert.Equal(t, time.December, parsed.Month())
	assert.Equal(t, 15, parsed.Day())
	assert.Equal(t, 10, parsed.Hour())
	assert.Equal(t, 30, parsed.Minute())
	assert.Equal(t, "Asia/Jakarta", parsed.Location().String())
}

func TestParseInTimezone_InvalidTimezone(t *testing.T) {
	_, err := ParseInTimezone("2006-01-02", "2025-12-15", "Invalid/Timezone")
	assert.Error(t, err)
}

func TestFormatDate(t *testing.T) {
	tm := time.Date(2025, 12, 15, 10, 30, 45, 0, time.UTC)

	assert.Equal(t, "2025-12-15", FormatDate(tm))
}

func TestFormatTime(t *testing.T) {
	tm := time.Date(2025, 12, 15, 10, 30, 45, 0, time.UTC)

	assert.Equal(t, "10:30", FormatTime(tm))
}

func TestFormatDateTime(t *testing.T) {
	tm := time.Date(2025, 12, 15, 10, 30, 45, 0, time.UTC)

	assert.Equal(t, "2025-12-15 10:30:45", FormatDateTime(tm))
}

func TestStartOfDay(t *testing.T) {
	tm := time.Date(2025, 12, 15, 14, 35, 22, 123456789, time.UTC)

	result := StartOfDay(tm)

	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.December, result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 0, result.Hour())
	assert.Equal(t, 0, result.Minute())
	assert.Equal(t, 0, result.Second())
	assert.Equal(t, 0, result.Nanosecond())
}

func TestEndOfDay(t *testing.T) {
	tm := time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC)

	result := EndOfDay(tm)

	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.December, result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 23, result.Hour())
	assert.Equal(t, 59, result.Minute())
	assert.Equal(t, 59, result.Second())
	assert.Equal(t, 999999999, result.Nanosecond())
}

func TestClearLocationCache(t *testing.T) {
	// Load some locations
	_, _ = GetLocation("UTC")
	_, _ = GetLocation("Asia/Jakarta")

	// Clear cache
	ClearLocationCache()

	// Verify cache is cleared by checking internal state
	// (indirect verification through successful re-loading)
	loc1, err := GetLocation("UTC")
	require.NoError(t, err)

	loc2, err := GetLocation("UTC")
	require.NoError(t, err)

	// After re-loading, should be cached again
	assert.Same(t, loc1, loc2)
}

func TestTimezoneConstants(t *testing.T) {
	// Verify all timezone constants are valid
	timezones := []string{UTC, WIB, WITA, WIT, SGT, JST}

	for _, tz := range timezones {
		loc, err := GetLocation(tz)
		assert.NoError(t, err, "timezone %s should be valid", tz)
		assert.NotNil(t, loc)
	}
}

func TestStartOfDay_PreservesLocation(t *testing.T) {
	loc := MustGetLocation("Asia/Jakarta")
	tm := time.Date(2025, 12, 15, 14, 35, 22, 0, loc)

	result := StartOfDay(tm)

	assert.Equal(t, loc, result.Location())
}

func TestEndOfDay_PreservesLocation(t *testing.T) {
	loc := MustGetLocation("Asia/Tokyo")
	tm := time.Date(2025, 12, 15, 10, 30, 0, 0, loc)

	result := EndOfDay(tm)

	assert.Equal(t, loc, result.Location())
}
