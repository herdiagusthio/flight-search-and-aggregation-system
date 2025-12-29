// Package usecase contains the business logic for flight search operations.
// It orchestrates provider calls using the Scatter-Gather concurrency pattern.
package usecase

import "github.com/flight-search/flight-search-and-aggregation-system/internal/domain"

// SearchOptions contains optional parameters for a flight search.
type SearchOptions struct {
	// Filters contains optional filtering criteria to apply to results
	Filters *domain.FilterOptions

	// SortBy specifies how to sort the results (default: best value)
	SortBy domain.SortOption
}

// DefaultSearchOptions returns SearchOptions with sensible defaults.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Filters: nil,
		SortBy:  domain.SortByBestValue,
	}
}
