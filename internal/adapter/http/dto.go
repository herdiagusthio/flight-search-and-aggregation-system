package http

import (
	"fmt"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// SearchResponseDTO is the data transfer object for search responses.
// It matches the expected API output format with snake_case fields.
type SearchResponseDTO struct {
	SearchCriteria SearchCriteriaDTO `json:"search_criteria"`
	Metadata       MetadataDTO       `json:"metadata"`
	Flights        []FlightDTO       `json:"flights"`
}

// SearchCriteriaDTO represents the search criteria in the response.
type SearchCriteriaDTO struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

// MetadataDTO contains metadata about the search execution.
type MetadataDTO struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

// FlightDTO is the data transfer object for flight responses.
type FlightDTO struct {
	ID             string        `json:"id"`
	Provider       string        `json:"provider"`
	Airline        AirlineDTO    `json:"airline"`
	FlightNumber   string        `json:"flight_number"`
	Departure      FlightPointDTO `json:"departure"`
	Arrival        FlightPointDTO `json:"arrival"`
	Duration       DurationDTO   `json:"duration"`
	Stops          int           `json:"stops"`
	Price          PriceDTO      `json:"price"`
	AvailableSeats *int          `json:"available_seats,omitempty"`
	CabinClass     string        `json:"cabin_class"`
	Aircraft       *string       `json:"aircraft"`
	Amenities      []string      `json:"amenities"`
	Baggage        BaggageDTO    `json:"baggage"`
}

// AirlineDTO represents airline information.
type AirlineDTO struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// FlightPointDTO represents a departure or arrival point.
type FlightPointDTO struct {
	Airport   string `json:"airport"`
	City      string `json:"city,omitempty"`
	DateTime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

// DurationDTO represents flight duration.
type DurationDTO struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

// PriceDTO represents price information.
type PriceDTO struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// BaggageDTO represents baggage information.
type BaggageDTO struct {
	CarryOn string `json:"carry_on,omitempty"`
	Checked string `json:"checked,omitempty"`
}

// ToSearchResponseDTO converts a domain SearchResponse to a SearchResponseDTO.
func ToSearchResponseDTO(resp *domain.SearchResponse) *SearchResponseDTO {
	if resp == nil {
		return nil
	}

	dto := &SearchResponseDTO{
		SearchCriteria: SearchCriteriaDTO{
			Origin:        resp.SearchCriteria.Origin,
			Destination:   resp.SearchCriteria.Destination,
			DepartureDate: resp.SearchCriteria.DepartureDate,
			Passengers:    resp.SearchCriteria.Passengers,
			CabinClass:    resp.SearchCriteria.CabinClass,
		},
		Metadata: MetadataDTO{
			TotalResults:       resp.Metadata.TotalResults,
			ProvidersQueried:   resp.Metadata.ProvidersQueried,
			ProvidersSucceeded: resp.Metadata.ProvidersSucceeded,
			ProvidersFailed:    resp.Metadata.ProvidersFailed,
			SearchTimeMs:       resp.Metadata.SearchTimeMs,
			CacheHit:           resp.Metadata.CacheHit,
		},
		Flights: make([]FlightDTO, len(resp.Flights)),
	}

	for i, flight := range resp.Flights {
		dto.Flights[i] = ToFlightDTO(&flight)
	}

	return dto
}

// ToFlightDTO converts a domain Flight to a FlightDTO.
func ToFlightDTO(flight *domain.Flight) FlightDTO {
	dto := FlightDTO{
		ID:           flight.ID,
		Provider:     flight.Provider,
		FlightNumber: flight.FlightNumber,
		Airline: AirlineDTO{
			Name: flight.Airline.Name,
			Code: flight.Airline.Code,
		},
		Departure: FlightPointDTO{
			Airport:   flight.Departure.AirportCode,
			DateTime:  flight.Departure.DateTime.Format("2006-01-02T15:04:05-07:00"),
			Timestamp: flight.Departure.DateTime.Unix(),
		},
		Arrival: FlightPointDTO{
			Airport:   flight.Arrival.AirportCode,
			DateTime:  flight.Arrival.DateTime.Format("2006-01-02T15:04:05-07:00"),
			Timestamp: flight.Arrival.DateTime.Unix(),
		},
		Duration: DurationDTO{
			TotalMinutes: flight.Duration.TotalMinutes,
			Formatted:    flight.Duration.Formatted,
		},
		Stops:      flight.Stops,
		CabinClass: flight.Class,
		Price: PriceDTO{
			Amount:   flight.Price.Amount,
			Currency: flight.Price.Currency,
		},
		Aircraft:  nil,
		Amenities: []string{},
		Baggage: BaggageDTO{
			CarryOn: formatBaggageKg(flight.Baggage.CabinKg),
			Checked: formatBaggageKg(flight.Baggage.CheckedKg),
		},
	}

	// Add city from airport name if available
	if flight.Departure.AirportName != "" {
		dto.Departure.City = extractCityFromAirportName(flight.Departure.AirportCode)
	}
	if flight.Arrival.AirportName != "" {
		dto.Arrival.City = extractCityFromAirportName(flight.Arrival.AirportCode)
	}

	return dto
}

// formatBaggageKg formats baggage weight in kg to a string.
func formatBaggageKg(kg int) string {
	if kg == 0 {
		return ""
	}
	if kg == 7 {
		return "Cabin baggage only"
	}
	return fmt.Sprintf("%d kg", kg)
}

// extractCityFromAirportName extracts city name from airport code.
// This is a simple mapping for common Indonesian airports.
func extractCityFromAirportName(code string) string {
	cities := map[string]string{
		"CGK": "Jakarta",
		"DPS": "Denpasar",
		"SUB": "Surabaya",
		"JOG": "Yogyakarta",
		"BDO": "Bandung",
		"MDC": "Manado",
		"UPG": "Makassar",
		"BPN": "Balikpapan",
	}
	if city, ok := cities[code]; ok {
		return city
	}
	return ""
}
