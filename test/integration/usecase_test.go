package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
	"github.com/flight-search/flight-search-and-aggregation-system/test/mock"
)

// TestFlightSearch_MultipleProviders_Success tests that the use case
// successfully aggregates results from multiple providers.
func TestFlightSearch_MultipleProviders_Success(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 2))
	provider2 := mock.NewProvider("lion").WithFlights(mock.SampleFlights("lion", 3))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Flights, 5) // 2 + 3

	// Verify metadata
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 2, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)
	assert.Equal(t, 5, result.Metadata.TotalResults)

	// Verify both providers were called
	assert.Equal(t, 1, provider1.CallCount())
	assert.Equal(t, 1, provider2.CallCount())
}

// TestFlightSearch_PartialFailure tests that the use case returns
// partial results when some providers fail.
func TestFlightSearch_PartialFailure(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 2))
	provider2 := mock.NewProvider("lion").WithError(errors.New("connection refused"))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert - Should succeed with partial results
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Flights, 2) // Only garuda's flights

	// Verify metadata
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)
}

// TestFlightSearch_AllProvidersFail tests that the use case returns
// an error when all providers fail.
func TestFlightSearch_AllProvidersFail(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithError(errors.New("network error"))
	provider2 := mock.NewProvider("lion").WithError(errors.New("timeout"))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, result)
}

// TestFlightSearch_ProviderTimeout tests that slow providers are timed out.
func TestFlightSearch_ProviderTimeout(t *testing.T) {
	// Arrange - Provider takes longer than provider timeout
	slowProvider := mock.NewProvider("slow").
		WithDelay(500 * time.Millisecond).
		WithFlights(mock.SampleFlights("slow", 1))

	config := &usecase.Config{
		GlobalTimeout:   2 * time.Second,
		ProviderTimeout: 100 * time.Millisecond, // Very short timeout
	}

	uc := CreateUseCaseWithConfig([]domain.FlightProvider{slowProvider}, config)

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert - Should fail because provider timed out
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, result)
}

// TestFlightSearch_GlobalTimeout tests global timeout behavior.
func TestFlightSearch_GlobalTimeout(t *testing.T) {
	// Arrange - Multiple slow providers
	provider1 := mock.NewProvider("slow1").
		WithDelay(300 * time.Millisecond).
		WithFlights(mock.SampleFlights("slow1", 1))
	provider2 := mock.NewProvider("slow2").
		WithDelay(300 * time.Millisecond).
		WithFlights(mock.SampleFlights("slow2", 1))

	config := &usecase.Config{
		GlobalTimeout:   100 * time.Millisecond, // Very short global timeout
		ProviderTimeout: 1 * time.Second,
	}

	uc := CreateUseCaseWithConfig([]domain.FlightProvider{provider1, provider2}, config)

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert - Should fail because global timeout exceeded
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, result)
}

// TestFlightSearch_ContextCancellation tests that context cancellation is respected.
func TestFlightSearch_ContextCancellation(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").
		WithDelay(1 * time.Second).
		WithFlights(mock.SampleFlights("garuda", 1))

	uc := CreateUseCase([]domain.FlightProvider{provider})

	criteria := DefaultSearchCriteria()

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Act
	result, err := uc.Search(ctx, criteria, usecase.SearchOptions{})

	// Assert - Should fail due to cancellation
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestFlightSearch_FilterIntegration tests that filters are applied correctly.
func TestFlightSearch_FilterIntegration(t *testing.T) {
	// Arrange - Create flights with different prices
	flights := []domain.Flight{
		{ID: "1", FlightNumber: "GA 100", Price: domain.PriceInfo{Amount: 500000, Currency: "IDR"}, Provider: "garuda"},
		{ID: "2", FlightNumber: "GA 101", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Provider: "garuda"},
		{ID: "3", FlightNumber: "GA 102", Price: domain.PriceInfo{Amount: 1500000, Currency: "IDR"}, Provider: "garuda"},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})

	criteria := DefaultSearchCriteria()

	maxPrice := 1000000.0
	opts := usecase.SearchOptions{
		Filters: &domain.FilterOptions{
			MaxPrice: &maxPrice,
		},
	}

	// Act
	result, err := uc.Search(context.Background(), criteria, opts)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Flights, 2) // Only flights with price <= 1,000,000

	for _, f := range result.Flights {
		assert.LessOrEqual(t, f.Price.Amount, maxPrice)
	}
}

// TestFlightSearch_SortingIntegration tests that sorting is applied correctly.
func TestFlightSearch_SortingIntegration(t *testing.T) {
	// Arrange - Create flights with different prices
	flights := []domain.Flight{
		{ID: "1", FlightNumber: "GA 100", Price: domain.PriceInfo{Amount: 1500000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 150}, Provider: "garuda"},
		{ID: "2", FlightNumber: "GA 101", Price: domain.PriceInfo{Amount: 500000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 180}, Provider: "garuda"},
		{ID: "3", FlightNumber: "GA 102", Price: domain.PriceInfo{Amount: 1000000, Currency: "IDR"}, Duration: domain.DurationInfo{TotalMinutes: 120}, Provider: "garuda"},
	}

	provider := mock.NewProvider("garuda").WithFlights(flights)
	uc := CreateUseCase([]domain.FlightProvider{provider})

	criteria := DefaultSearchCriteria()

	opts := usecase.SearchOptions{
		SortBy: domain.SortByPrice,
	}

	// Act
	result, err := uc.Search(context.Background(), criteria, opts)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Flights, 3)

	// Verify sorted by price ascending
	assert.Equal(t, 500000.0, result.Flights[0].Price.Amount)
	assert.Equal(t, 1000000.0, result.Flights[1].Price.Amount)
	assert.Equal(t, 1500000.0, result.Flights[2].Price.Amount)
}

// TestFlightSearch_NoProviders tests behavior with no providers configured.
func TestFlightSearch_NoProviders(t *testing.T) {
	// Arrange
	uc := CreateUseCase([]domain.FlightProvider{})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, result)
}

// TestFlightSearch_EmptyResults tests behavior when providers return no flights.
func TestFlightSearch_EmptyResults(t *testing.T) {
	// Arrange - Provider returns empty slice (no error)
	provider := mock.NewProvider("garuda").WithFlights([]domain.Flight{})

	uc := CreateUseCase([]domain.FlightProvider{provider})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert - Should succeed with empty results
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Flights)
	assert.Equal(t, 0, result.Metadata.TotalResults)
}

// TestFlightSearch_MixedProviderResults tests providers with varying result counts.
func TestFlightSearch_MixedProviderResults(t *testing.T) {
	// Arrange
	provider1 := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 5))
	provider2 := mock.NewProvider("lion").WithFlights([]domain.Flight{}) // No flights
	provider3 := mock.NewProvider("batik").WithFlights(mock.SampleFlights("batik", 3))
	provider4 := mock.NewProvider("airasia").WithError(errors.New("unavailable"))

	uc := CreateUseCase([]domain.FlightProvider{provider1, provider2, provider3, provider4})

	criteria := DefaultSearchCriteria()

	// Act
	result, err := uc.Search(context.Background(), criteria, usecase.SearchOptions{})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Flights, 8) // 5 + 0 + 3

	// All four providers should be queried
	assert.Equal(t, 4, result.Metadata.ProvidersQueried)
	assert.Equal(t, 3, result.Metadata.ProvidersSucceeded)
	// Only airasia should have failed
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)
}
