package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// createTestFlight creates a flight for testing with the given parameters.
func createTestFlight(id, provider string, price float64, durationMin int, stops int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: "FL-" + id,
		Airline: domain.AirlineInfo{
			Code: "AA",
			Name: "Test Airline",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			DateTime:    time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC),
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			DateTime:    time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC),
		},
		Duration: domain.DurationInfo{
			TotalMinutes: durationMin,
			Formatted:    "2h 0m",
		},
		Price: domain.PriceInfo{
			Amount:   price,
			Currency: "IDR",
		},
		Baggage: domain.BaggageInfo{
			CabinKg:   7,
			CheckedKg: 20,
		},
		Class:    "economy",
		Stops:    stops,
		Provider: provider,
	}
}

// setupMockProvider creates a mock provider with standard behavior.
func setupMockProvider(ctrl *gomock.Controller, name string, flights []domain.Flight, err error) *domain.MockFlightProvider {
	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return(name).AnyTimes()
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).Return(flights, err).AnyTimes()
	return mock
}

// setupMockProviderWithDelay creates a mock provider that simulates network delay.
func setupMockProviderWithDelay(ctrl *gomock.Controller, name string, flights []domain.Flight, delay time.Duration) *domain.MockFlightProvider {
	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return(name).AnyTimes()
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
			select {
			case <-time.After(delay):
				return flights, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	).AnyTimes()
	return mock
}

// setupMockProviderWithPanic creates a mock provider that panics.
func setupMockProviderWithPanic(ctrl *gomock.Controller, name string, panicMsg string) *domain.MockFlightProvider {
	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return(name).AnyTimes()
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, criteria domain.SearchCriteria) ([]domain.Flight, error) {
			panic(panicMsg)
		},
	).AnyTimes()
	return mock
}

// TestNewFlightSearchUseCase tests the constructor.
func TestNewFlightSearchUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return("test").AnyTimes()

	tests := []struct {
		name      string
		providers []domain.FlightProvider
		config    *Config
	}{
		{
			name:      "with default config",
			providers: []domain.FlightProvider{mock},
			config:    nil,
		},
		{
			name:      "with custom config",
			providers: []domain.FlightProvider{mock},
			config: &Config{
				GlobalTimeout:   10 * time.Second,
				ProviderTimeout: 3 * time.Second,
			},
		},
		{
			name:      "with empty providers",
			providers: []domain.FlightProvider{},
			config:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewFlightSearchUseCase(tt.providers, tt.config)
			require.NotNil(t, uc)
		})
	}
}

// TestSearch_MultipleProvidersSuccess tests successful aggregation from multiple providers.
func TestSearch_MultipleProvidersSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights1 := []domain.Flight{
		createTestFlight("1", "provider1", 1000000, 120, 0),
		createTestFlight("2", "provider1", 1200000, 100, 1),
	}
	flights2 := []domain.Flight{
		createTestFlight("3", "provider2", 900000, 130, 0),
	}
	flights3 := []domain.Flight{
		createTestFlight("4", "provider3", 1100000, 110, 0),
		createTestFlight("5", "provider3", 950000, 140, 2),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "provider1", flights1, nil),
		setupMockProvider(ctrl, "provider2", flights2, nil),
		setupMockProvider(ctrl, "provider3", flights3, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Flights, 5)
	assert.Equal(t, 5, response.Metadata.TotalResults)
	assert.Len(t, response.Metadata.ProvidersQueried, 3)
	assert.Empty(t, response.Metadata.ProvidersFailed)
	assert.GreaterOrEqual(t, response.Metadata.SearchDurationMs, int64(0))
}

// TestSearch_PartialFailure tests graceful handling when some providers fail.
func TestSearch_PartialFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights1 := []domain.Flight{
		createTestFlight("1", "provider1", 1000000, 120, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "provider1", flights1, nil),
		setupMockProvider(ctrl, "provider2", nil, errors.New("provider error")),
		setupMockProvider(ctrl, "provider3", nil, errors.New("another error")),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Flights, 1) // Only from provider1
	assert.Len(t, response.Metadata.ProvidersQueried, 3)
	assert.Len(t, response.Metadata.ProvidersFailed, 2)
}

// TestSearch_AllProvidersFail tests error when all providers fail.
func TestSearch_AllProvidersFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "provider1", nil, errors.New("error1")),
		setupMockProvider(ctrl, "provider2", nil, errors.New("error2")),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, response)
}

// TestSearch_NoProviders tests error when no providers are registered.
func TestSearch_NoProviders(t *testing.T) {
	uc := NewFlightSearchUseCase([]domain.FlightProvider{}, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))
	assert.Nil(t, response)
}

// TestSearch_EmptyResults tests handling when providers return empty results.
func TestSearch_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "provider1", []domain.Flight{}, nil),
		setupMockProvider(ctrl, "provider2", []domain.Flight{}, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Empty(t, response.Flights)
	assert.Equal(t, 0, response.Metadata.TotalResults)
	assert.Empty(t, response.Metadata.ProvidersFailed)
}

// TestSearch_ProviderTimeout tests per-provider timeout behavior.
func TestSearch_ProviderTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fastFlights := []domain.Flight{
		createTestFlight("1", "fast", 1000000, 120, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProviderWithDelay(ctrl, "fast", fastFlights, 10*time.Millisecond),
		setupMockProviderWithDelay(ctrl, "slow", nil, 5*time.Second), // Will timeout
	}

	config := &Config{
		GlobalTimeout:   1 * time.Second,
		ProviderTimeout: 100 * time.Millisecond,
	}

	uc := NewFlightSearchUseCase(providers, config)
	ctx := context.Background()

	start := time.Now()
	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should complete faster than global timeout
	assert.Less(t, elapsed, 500*time.Millisecond)

	// Fast provider should succeed
	assert.Len(t, response.Flights, 1)
	assert.Contains(t, response.Metadata.ProvidersFailed, "slow")
}

// TestSearch_GlobalTimeout tests global timeout behavior.
func TestSearch_GlobalTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	providers := []domain.FlightProvider{
		setupMockProviderWithDelay(ctrl, "slow1", nil, 10*time.Second),
		setupMockProviderWithDelay(ctrl, "slow2", nil, 10*time.Second),
	}

	config := &Config{
		GlobalTimeout:   100 * time.Millisecond,
		ProviderTimeout: 5 * time.Second,
	}

	uc := NewFlightSearchUseCase(providers, config)
	ctx := context.Background()

	start := time.Now()
	_, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})
	elapsed := time.Since(start)

	// Should fail with all providers failed
	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrAllProvidersFailed))

	// Should complete around global timeout
	assert.Less(t, elapsed, 500*time.Millisecond)
}

// TestSearch_ContextCancellation tests context cancellation handling.
func TestSearch_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	providers := []domain.FlightProvider{
		setupMockProviderWithDelay(ctrl, "slow", nil, 5*time.Second),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

// TestSearch_ProviderPanic tests panic recovery in provider calls.
func TestSearch_ProviderPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	goodFlights := []domain.Flight{
		createTestFlight("1", "good", 1000000, 120, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "good", goodFlights, nil),
		setupMockProviderWithPanic(ctrl, "panicking", "something went wrong"),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Flights, 1)
	assert.Contains(t, response.Metadata.ProvidersFailed, "panicking")
}

// TestSearch_SingleProvider tests the scatter-gather pattern with a single provider.
func TestSearch_SingleProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights := []domain.Flight{
		createTestFlight("1", "single", 1000000, 120, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "single", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Flights, 1)
	assert.Len(t, response.Metadata.ProvidersQueried, 1)
}

// TestSearch_WithFiltering tests that filters are applied to results.
func TestSearch_WithFiltering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights := []domain.Flight{
		createTestFlight("1", "test", 500000, 120, 0),  // Below max price
		createTestFlight("2", "test", 1500000, 100, 0), // Above max price
		createTestFlight("3", "test", 800000, 110, 2),  // Above max stops
		createTestFlight("4", "test", 600000, 130, 0),  // Passes all filters
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	maxPrice := float64(1000000)
	maxStops := 0
	opts := SearchOptions{
		Filters: &domain.FilterOptions{
			MaxPrice: &maxPrice,
			MaxStops: &maxStops,
		},
	}

	response, err := uc.Search(ctx, domain.SearchCriteria{}, opts)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Flights, 2) // Only flights 1 and 4
}

// TestSearch_SortByPrice tests sorting by price.
func TestSearch_SortByPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights := []domain.Flight{
		createTestFlight("1", "test", 1500000, 120, 0),
		createTestFlight("2", "test", 500000, 100, 0),
		createTestFlight("3", "test", 1000000, 110, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	opts := SearchOptions{
		SortBy: domain.SortByPrice,
	}

	response, err := uc.Search(ctx, domain.SearchCriteria{}, opts)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.Flights, 3)

	// Should be sorted by price ascending
	assert.Equal(t, float64(500000), response.Flights[0].Price.Amount)
	assert.Equal(t, float64(1000000), response.Flights[1].Price.Amount)
	assert.Equal(t, float64(1500000), response.Flights[2].Price.Amount)
}

// TestSearch_SortByDuration tests sorting by duration.
func TestSearch_SortByDuration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights := []domain.Flight{
		createTestFlight("1", "test", 1000000, 180, 0),
		createTestFlight("2", "test", 1000000, 90, 0),
		createTestFlight("3", "test", 1000000, 120, 0),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	opts := SearchOptions{
		SortBy: domain.SortByDuration,
	}

	response, err := uc.Search(ctx, domain.SearchCriteria{}, opts)

	require.NoError(t, err)
	require.Len(t, response.Flights, 3)

	// Should be sorted by duration ascending
	assert.Equal(t, 90, response.Flights[0].Duration.TotalMinutes)
	assert.Equal(t, 120, response.Flights[1].Duration.TotalMinutes)
	assert.Equal(t, 180, response.Flights[2].Duration.TotalMinutes)
}

// TestSearch_SortByDeparture tests sorting by departure time.
func TestSearch_SortByDeparture(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f1 := createTestFlight("1", "test", 1000000, 120, 0)
	f1.Departure.DateTime = time.Date(2025, 12, 15, 18, 0, 0, 0, time.UTC)

	f2 := createTestFlight("2", "test", 1000000, 120, 0)
	f2.Departure.DateTime = time.Date(2025, 12, 15, 6, 0, 0, 0, time.UTC)

	f3 := createTestFlight("3", "test", 1000000, 120, 0)
	f3.Departure.DateTime = time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)

	flights := []domain.Flight{f1, f2, f3}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	opts := SearchOptions{
		SortBy: domain.SortByDeparture,
	}

	response, err := uc.Search(ctx, domain.SearchCriteria{}, opts)

	require.NoError(t, err)
	require.Len(t, response.Flights, 3)

	// Should be sorted by departure time ascending
	assert.Equal(t, 6, response.Flights[0].Departure.DateTime.Hour())
	assert.Equal(t, 12, response.Flights[1].Departure.DateTime.Hour())
	assert.Equal(t, 18, response.Flights[2].Departure.DateTime.Hour())
}

// TestSearch_SortByBestValue tests sorting by ranking score.
func TestSearch_SortByBestValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create flights with varying quality
	// Best value: cheap, short duration, direct
	best := createTestFlight("best", "test", 500000, 90, 0)

	// Worst value: expensive, long duration, many stops
	worst := createTestFlight("worst", "test", 2000000, 300, 3)

	// Middle: average price, duration, 1 stop
	middle := createTestFlight("middle", "test", 1000000, 150, 1)

	flights := []domain.Flight{worst, middle, best}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	opts := SearchOptions{
		SortBy: domain.SortByBestValue,
	}

	response, err := uc.Search(ctx, domain.SearchCriteria{}, opts)

	require.NoError(t, err)
	require.Len(t, response.Flights, 3)

	// Best value should be first
	assert.Equal(t, "best", response.Flights[0].ID)
	// Worst value should be last
	assert.Equal(t, "worst", response.Flights[2].ID)

	// All flights should have ranking scores calculated
	for _, f := range response.Flights {
		assert.GreaterOrEqual(t, f.RankingScore, float64(0))
	}
}

// TestSearch_RankingScoresCalculated tests that ranking scores are calculated.
func TestSearch_RankingScoresCalculated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	flights := []domain.Flight{
		createTestFlight("1", "test", 1000000, 120, 0),
		createTestFlight("2", "test", 1500000, 150, 1),
	}

	providers := []domain.FlightProvider{
		setupMockProvider(ctrl, "test", flights, nil),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})

	require.NoError(t, err)
	for _, f := range response.Flights {
		assert.Greater(t, f.RankingScore, float64(0))
	}
}

// TestSearch_ConcurrentProviderCalls verifies providers are called concurrently.
func TestSearch_ConcurrentProviderCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Each provider has a small delay
	delayMs := 50 * time.Millisecond

	providers := []domain.FlightProvider{
		setupMockProviderWithDelay(ctrl, "p1", []domain.Flight{createTestFlight("1", "p1", 1000000, 120, 0)}, delayMs),
		setupMockProviderWithDelay(ctrl, "p2", []domain.Flight{createTestFlight("2", "p2", 1100000, 110, 0)}, delayMs),
		setupMockProviderWithDelay(ctrl, "p3", []domain.Flight{createTestFlight("3", "p3", 900000, 130, 0)}, delayMs),
		setupMockProviderWithDelay(ctrl, "p4", []domain.Flight{createTestFlight("4", "p4", 1200000, 100, 0)}, delayMs),
	}

	uc := NewFlightSearchUseCase(providers, nil)
	ctx := context.Background()

	start := time.Now()
	response, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, response.Flights, 4)

	// If run sequentially, would take 4 * 50ms = 200ms
	// If run concurrently, should take ~50ms (plus overhead)
	// Allow up to 150ms for concurrent execution
	assert.Less(t, elapsed, 150*time.Millisecond, "Providers should be called concurrently")
}

// TestDefaultSearchOptions tests the default options factory.
func TestDefaultSearchOptions(t *testing.T) {
	opts := DefaultSearchOptions()

	assert.Nil(t, opts.Filters)
	assert.Equal(t, domain.SortByBestValue, opts.SortBy)
}

// TestDefaultConfig tests the default config factory.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, DefaultGlobalTimeout, cfg.GlobalTimeout)
	assert.Equal(t, DefaultProviderTimeout, cfg.ProviderTimeout)
}

// TestApplyFilters_Basic tests the filter application function directly.
// Note: Comprehensive filter tests are in filter_test.go
func TestApplyFilters_Basic(t *testing.T) {
	flights := []domain.Flight{
		createTestFlight("1", "test", 500000, 120, 0),
		createTestFlight("2", "test", 1500000, 100, 0),
		createTestFlight("3", "test", 800000, 110, 2),
	}

	// Helper for pointer values
	floatPtr := func(f float64) *float64 { return &f }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name     string
		opts     *domain.FilterOptions
		expected int
	}{
		{
			name:     "nil filters returns all",
			opts:     nil,
			expected: 3,
		},
		{
			name:     "empty filters returns all",
			opts:     &domain.FilterOptions{},
			expected: 3,
		},
		{
			name:     "max price filter",
			opts:     &domain.FilterOptions{MaxPrice: floatPtr(1000000)},
			expected: 2,
		},
		{
			name:     "max stops filter",
			opts:     &domain.FilterOptions{MaxStops: intPtr(0)},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilters(flights, tt.opts)
			assert.Len(t, result, tt.expected)
		})
	}
}

// TestCalculateRankingScores tests the ranking score calculation.
func TestCalculateRankingScores(t *testing.T) {
	tests := []struct {
		name           string
		flights        []domain.Flight
		expectedLen    int
		checkScoreFunc func(t *testing.T, result []domain.Flight)
	}{
		{
			name:        "empty flights",
			flights:     []domain.Flight{},
			expectedLen: 0,
		},
		{
			name: "single flight gets max scores",
			flights: []domain.Flight{
				createTestFlight("1", "test", 1000000, 120, 0),
			},
			expectedLen: 1,
			checkScoreFunc: func(t *testing.T, result []domain.Flight) {
				// Should get max score (40 + 30 + 30 = 100 for direct flight)
				assert.Equal(t, float64(100), result[0].RankingScore)
			},
		},
		{
			name: "cheaper flight scores higher on price",
			flights: []domain.Flight{
				createTestFlight("expensive", "test", 1500000, 120, 0),
				createTestFlight("cheap", "test", 500000, 120, 0),
			},
			expectedLen: 2,
			checkScoreFunc: func(t *testing.T, result []domain.Flight) {
				var cheapScore, expensiveScore float64
				for _, f := range result {
					if f.ID == "cheap" {
						cheapScore = f.RankingScore
					} else {
						expensiveScore = f.RankingScore
					}
				}
				assert.Greater(t, cheapScore, expensiveScore)
			},
		},
		{
			name: "direct flight scores higher on stops",
			flights: []domain.Flight{
				createTestFlight("two-stops", "test", 1000000, 120, 2),
				createTestFlight("one-stop", "test", 1000000, 120, 1),
				createTestFlight("direct", "test", 1000000, 120, 0),
			},
			expectedLen: 3,
			checkScoreFunc: func(t *testing.T, result []domain.Flight) {
				scores := make(map[string]float64)
				for _, f := range result {
					scores[f.ID] = f.RankingScore
				}
				assert.Greater(t, scores["direct"], scores["one-stop"])
				assert.Greater(t, scores["one-stop"], scores["two-stops"])
			},
		},
		{
			name: "shorter duration scores higher",
			flights: []domain.Flight{
				createTestFlight("long", "test", 1000000, 240, 0),
				createTestFlight("short", "test", 1000000, 60, 0),
			},
			expectedLen: 2,
			checkScoreFunc: func(t *testing.T, result []domain.Flight) {
				var shortScore, longScore float64
				for _, f := range result {
					if f.ID == "short" {
						shortScore = f.RankingScore
					} else {
						longScore = f.RankingScore
					}
				}
				assert.Greater(t, shortScore, longScore)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRankingScores(tt.flights)
			assert.Len(t, result, tt.expectedLen)
			if tt.checkScoreFunc != nil {
				tt.checkScoreFunc(t, result)
			}
		})
	}
}

// TestSortFlights tests the sorting functions directly.
func TestSortFlights(t *testing.T) {
	tests := []struct {
		name           string
		flights        []domain.Flight
		sortBy         domain.SortOption
		expectedLen    int
		expectedFirst  string
		checkOriginal  bool
	}{
		{
			name:          "empty slice",
			flights:       []domain.Flight{},
			sortBy:        domain.SortByPrice,
			expectedLen:   0,
			expectedFirst: "",
		},
		{
			name:          "single flight",
			flights:       []domain.Flight{createTestFlight("1", "test", 1000000, 120, 0)},
			sortBy:        domain.SortByPrice,
			expectedLen:   1,
			expectedFirst: "1",
		},
		{
			name: "sort by price ascending",
			flights: []domain.Flight{
				createTestFlight("2", "test", 2000000, 120, 0),
				createTestFlight("1", "test", 1000000, 120, 0),
			},
			sortBy:        domain.SortByPrice,
			expectedLen:   2,
			expectedFirst: "1",
			checkOriginal: true,
		},
		{
			name: "sort by duration ascending",
			flights: []domain.Flight{
				createTestFlight("2", "test", 1000000, 180, 0),
				createTestFlight("1", "test", 1000000, 90, 0),
			},
			sortBy:        domain.SortByDuration,
			expectedLen:   2,
			expectedFirst: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var originalFirst string
			if tt.checkOriginal && len(tt.flights) > 0 {
				originalFirst = tt.flights[0].ID
			}

			result := sortFlights(tt.flights, tt.sortBy)

			assert.Len(t, result, tt.expectedLen)
			if tt.expectedFirst != "" && len(result) > 0 {
				assert.Equal(t, tt.expectedFirst, result[0].ID)
			}
			if tt.checkOriginal {
				assert.Equal(t, originalFirst, tt.flights[0].ID, "original slice should not be modified")
			}
		})
	}
}

// TestSearch_VerifyCriteriaPassedCorrectly verifies criteria are passed to providers.
func TestSearch_VerifyCriteriaPassedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedCriteria := domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		Class:         "business",
	}

	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return("test").AnyTimes()
	// Use gomock.Eq to verify exact criteria match
	mock.EXPECT().Search(gomock.Any(), gomock.Eq(expectedCriteria)).Return([]domain.Flight{}, nil)

	uc := NewFlightSearchUseCase([]domain.FlightProvider{mock}, nil)
	ctx := context.Background()

	response, err := uc.Search(ctx, expectedCriteria, SearchOptions{})

	require.NoError(t, err)
	require.NotNil(t, response)
}

// TestSearch_VerifySearchCalledOnce verifies search is called exactly once per provider.
func TestSearch_VerifySearchCalledOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := domain.NewMockFlightProvider(ctrl)
	mock.EXPECT().Name().Return("test").AnyTimes()
	// Expect Search to be called exactly once
	mock.EXPECT().Search(gomock.Any(), gomock.Any()).Times(1).Return([]domain.Flight{}, nil)

	uc := NewFlightSearchUseCase([]domain.FlightProvider{mock}, nil)
	ctx := context.Background()

	_, err := uc.Search(ctx, domain.SearchCriteria{}, SearchOptions{})
	require.NoError(t, err)

	// gomock will automatically verify expectations when ctrl.Finish() is called
}
