// Package garuda provides the Garuda Indonesia flight provider adapter.
// It reads from mock JSON data and normalizes it to the unified Flight domain model.
package garuda

// GarudaResponse represents the root response structure from Garuda Indonesia API.
type GarudaResponse struct {
	Status  string         `json:"status"`
	Flights []GarudaFlight `json:"flights"`
}

// GarudaFlight represents a single flight from the Garuda Indonesia API.
type GarudaFlight struct {
	FlightID        string          `json:"flight_id"`
	Airline         string          `json:"airline"`
	AirlineCode     string          `json:"airline_code"`
	Departure       GarudaEndpoint  `json:"departure"`
	Arrival         GarudaEndpoint  `json:"arrival"`
	DurationMinutes int             `json:"duration_minutes"`
	Stops           int             `json:"stops"`
	Aircraft        string          `json:"aircraft"`
	Price           GarudaPrice     `json:"price"`
	AvailableSeats  int             `json:"available_seats"`
	FareClass       string          `json:"fare_class"`
	Baggage         GarudaBaggage   `json:"baggage"`
	Amenities       []string        `json:"amenities,omitempty"`
	Segments        []GarudaSegment `json:"segments,omitempty"`
}

// GarudaEndpoint represents a departure or arrival point.
type GarudaEndpoint struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal,omitempty"`
}

// GarudaPrice contains pricing information.
type GarudaPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// GarudaBaggage contains baggage allowance information.
// Values are in number of pieces, not weight.
type GarudaBaggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

// GarudaSegment represents a flight segment for multi-leg flights.
type GarudaSegment struct {
	FlightNumber    string               `json:"flight_number"`
	Departure       GarudaSegmentPoint   `json:"departure"`
	Arrival         GarudaSegmentPoint   `json:"arrival"`
	DurationMinutes int                  `json:"duration_minutes"`
	LayoverMinutes  int                  `json:"layover_minutes,omitempty"`
}

// GarudaSegmentPoint represents a point within a segment.
type GarudaSegmentPoint struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}
