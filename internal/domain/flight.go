// Package domain contains the core business entities and rules for the flight search system.
// These entities are provider-agnostic and form the foundation upon which all other components are built.
package domain

import "time"

// Flight represents a single flight offering from a provider.
// It contains all the information needed to display and compare flights.
type Flight struct {
	// ID is a unique identifier for this flight result (generated internally)
	ID string `json:"id"`

	// FlightNumber is the airline's flight number (e.g., "GA-123")
	FlightNumber string `json:"flightNumber"`

	// Airline contains information about the operating airline
	Airline AirlineInfo `json:"airline"`

	// Departure contains departure airport and time information
	Departure FlightPoint `json:"departure"`

	// Arrival contains arrival airport and time information
	Arrival FlightPoint `json:"arrival"`

	// Duration contains the total flight duration
	Duration DurationInfo `json:"duration"`

	// Price contains pricing information
	Price PriceInfo `json:"price"`

	// Baggage contains baggage allowance information
	Baggage BaggageInfo `json:"baggage"`

	// Class is the travel class (economy, business, first)
	Class string `json:"class"`

	// Stops is the number of stops (0 = direct flight)
	Stops int `json:"stops"`

	// Provider identifies which flight provider this result came from
	Provider string `json:"provider"`

	// RankingScore is the calculated score for sorting by "best value"
	// Higher scores indicate better value (considers price, duration, stops)
	RankingScore float64 `json:"rankingScore,omitempty"`
}

// AirlineInfo contains information about an airline.
type AirlineInfo struct {
	// Code is the IATA airline code (e.g., "GA" for Garuda Indonesia)
	Code string `json:"code"`

	// Name is the full airline name (e.g., "Garuda Indonesia")
	Name string `json:"name"`

	// Logo is an optional URL to the airline's logo image
	Logo string `json:"logo,omitempty"`
}

// FlightPoint represents a point in a flight journey (departure or arrival).
type FlightPoint struct {
	// AirportCode is the IATA airport code (e.g., "CGK")
	AirportCode string `json:"airportCode"`

	// AirportName is the full airport name (e.g., "Soekarno-Hatta International Airport")
	AirportName string `json:"airportName,omitempty"`

	// Terminal is the terminal identifier (e.g., "3")
	Terminal string `json:"terminal,omitempty"`

	// DateTime is the scheduled departure or arrival time
	DateTime time.Time `json:"dateTime"`

	// Timezone is the IANA timezone identifier (e.g., "Asia/Jakarta")
	Timezone string `json:"timezone,omitempty"`
}

// DurationInfo contains flight duration information.
type DurationInfo struct {
	// TotalMinutes is the total flight duration in minutes
	TotalMinutes int `json:"totalMinutes"`

	// Formatted is a human-readable duration string (e.g., "2h 30m")
	Formatted string `json:"formatted"`
}

// PriceInfo contains pricing information for a flight.
type PriceInfo struct {
	// Amount is the numeric price value
	Amount float64 `json:"amount"`

	// Currency is the ISO 4217 currency code (e.g., "IDR", "USD")
	Currency string `json:"currency"`

	// Formatted is an optional human-readable price string (e.g., "IDR 1,500,000")
	Formatted string `json:"formatted,omitempty"`
}

// BaggageInfo contains baggage allowance information.
type BaggageInfo struct {
	// CabinKg is the cabin baggage allowance in kilograms
	CabinKg int `json:"cabinKg"`

	// CheckedKg is the checked baggage allowance in kilograms
	CheckedKg int `json:"checkedKg"`
}

// NewDurationInfo creates a DurationInfo from total minutes and formats it.
func NewDurationInfo(totalMinutes int) DurationInfo {
	hours := totalMinutes / 60
	mins := totalMinutes % 60

	var formatted string
	if hours > 0 && mins > 0 {
		formatted = formatDuration(hours, mins)
	} else if hours > 0 {
		formatted = formatHoursOnly(hours)
	} else {
		formatted = formatMinutesOnly(mins)
	}

	return DurationInfo{
		TotalMinutes: totalMinutes,
		Formatted:    formatted,
	}
}

// formatDuration formats hours and minutes as "Xh Ym".
func formatDuration(hours, mins int) string {
	return intToString(hours) + "h " + intToString(mins) + "m"
}

// formatHoursOnly formats hours as "Xh".
func formatHoursOnly(hours int) string {
	return intToString(hours) + "h"
}

// formatMinutesOnly formats minutes as "Xm".
func formatMinutesOnly(mins int) string {
	return intToString(mins) + "m"
}

// intToString converts an integer to a string without importing strconv.
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}