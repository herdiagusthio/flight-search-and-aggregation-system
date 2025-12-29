package domain

// SearchResponse represents the aggregated response from a flight search.
type SearchResponse struct {
	// Flights contains the list of flight results after filtering and sorting
	Flights []Flight `json:"flights"`

	// Metadata contains information about the search execution
	Metadata SearchMetadata `json:"metadata"`
}

// SearchMetadata contains metadata about the search execution.
type SearchMetadata struct {
	// TotalResults is the total number of flights returned
	TotalResults int `json:"totalResults"`

	// SearchDurationMs is the total search duration in milliseconds
	SearchDurationMs int64 `json:"searchDurationMs"`

	// ProvidersQueried is the list of provider names that were queried
	ProvidersQueried []string `json:"providersQueried"`

	// ProvidersFailed is the list of provider names that failed or timed out
	ProvidersFailed []string `json:"providersFailed,omitempty"`
}

// NewSearchResponse creates a new SearchResponse with the given flights and metadata.
func NewSearchResponse(flights []Flight, metadata SearchMetadata) SearchResponse {
	if flights == nil {
		flights = []Flight{}
	}
	metadata.TotalResults = len(flights)

	return SearchResponse{
		Flights:  flights,
		Metadata: metadata,
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
