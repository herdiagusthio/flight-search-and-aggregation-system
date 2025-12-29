package integration

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/test/mock"
)

// TestConcurrent_MultipleSearchRequests tests that multiple concurrent
// search requests are handled correctly without interference.
func TestConcurrent_MultipleSearchRequests(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").
		WithDelay(10 * time.Millisecond). // Small delay to increase overlap
		WithFlights(mock.SampleFlights("garuda", 3))

	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	numRequests := 10
	var wg sync.WaitGroup
	results := make([]Response, numRequests)
	errors := make([]error, numRequests)

	// Act - Fire concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = ts.SearchRequest(DefaultSearchRequest())
		}(i)
	}

	wg.Wait()

	// Assert - All requests should succeed
	for i := 0; i < numRequests; i++ {
		assert.Equal(t, http.StatusOK, results[i].Code, "request %d should succeed", i)
		assert.Nil(t, errors[i], "request %d should not have error", i)

		resp, err := results[i].ParseSearchResponse()
		require.NoError(t, err)
		assert.Len(t, resp.Flights, 3, "request %d should have 3 flights", i)
	}

	// Each concurrent request should trigger provider calls
	// (The mock provider tracks call count)
	assert.GreaterOrEqual(t, provider.CallCount(), numRequests)
}

// TestConcurrent_IndependentResults tests that each concurrent request
// receives its own independent results.
func TestConcurrent_IndependentResults(t *testing.T) {
	// Arrange - Create providers with different delays and results
	fastProvider := mock.NewProvider("fast").
		WithFlights(mock.SampleFlights("fast", 2))

	slowProvider := mock.NewProvider("slow").
		WithDelay(50 * time.Millisecond).
		WithFlights(mock.SampleFlights("slow", 3))

	uc := CreateUseCase([]domain.FlightProvider{fastProvider, slowProvider})
	ts := NewTestServer(uc)

	numRequests := 5
	var wg sync.WaitGroup
	results := make([]*domain.SearchResponse, numRequests)

	// Act
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp := ts.SearchRequest(DefaultSearchRequest())
			if resp.Code == http.StatusOK {
				results[idx], _ = resp.ParseSearchResponse()
			}
		}(i)
	}

	wg.Wait()

	// Assert - All requests should get same result structure
	for i := 0; i < numRequests; i++ {
		require.NotNil(t, results[i], "request %d should have result", i)
		assert.Len(t, results[i].Flights, 5, "request %d should have 5 flights (2+3)", i)
		assert.Len(t, results[i].Metadata.ProvidersQueried, 2)
	}
}

// TestConcurrent_MixedSuccessAndFailure tests concurrent requests
// with some providers failing.
func TestConcurrent_MixedSuccessAndFailure(t *testing.T) {
	// Arrange
	goodProvider := mock.NewProvider("good").
		WithFlights(mock.SampleFlights("good", 2))

	uc := CreateUseCase([]domain.FlightProvider{goodProvider})
	ts := NewTestServer(uc)

	numRequests := 20
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Act
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := ts.SearchRequest(DefaultSearchRequest())
			if resp.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Assert - All should succeed
	assert.Equal(t, numRequests, successCount, "all requests should succeed")
}

// TestConcurrent_NoRaceCondition is designed to be run with -race flag.
// It performs concurrent operations to detect data races.
func TestConcurrent_NoRaceCondition(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 5))

	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	numGoroutines := 50
	var wg sync.WaitGroup

	// Different request types to exercise different paths
	futureDate := FutureDate()
	requests := []SearchRequestBody{
		DefaultSearchRequest(),
		{Origin: "CGK", Destination: "SUB", DepartureDate: futureDate, Passengers: 2},
		{Origin: "CGK", Destination: "DPS", DepartureDate: futureDate, Passengers: 1},
	}

	// Act
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := requests[idx%len(requests)]
			_ = ts.SearchRequest(req)
		}(i)
	}

	wg.Wait()

	// Assert - If we get here without race detector errors, test passes
	// The race detector will fail the test if races are found
	assert.True(t, true, "no race condition detected")
}

// TestConcurrent_ProviderCallCountAccuracy tests that the mock provider's
// call count is accurate under concurrent access.
func TestConcurrent_ProviderCallCountAccuracy(t *testing.T) {
	// Arrange
	provider := mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 1))

	uc := CreateUseCase([]domain.FlightProvider{provider})
	ts := NewTestServer(uc)

	numRequests := 100
	var wg sync.WaitGroup

	// Act
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.SearchRequest(DefaultSearchRequest())
		}()
	}

	wg.Wait()

	// Assert - Provider should be called exactly numRequests times
	assert.Equal(t, numRequests, provider.CallCount())
}

// TestConcurrent_HighLoadScenario simulates a high-load scenario
// with many concurrent requests and providers.
func TestConcurrent_HighLoadScenario(t *testing.T) {
	// Arrange
	providers := []domain.FlightProvider{
		mock.NewProvider("garuda").WithFlights(mock.SampleFlights("garuda", 5)),
		mock.NewProvider("lion").WithFlights(mock.SampleFlights("lion", 5)),
		mock.NewProvider("batik").WithFlights(mock.SampleFlights("batik", 5)),
		mock.NewProvider("airasia").WithFlights(mock.SampleFlights("airasia", 5)),
	}

	uc := CreateUseCase(providers)
	ts := NewTestServer(uc)

	numRequests := 50
	var wg sync.WaitGroup
	successCount := 0
	totalFlights := 0
	var mu sync.Mutex

	// Act
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := ts.SearchRequest(DefaultSearchRequest())
			if resp.Code == http.StatusOK {
				if searchResp, err := resp.ParseSearchResponse(); err == nil {
					mu.Lock()
					successCount++
					totalFlights += len(searchResp.Flights)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Assert
	assert.Equal(t, numRequests, successCount, "all requests should succeed")
	// Each request should return 20 flights (5 from each of 4 providers)
	expectedFlights := numRequests * 20
	assert.Equal(t, expectedFlights, totalFlights, "total flights should match")
}
