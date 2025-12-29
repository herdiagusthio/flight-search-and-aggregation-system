// Package http provides swagger type definitions for API documentation.
// These types mirror domain types but are defined here to help swag generate proper documentation.
package http

import "time"

// SwaggerSearchResponse represents the search API response for swagger documentation.
// @Description Flight search results with metadata
type SwaggerSearchResponse struct {
	// Flights contains the list of flight results after filtering and sorting
	Flights []SwaggerFlight `json:"flights"`

	// Metadata contains information about the search execution
	Metadata SwaggerSearchMetadata `json:"metadata"`
}

// SwaggerSearchMetadata contains metadata about the search execution.
// @Description Metadata about the search execution
type SwaggerSearchMetadata struct {
	// TotalResults is the total number of flights returned
	TotalResults int `json:"totalResults" example:"15"`

	// SearchDurationMs is the total search duration in milliseconds
	SearchDurationMs int64 `json:"searchDurationMs" example:"1250"`

	// ProvidersQueried is the list of provider names that were queried
	ProvidersQueried []string `json:"providersQueried" example:"garuda,lionair,batikair,airasia"`

	// ProvidersFailed is the list of provider names that failed or timed out
	ProvidersFailed []string `json:"providersFailed,omitempty" example:""`
}

// SwaggerFlight represents a single flight offering.
// @Description Flight information from an airline provider
type SwaggerFlight struct {
	// ID is a unique identifier for this flight result
	ID string `json:"id" example:"garuda-ga123-cgk-dps-20251215"`

	// FlightNumber is the airline's flight number
	FlightNumber string `json:"flightNumber" example:"GA-123"`

	// Airline contains information about the operating airline
	Airline SwaggerAirlineInfo `json:"airline"`

	// Departure contains departure airport and time information
	Departure SwaggerFlightPoint `json:"departure"`

	// Arrival contains arrival airport and time information
	Arrival SwaggerFlightPoint `json:"arrival"`

	// Duration contains the total flight duration
	Duration SwaggerDurationInfo `json:"duration"`

	// Price contains pricing information
	Price SwaggerPriceInfo `json:"price"`

	// Baggage contains baggage allowance information
	Baggage SwaggerBaggageInfo `json:"baggage"`

	// Class is the travel class
	Class string `json:"class" example:"economy"`

	// Stops is the number of stops (0 = direct flight)
	Stops int `json:"stops" example:"0"`

	// Provider identifies which flight provider this result came from
	Provider string `json:"provider" example:"garuda"`

	// RankingScore is the calculated score for sorting by "best value"
	RankingScore float64 `json:"rankingScore,omitempty" example:"85.5"`
}

// SwaggerAirlineInfo contains information about an airline.
// @Description Airline information
type SwaggerAirlineInfo struct {
	// Code is the IATA airline code
	Code string `json:"code" example:"GA"`

	// Name is the full airline name
	Name string `json:"name" example:"Garuda Indonesia"`

	// Logo is an optional URL to the airline's logo image
	Logo string `json:"logo,omitempty" example:"https://example.com/ga-logo.png"`
}

// SwaggerFlightPoint represents a point in a flight journey.
// @Description Departure or arrival point information
type SwaggerFlightPoint struct {
	// AirportCode is the IATA airport code
	AirportCode string `json:"airportCode" example:"CGK"`

	// AirportName is the full airport name
	AirportName string `json:"airportName,omitempty" example:"Soekarno-Hatta International Airport"`

	// Terminal is the terminal identifier
	Terminal string `json:"terminal,omitempty" example:"3"`

	// DateTime is the scheduled departure or arrival time
	DateTime time.Time `json:"dateTime" example:"2025-12-15T08:00:00Z"`

	// Timezone is the IANA timezone identifier
	Timezone string `json:"timezone,omitempty" example:"Asia/Jakarta"`
}

// SwaggerDurationInfo contains flight duration information.
// @Description Flight duration information
type SwaggerDurationInfo struct {
	// TotalMinutes is the total flight duration in minutes
	TotalMinutes int `json:"totalMinutes" example:"95"`

	// Display is a human-readable duration string
	Display string `json:"display" example:"1h 35m"`
}

// SwaggerPriceInfo contains pricing information.
// @Description Price information
type SwaggerPriceInfo struct {
	// Amount is the price value
	Amount float64 `json:"amount" example:"1250000"`

	// Currency is the ISO 4217 currency code
	Currency string `json:"currency" example:"IDR"`

	// Display is a formatted price string
	Display string `json:"display,omitempty" example:"IDR 1,250,000"`
}

// SwaggerBaggageInfo contains baggage allowance information.
// @Description Baggage allowance information
type SwaggerBaggageInfo struct {
	// Cabin is the cabin baggage allowance in kg
	Cabin int `json:"cabin" example:"7"`

	// Checked is the checked baggage allowance in kg
	Checked int `json:"checked" example:"20"`
}

// SwaggerErrorResponse represents an error response.
// @Description Error response from the API
type SwaggerErrorResponse struct {
	// Success is always false for error responses
	Success bool `json:"success" example:"false"`

	// Error contains error details
	Error SwaggerErrorDetail `json:"error"`
}

// SwaggerErrorDetail contains structured error information.
// @Description Error details
type SwaggerErrorDetail struct {
	// Code is a machine-readable error code
	Code string `json:"code" example:"validation_error"`

	// Message is a human-readable error message
	Message string `json:"message" example:"Request validation failed"`

	// Details contains field-specific error details
	Details map[string]string `json:"details,omitempty"`
}
