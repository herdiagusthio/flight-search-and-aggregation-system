package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

//go:generate mockgen -destination=flight_search_mock.go -package=usecase github.com/flight-search/flight-search-and-aggregation-system/internal/usecase FlightSearchUseCase

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

	// Apply filtering using the dedicated filter module
	filtered := ApplyFilters(allFlights, opts.Filters)

	// Calculate ranking scores using the dedicated ranking module
	ranked := CalculateRankingScores(filtered)

	// Sort results using the dedicated sorting module
	sorted := SortFlights(ranked, opts.SortBy)

	// Build response with new format
	successfulProviders := len(uc.providers) - len(failedProviders)
	response := domain.NewSearchResponse(
		&criteria,
		sorted,
		domain.SearchMetadata{
			TotalResults:       len(sorted),
			ProvidersQueried:   len(uc.providers),
			ProvidersSucceeded: successfulProviders,
			ProvidersFailed:    len(failedProviders),
			SearchTimeMs:       time.Since(startTime).Milliseconds(),
			CacheHit:           false, // Not implemented yet
		},
	)

	return &response, nil
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

// Ensure flightSearchUseCase implements FlightSearchUseCase at compile time.
var _ FlightSearchUseCase = (*flightSearchUseCase)(nil)
