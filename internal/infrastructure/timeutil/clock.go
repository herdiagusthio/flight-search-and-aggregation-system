// Package timeutil provides time-related utilities for testability and convenience.
package timeutil

import (
	"time"
)

// Clock provides an abstraction over time.Now() for testability.
// Use RealClock in production and MockClock in tests.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// RealClock uses the actual system time.
type RealClock struct{}

// NewRealClock creates a new RealClock instance.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current system time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// MockClock returns a controllable time for testing.
type MockClock struct {
	fixedTime time.Time
}

// NewMockClock creates a mock clock with the given fixed time.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{fixedTime: t}
}

// NewMockClockFromString creates a mock clock from an RFC3339 time string.
// Panics if the time string is invalid (for use in tests only).
func NewMockClockFromString(timeStr string) *MockClock {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic("invalid time string: " + err.Error())
	}
	return &MockClock{fixedTime: t}
}

// Now returns the fixed time.
func (m *MockClock) Now() time.Time {
	return m.fixedTime
}

// Set sets the mock clock to a specific time.
func (m *MockClock) Set(t time.Time) {
	m.fixedTime = t
}

// Advance moves the mock clock forward by the given duration.
func (m *MockClock) Advance(d time.Duration) {
	m.fixedTime = m.fixedTime.Add(d)
}

// AdvanceMinutes moves the mock clock forward by the given number of minutes.
func (m *MockClock) AdvanceMinutes(minutes int) {
	m.Advance(time.Duration(minutes) * time.Minute)
}

// AdvanceHours moves the mock clock forward by the given number of hours.
func (m *MockClock) AdvanceHours(hours int) {
	m.Advance(time.Duration(hours) * time.Hour)
}

// AdvanceDays moves the mock clock forward by the given number of days.
func (m *MockClock) AdvanceDays(days int) {
	m.Advance(time.Duration(days) * 24 * time.Hour)
}

// Ensure interfaces are implemented.
var (
	_ Clock = (*RealClock)(nil)
	_ Clock = (*MockClock)(nil)
)
