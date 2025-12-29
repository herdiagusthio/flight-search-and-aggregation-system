package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// Default timeout values.
const (
	DefaultGlobalTimeout   = 5 * time.Second
	DefaultProviderTimeout = 2 * time.Second
)

// FlightSearchUseCase defines the interface for flight search operations.
type FlightSearchUseCase interface {
	// Search queries all registered providers and returns aggregated flight results.
	// It applies the Scatter-Gather pattern with timeout handling.
	Search(ctx context.Context, criteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error)
}

// flightSearchUseCase implements FlightSearchUseCase using the Scatter-Gather pattern.
type flightSearchUseCase struct {
	providers       []domain.FlightProvider
	globalTimeout   time.Duration
	providerTimeout time.Duration
}

// Config contains configuration options for the use case.
type Config struct {
	GlobalTimeout   time.Duration
	ProviderTimeout time.Duration
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		GlobalTimeout:   DefaultGlobalTimeout,
		ProviderTimeout: DefaultProviderTimeout,
	}
}

// NewFlightSearchUseCase creates a new FlightSearchUseCase with the given providers and configuration.
// If config is nil, default timeout values are used.
func NewFlightSearchUseCase(providers []domain.FlightProvider, config *Config) FlightSearchUseCase {
	cfg := DefaultConfig()
	if config != nil {
		if config.GlobalTimeout > 0 {
			cfg.GlobalTimeout = config.GlobalTimeout
		}
		if config.ProviderTimeout > 0 {
			cfg.ProviderTimeout = config.ProviderTimeout
		}
	}

	return &flightSearchUseCase{
		providers:       providers,
		globalTimeout:   cfg.GlobalTimeout,
		providerTimeout: cfg.ProviderTimeout,
	}
}

// providerResult holds the result from a single provider query.
type providerResult struct {
	Provider string
	Flights  []domain.Flight
	Error    error
	Duration time.Duration
}

// Search implements FlightSearchUseCase.Search using the Scatter-Gather pattern.
func (uc *flightSearchUseCase) Search(ctx context.Context, criteria domain.SearchCriteria, opts SearchOptions) (*domain.SearchResponse, error) {
	startTime := time.Now()

	// Handle case with no providers
	if len(uc.providers) == 0 {
		return nil, domain.ErrAllProvidersFailed
	}

	// Create context with global timeout
	ctx, cancel := context.WithTimeout(ctx, uc.globalTimeout)
	defer cancel()

	// Buffered channel to prevent goroutine blocking
	resultsChan := make(chan providerResult, len(uc.providers))

	// WaitGroup to track goroutine completion
	var wg sync.WaitGroup

	// Scatter: launch goroutines for each provider
	for _, provider := range uc.providers {
		wg.Add(1)
		go func(p domain.FlightProvider) {
			defer wg.Done()
			uc.queryProvider(ctx, p, criteria, resultsChan)
		}(provider)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Gather: collect results
	var allFlights []domain.Flight
	var failedProviders []string
	queriedProviders := make([]string, 0, len(uc.providers))

	for result := range resultsChan {
		queriedProviders = append(queriedProviders, result.Provider)
		if result.Error != nil {
			failedProviders = append(failedProviders, result.Provider)
			continue
		}
		allFlights = append(allFlights, result.Flights...)
	}

	// Check if context was cancelled before we got all results
	if ctx.Err() != nil && len(queriedProviders) < len(uc.providers) {
		// Record remaining providers as failed
		for _, p := range uc.providers {
			found := false
			for _, q := range queriedProviders {
				if q == p.Name() {
					found = true
					break
				}
			}
			if !found {
				queriedProviders = append(queriedProviders, p.Name())
				failedProviders = append(failedProviders, p.Name())
			}
		}
	}

	// Check if all providers failed
	if len(failedProviders) == len(uc.providers) {
		return nil, domain.ErrAllProvidersFailed
	}

	// Apply filtering (delegates to filter logic)
	filtered := applyFilters(allFlights, opts.Filters)

	// Calculate ranking scores (delegates to ranking logic)
	ranked := calculateRankingScores(filtered)

	// Sort results (delegates to sorting logic)
	sorted := sortFlights(ranked, opts.SortBy)

	// Build response
	response := &domain.SearchResponse{
		Flights: sorted,
		Metadata: domain.SearchMetadata{
			TotalResults:     len(sorted),
			SearchDurationMs: time.Since(startTime).Milliseconds(),
			ProvidersQueried: queriedProviders,
			ProvidersFailed:  failedProviders,
		},
	}

	return response, nil
}

// queryProvider queries a single provider with timeout and panic recovery.
func (uc *flightSearchUseCase) queryProvider(ctx context.Context, provider domain.FlightProvider, criteria domain.SearchCriteria, results chan<- providerResult) {
	// Per-provider timeout
	ctx, cancel := context.WithTimeout(ctx, uc.providerTimeout)
	defer cancel()

	start := time.Now()
	providerName := provider.Name()

	// Panic recovery to prevent one provider from crashing the whole search
	defer func() {
		if r := recover(); r != nil {
			results <- providerResult{
				Provider: providerName,
				Error:    fmt.Errorf("provider panic: %v", r),
				Duration: time.Since(start),
			}
		}
	}()

	flights, err := provider.Search(ctx, criteria)

	results <- providerResult{
		Provider: providerName,
		Flights:  flights,
		Error:    err,
		Duration: time.Since(start),
	}
}

// applyFilters applies filter options to the flight list.
// This is a stub that will be replaced by TKT-09 integration.
func applyFilters(flights []domain.Flight, opts *domain.FilterOptions) []domain.Flight {
	if opts == nil {
		return flights
	}

	result := make([]domain.Flight, 0, len(flights))
	for _, f := range flights {
		if opts.MatchesFlight(f) {
			result = append(result, f)
		}
	}
	return result
}

// calculateRankingScores calculates the ranking score for each flight.
// This is a stub that will be enhanced by TKT-10.
// For now, implements a simple ranking based on normalized price and duration.
func calculateRankingScores(flights []domain.Flight) []domain.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Find min/max for normalization
	minPrice, maxPrice := flights[0].Price.Amount, flights[0].Price.Amount
	minDuration, maxDuration := flights[0].Duration.TotalMinutes, flights[0].Duration.TotalMinutes

	for _, f := range flights {
		if f.Price.Amount < minPrice {
			minPrice = f.Price.Amount
		}
		if f.Price.Amount > maxPrice {
			maxPrice = f.Price.Amount
		}
		if f.Duration.TotalMinutes < minDuration {
			minDuration = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > maxDuration {
			maxDuration = f.Duration.TotalMinutes
		}
	}

	// Calculate scores (higher is better)
	priceRange := maxPrice - minPrice
	durationRange := maxDuration - minDuration

	result := make([]domain.Flight, len(flights))
	copy(result, flights)

	for i := range result {
		var priceScore, durationScore, stopsScore float64

		// Price score: cheaper is better (0-40 points)
		if priceRange > 0 {
			priceScore = (1 - (result[i].Price.Amount-minPrice)/priceRange) * 40
		} else {
			priceScore = 40
		}

		// Duration score: shorter is better (0-30 points)
		if durationRange > 0 {
			durationScore = (1 - float64(result[i].Duration.TotalMinutes-minDuration)/float64(durationRange)) * 30
		} else {
			durationScore = 30
		}

		// Stops score: fewer stops is better (0-30 points)
		switch result[i].Stops {
		case 0:
			stopsScore = 30
		case 1:
			stopsScore = 20
		case 2:
			stopsScore = 10
		default:
			stopsScore = 0
		}

		result[i].RankingScore = priceScore + durationScore + stopsScore
	}

	return result
}

// sortFlights sorts flights according to the specified sort option.
// This is a stub that will be enhanced by TKT-10.
func sortFlights(flights []domain.Flight, sortBy domain.SortOption) []domain.Flight {
	if len(flights) <= 1 {
		return flights
	}

	result := make([]domain.Flight, len(flights))
	copy(result, flights)

	switch sortBy {
	case domain.SortByPrice:
		sort.Slice(result, func(i, j int) bool {
			return result[i].Price.Amount < result[j].Price.Amount
		})
	case domain.SortByDuration:
		sort.Slice(result, func(i, j int) bool {
			return result[i].Duration.TotalMinutes < result[j].Duration.TotalMinutes
		})
	case domain.SortByDeparture:
		sort.Slice(result, func(i, j int) bool {
			return result[i].Departure.DateTime.Before(result[j].Departure.DateTime)
		})
	case domain.SortByBestValue:
		fallthrough
	default:
		// Sort by ranking score (highest first)
		sort.Slice(result, func(i, j int) bool {
			return result[i].RankingScore > result[j].RankingScore
		})
	}

	return result
}

// Ensure flightSearchUseCase implements FlightSearchUseCase at compile time.
var _ FlightSearchUseCase = (*flightSearchUseCase)(nil)
