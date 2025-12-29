// Package airasia provides the AirAsia flight provider adapter.
package airasia

// AirAsiaResponse represents the root response from AirAsia API.
type AirAsiaResponse struct {
	Status  string           `json:"status"`
	Flights []AirAsiaFlight `json:"flights"`
}

// AirAsiaFlight represents a single flight in the AirAsia response.
// Note: AirAsia uses a flat structure with some unique field naming.
type AirAsiaFlight struct {
	// FlightCode is the flight identifier (e.g., "QZ520")
	FlightCode string `json:"flight_code"`

	// Airline is the airline name (e.g., "AirAsia")
	Airline string `json:"airline"`

	// FromAirport is the origin airport IATA code
	FromAirport string `json:"from_airport"`

	// ToAirport is the destination airport IATA code
	ToAirport string `json:"to_airport"`

	// DepartTime is the departure datetime in ISO 8601 format
	DepartTime string `json:"depart_time"`

	// ArriveTime is the arrival datetime in ISO 8601 format
	ArriveTime string `json:"arrive_time"`

	// DurationHours is the flight duration as a float (e.g., 1.75 = 1h 45m)
	DurationHours float64 `json:"duration_hours"`

	// DirectFlight indicates whether this is a non-stop flight
	DirectFlight bool `json:"direct_flight"`

	// Stops contains stop information when DirectFlight is false
	Stops []AirAsiaStop `json:"stops,omitempty"`

	// PriceIDR is the ticket price in Indonesian Rupiah
	PriceIDR float64 `json:"price_idr"`

	// Seats is the number of available seats
	Seats int `json:"seats"`

	// CabinClass is the travel class (e.g., "economy")
	CabinClass string `json:"cabin_class"`

	// BaggageNote contains baggage allowance information as a descriptive string
	BaggageNote string `json:"baggage_note"`
}

// AirAsiaStop represents a stop on a connecting flight.
type AirAsiaStop struct {
	// Airport is the IATA code of the stop airport
	Airport string `json:"airport"`

	// WaitTimeMinutes is the layover time in minutes
	WaitTimeMinutes int `json:"wait_time_minutes"`
}
