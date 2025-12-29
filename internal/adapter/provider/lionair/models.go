// Package lionair provides the Lion Air flight provider adapter.
// It reads from mock JSON data and normalizes it to the unified Flight domain model.
package lionair

// LionAirResponse represents the root response structure from Lion Air API.
type LionAirResponse struct {
	Success bool           `json:"success"`
	Data    LionAirData    `json:"data"`
}

// LionAirData contains the flight data.
type LionAirData struct {
	AvailableFlights []LionAirFlight `json:"available_flights"`
}

// LionAirFlight represents a single flight from the Lion Air API.
type LionAirFlight struct {
	ID         string           `json:"id"`
	Carrier    LionAirCarrier   `json:"carrier"`
	Route      LionAirRoute     `json:"route"`
	Schedule   LionAirSchedule  `json:"schedule"`
	FlightTime int              `json:"flight_time"`
	IsDirect   bool             `json:"is_direct"`
	StopCount  int              `json:"stop_count,omitempty"`
	Layovers   []LionAirLayover `json:"layovers,omitempty"`
	Pricing    LionAirPricing   `json:"pricing"`
	SeatsLeft  int              `json:"seats_left"`
	PlaneType  string           `json:"plane_type"`
	Services   LionAirServices  `json:"services"`
}

// LionAirCarrier contains carrier information.
type LionAirCarrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

// LionAirRoute contains route information.
type LionAirRoute struct {
	From LionAirAirport `json:"from"`
	To   LionAirAirport `json:"to"`
}

// LionAirAirport contains airport information.
type LionAirAirport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

// LionAirSchedule contains schedule information.
type LionAirSchedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

// LionAirLayover contains layover information for connecting flights.
type LionAirLayover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

// LionAirPricing contains pricing information.
type LionAirPricing struct {
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
	FareType string  `json:"fare_type"`
}

// LionAirServices contains additional service information.
type LionAirServices struct {
	WiFiAvailable    bool                  `json:"wifi_available"`
	MealsIncluded    bool                  `json:"meals_included"`
	BaggageAllowance LionAirBaggageAllowance `json:"baggage_allowance"`
}

// LionAirBaggageAllowance contains baggage allowance information.
type LionAirBaggageAllowance struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}
