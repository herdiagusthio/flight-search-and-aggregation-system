package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealClock_Now(t *testing.T) {
	clock := NewRealClock()

	before := time.Now()
	now := clock.Now()
	after := time.Now()

	// The clock time should be between before and after
	assert.False(t, now.Before(before), "clock time should not be before start")
	assert.False(t, now.After(after), "clock time should not be after end")
}

func TestRealClock_Interface(t *testing.T) {
	// Ensure RealClock implements Clock interface
	var _ Clock = (*RealClock)(nil)
	var _ Clock = NewRealClock()
}

func TestMockClock_Now(t *testing.T) {
	fixedTime := time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC)
	clock := NewMockClock(fixedTime)

	// Should always return the fixed time
	assert.Equal(t, fixedTime, clock.Now())
	assert.Equal(t, fixedTime, clock.Now())
	assert.Equal(t, fixedTime, clock.Now())
}

func TestMockClock_Set(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2025, 12, 15, 14, 30, 0, 0, time.UTC)

	clock := NewMockClock(initialTime)
	assert.Equal(t, initialTime, clock.Now())

	clock.Set(newTime)
	assert.Equal(t, newTime, clock.Now())
}

func TestMockClock_Advance(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	clock.Advance(30 * time.Minute)

	expected := time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_AdvanceMinutes(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	clock.AdvanceMinutes(45)

	expected := time.Date(2025, 12, 15, 10, 45, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_AdvanceHours(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	clock.AdvanceHours(3)

	expected := time.Date(2025, 12, 15, 13, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_AdvanceDays(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	clock.AdvanceDays(5)

	expected := time.Date(2025, 12, 20, 10, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_Interface(t *testing.T) {
	// Ensure MockClock implements Clock interface
	var _ Clock = (*MockClock)(nil)
	var _ Clock = NewMockClock(time.Now())
}

func TestNewMockClockFromString(t *testing.T) {
	clock := NewMockClockFromString("2025-12-15T10:30:00Z")

	expected := time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestNewMockClockFromString_Panic(t *testing.T) {
	assert.Panics(t, func() {
		NewMockClockFromString("invalid-time")
	})
}

func TestClock_UsageInCode(t *testing.T) {
	// This test demonstrates how Clock can be used for dependency injection
	type FlightService struct {
		clock Clock
	}

	getDepartureTime := func(s *FlightService) time.Time {
		return s.clock.Now().Add(2 * time.Hour)
	}

	// In tests, use MockClock
	fixedTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	service := &FlightService{clock: NewMockClock(fixedTime)}

	departure := getDepartureTime(service)
	expected := time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, departure)

	// In production, use RealClock
	realService := &FlightService{clock: NewRealClock()}
	realDeparture := getDepartureTime(realService)
	assert.True(t, realDeparture.After(time.Now()))
}

func TestMockClock_MultipleAdvances(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	clock.AdvanceMinutes(30) // 10:30
	clock.AdvanceHours(2)    // 12:30
	clock.AdvanceDays(1)     // Next day 12:30

	expected := time.Date(2025, 12, 16, 12, 30, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_NegativeAdvance(t *testing.T) {
	initialTime := time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC)
	clock := NewMockClock(initialTime)

	// Can go backwards too
	clock.Advance(-2 * time.Hour)

	expected := time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, clock.Now())
}

func TestMockClock_WithTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	require.NoError(t, err)

	// Create time in Jakarta timezone
	jakartaTime := time.Date(2025, 12, 15, 17, 0, 0, 0, loc)
	clock := NewMockClock(jakartaTime)

	now := clock.Now()
	assert.Equal(t, loc, now.Location())
	assert.Equal(t, 17, now.Hour())
}
