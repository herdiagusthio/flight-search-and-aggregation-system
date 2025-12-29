// Package timeutil provides time-related utilities for testability and convenience.
package timeutil

import (
	"fmt"
	"sync"
	"time"
)

// locationCache stores cached timezone locations for performance.
var locationCache sync.Map

// Common timezone names for convenience.
const (
	// UTC is the Coordinated Universal Time.
	UTC = "UTC"

	// WIB is Western Indonesian Time (Jakarta, Bandung).
	WIB = "Asia/Jakarta"

	// WITA is Central Indonesian Time (Bali, Makassar).
	WITA = "Asia/Makassar"

	// WIT is Eastern Indonesian Time (Jayapura).
	WIT = "Asia/Jayapura"

	// SGT is Singapore Time.
	SGT = "Asia/Singapore"

	// JST is Japan Standard Time.
	JST = "Asia/Tokyo"
)

// GetLocation returns a cached timezone location.
// It caches the result for subsequent calls with the same name.
func GetLocation(name string) (*time.Location, error) {
	// Check cache first
	if loc, ok := locationCache.Load(name); ok {
		return loc.(*time.Location), nil
	}

	// Load location
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %q: %w", name, err)
	}

	// Store in cache
	locationCache.Store(name, loc)
	return loc, nil
}

// MustGetLocation returns a cached timezone location or panics on error.
// Use this for known-good timezone names (e.g., constants).
func MustGetLocation(name string) *time.Location {
	loc, err := GetLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// InTimezone converts a time to the specified timezone.
func InTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		return t, err
	}
	return t.In(loc), nil
}

// NowIn returns the current time in the specified timezone.
func NowIn(timezone string) (time.Time, error) {
	return InTimezone(time.Now(), timezone)
}

// NowInJakarta returns the current time in Jakarta timezone (WIB).
func NowInJakarta() time.Time {
	loc := MustGetLocation(WIB)
	return time.Now().In(loc)
}

// NowInUTC returns the current time in UTC.
func NowInUTC() time.Time {
	return time.Now().UTC()
}

// ParseInTimezone parses a time string in the specified timezone.
func ParseInTimezone(layout, value, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, loc)
}

// FormatDate formats a time as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTime formats a time as HH:MM.
func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

// FormatDateTime formats a time as YYYY-MM-DD HH:MM:SS.
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// StartOfDay returns the start of the day (00:00:00) for the given time.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day (23:59:59.999999999) for the given time.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// ClearLocationCache clears the cached timezone locations.
// This is primarily useful for testing.
func ClearLocationCache() {
	locationCache.Range(func(key, _ interface{}) bool {
		locationCache.Delete(key)
		return true
	})
}
