package domain

// SearchResponse represents the aggregated response from a flight search.
type SearchResponse struct {
	// SearchCriteria contains the original search parameters
	SearchCriteria SearchCriteriaResponse `json:"search_criteria"`

	// Metadata contains information about the search execution
	Metadata SearchMetadata `json:"metadata"`

	// Flights contains the list of flight results after filtering and sorting
	Flights []Flight `json:"flights"`
}

// SearchCriteriaResponse represents the search criteria in the response.
type SearchCriteriaResponse struct {
	// Origin is the IATA code of the departure airport
	Origin string `json:"origin"`

	// Destination is the IATA code of the arrival airport
	Destination string `json:"destination"`

	// DepartureDate is the desired departure date in YYYY-MM-DD format
	DepartureDate string `json:"departure_date"`

	// Passengers is the number of passengers
	Passengers int `json:"passengers"`

	// CabinClass is the travel class
	CabinClass string `json:"cabin_class"`
}

// SearchMetadata contains metadata about the search execution.
type SearchMetadata struct {
	// TotalResults is the total number of flights returned
	TotalResults int `json:"total_results"`

	// ProvidersQueried is the number of providers that were queried
	ProvidersQueried int `json:"providers_queried"`

	// ProvidersSucceeded is the number of providers that returned results successfully
	ProvidersSucceeded int `json:"providers_succeeded"`

	// ProvidersFailed is the number of providers that failed or timed out
	ProvidersFailed int `json:"providers_failed"`

	// SearchTimeMs is the total search duration in milliseconds
	SearchTimeMs int64 `json:"search_time_ms"`

	// CacheHit indicates whether the results came from cache
	CacheHit bool `json:"cache_hit"`
}

// NewSearchResponse creates a new SearchResponse with the given criteria, flights, and metadata.
func NewSearchResponse(criteria *SearchCriteria, flights []Flight, metadata SearchMetadata) SearchResponse {
	if flights == nil {
		flights = []Flight{}
	}
	metadata.TotalResults = len(flights)

	// Convert SearchCriteria to SearchCriteriaResponse
	criteriaResp := SearchCriteriaResponse{
		Origin:        criteria.Origin,
		Destination:   criteria.Destination,
		DepartureDate: criteria.DepartureDate,
		Passengers:    criteria.Passengers,
		CabinClass:    criteria.Class,
	}

	return SearchResponse{
		SearchCriteria: criteriaResp,
		Metadata:       metadata,
		Flights:        flights,
	}
}

// ProviderResult represents the result from a single provider query.
// This is used internally for aggregating results.
type ProviderResult struct {
	// Provider is the name of the provider
	Provider string

	// Flights contains the flights returned by this provider
	Flights []Flight

	// Error is set if the provider query failed
	Error error

	// DurationMs is how long the provider query took
	DurationMs int64
}

// IsSuccess returns true if the provider query succeeded.
func (pr *ProviderResult) IsSuccess() bool {
	return pr.Error == nil
}
