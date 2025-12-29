// Package batikair provides the Batik Air flight provider adapter.
// It reads from mock JSON data and normalizes it to the unified Flight domain model.
package batikair

// BatikAirResponse represents the root response structure from Batik Air API.
type BatikAirResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Results []BatikAirFlight `json:"results"`
}

// BatikAirFlight represents a single flight from the Batik Air API.
type BatikAirFlight struct {
	FlightNumber      string              `json:"flightNumber"`
	AirlineName       string              `json:"airlineName"`
	AirlineIATA       string              `json:"airlineIATA"`
	Origin            string              `json:"origin"`
	Destination       string              `json:"destination"`
	DepartureDateTime string              `json:"departureDateTime"`
	ArrivalDateTime   string              `json:"arrivalDateTime"`
	TravelTime        string              `json:"travelTime"`
	NumberOfStops     int                 `json:"numberOfStops"`
	Connections       []BatikAirConnection `json:"connections,omitempty"`
	Fare              BatikAirFare        `json:"fare"`
	SeatsAvailable    int                 `json:"seatsAvailable"`
	AircraftModel     string              `json:"aircraftModel"`
	BaggageInfo       string              `json:"baggageInfo"`
	OnboardServices   []string            `json:"onboardServices,omitempty"`
}

// BatikAirConnection represents a connection/layover in the journey.
type BatikAirConnection struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

// BatikAirFare contains pricing information.
type BatikAirFare struct {
	BasePrice    float64 `json:"basePrice"`
	Taxes        float64 `json:"taxes"`
	TotalPrice   float64 `json:"totalPrice"`
	CurrencyCode string  `json:"currencyCode"`
	Class        string  `json:"class"`
}
