package usecase

import (
	"math"
	"testing"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRankingTestFlight creates a flight for ranking/sorting testing.
func createRankingTestFlight(id string, price float64, durationMinutes int, stops int, departureHour int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: "FL-" + id,
		Airline: domain.AirlineInfo{
			Code: "GA",
			Name: "Test Airline",
		},
		Departure: domain.FlightPoint{
			AirportCode: "CGK",
			DateTime:    time.Date(2025, 12, 15, departureHour, 0, 0, 0, time.UTC),
		},
		Arrival: domain.FlightPoint{
			AirportCode: "DPS",
			DateTime:    time.Date(2025, 12, 15, departureHour+durationMinutes/60, durationMinutes%60, 0, 0, time.UTC),
		},
		Duration: domain.DurationInfo{
			TotalMinutes: durationMinutes,
			Formatted:    formatDuration(durationMinutes),
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
		Provider: "test",
	}
}

// formatDuration formats minutes to a human-readable string.
func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	if h > 0 && m > 0 {
		return string(rune('0'+h)) + "h " + string(rune('0'+m/10)) + string(rune('0'+m%10)) + "m"
	} else if h > 0 {
		return string(rune('0'+h)) + "h 0m"
	}
	return string(rune('0'+m/10)) + string(rune('0'+m%10)) + "m"
}

// =====================================================
// CalculateRankingScores Tests
// =====================================================

func TestCalculateRankingScores_Empty(t *testing.T) {
	result := CalculateRankingScores([]domain.Flight{})
	assert.Empty(t, result)
}

func TestCalculateRankingScores_SingleFlight(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 1000000, 120, 0, 8),
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 1)
	// Single flight: all normalized values are 0 (min == max), so score = 0
	assert.Equal(t, float64(0), result[0].RankingScore)
}

func TestCalculateRankingScores_AllEqual(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 1000000, 120, 0, 8),
		createRankingTestFlight("2", 1000000, 120, 0, 10),
		createRankingTestFlight("3", 1000000, 120, 0, 12),
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 3)
	// All equal values: all scores should be 0
	for _, f := range result {
		assert.Equal(t, float64(0), f.RankingScore, "Flight %s should have score 0", f.ID)
	}
}

func TestCalculateRankingScores_PriceVariation(t *testing.T) {
	// Only price varies; duration and stops are equal
	flights := []domain.Flight{
		createRankingTestFlight("cheap", 500000, 120, 0, 8),      // Normalized: 0
		createRankingTestFlight("expensive", 1500000, 120, 0, 8), // Normalized: 1
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 2)

	var cheapScore, expensiveScore float64
	for _, f := range result {
		if f.ID == "cheap" {
			cheapScore = f.RankingScore
		} else {
			expensiveScore = f.RankingScore
		}
	}

	// Cheap flight should have lower score (better)
	assert.Less(t, cheapScore, expensiveScore)
	// With only price varying: cheap = 0.5*0 = 0, expensive = 0.5*1 = 0.5
	assert.Equal(t, float64(0), cheapScore)
	assert.Equal(t, float64(0.5), expensiveScore)
}

func TestCalculateRankingScores_DurationVariation(t *testing.T) {
	// Only duration varies; price and stops are equal
	flights := []domain.Flight{
		createRankingTestFlight("short", 1000000, 90, 0, 8),  // Normalized: 0
		createRankingTestFlight("long", 1000000, 150, 0, 8),  // Normalized: 1
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 2)

	var shortScore, longScore float64
	for _, f := range result {
		if f.ID == "short" {
			shortScore = f.RankingScore
		} else {
			longScore = f.RankingScore
		}
	}

	// Shorter flight should have lower score (better)
	assert.Less(t, shortScore, longScore)
	// With only duration varying: short = 0.3*0 = 0, long = 0.3*1 = 0.3
	assert.Equal(t, float64(0), shortScore)
	assert.Equal(t, float64(0.3), longScore)
}

func TestCalculateRankingScores_StopsVariation(t *testing.T) {
	// Only stops varies; price and duration are equal
	flights := []domain.Flight{
		createRankingTestFlight("direct", 1000000, 120, 0, 8), // Normalized: 0
		createRankingTestFlight("stops", 1000000, 120, 2, 8),  // Normalized: 1
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 2)

	var directScore, stopsScore float64
	for _, f := range result {
		if f.ID == "direct" {
			directScore = f.RankingScore
		} else {
			stopsScore = f.RankingScore
		}
	}

	// Direct flight should have lower score (better)
	assert.Less(t, directScore, stopsScore)
	// With only stops varying: direct = 0.2*0 = 0, stops = 0.2*1 = 0.2
	assert.Equal(t, float64(0), directScore)
	assert.Equal(t, float64(0.2), stopsScore)
}

func TestCalculateRankingScores_ComplexScenario(t *testing.T) {
	// Example from the ticket
	// Flight A: Price 800,000 | Duration 105 | Stops 0
	// Flight B: Price 1,200,000 | Duration 90 | Stops 0
	// Flight C: Price 650,000 | Duration 150 | Stops 1
	flights := []domain.Flight{
		createRankingTestFlight("A", 800000, 105, 0, 8),
		createRankingTestFlight("B", 1200000, 90, 0, 10),
		createRankingTestFlight("C", 650000, 150, 1, 12),
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 3)

	scores := make(map[string]float64)
	for _, f := range result {
		scores[f.ID] = f.RankingScore
	}

	// Price range: 650,000 - 1,200,000 (550,000)
	// Duration range: 90 - 150 (60)
	// Stops range: 0 - 1 (1)

	// Flight A: price=(800k-650k)/550k=0.27, duration=(105-90)/60=0.25, stops=0/1=0
	// Score A = 0.5*0.27 + 0.3*0.25 + 0.2*0 ≈ 0.135 + 0.075 + 0 = 0.21
	assert.InDelta(t, 0.21, scores["A"], 0.01, "Flight A score")

	// Flight B: price=(1.2M-650k)/550k=1.0, duration=(90-90)/60=0, stops=0/1=0
	// Score B = 0.5*1.0 + 0.3*0 + 0.2*0 = 0.5
	assert.InDelta(t, 0.50, scores["B"], 0.01, "Flight B score")

	// Flight C: price=(650k-650k)/550k=0, duration=(150-90)/60=1.0, stops=1/1=1
	// Score C = 0.5*0 + 0.3*1.0 + 0.2*1.0 = 0.5
	assert.InDelta(t, 0.50, scores["C"], 0.01, "Flight C score")

	// Flight A should be best (lowest score)
	assert.Less(t, scores["A"], scores["B"])
	assert.Less(t, scores["A"], scores["C"])
}

func TestCalculateRankingScores_DoesNotMutateOriginal(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 500000, 90, 0, 8),
		createRankingTestFlight("2", 1000000, 120, 1, 10),
	}

	originalScore1 := flights[0].RankingScore
	originalScore2 := flights[1].RankingScore

	result := CalculateRankingScores(flights)

	// Original flights should not be modified
	assert.Equal(t, originalScore1, flights[0].RankingScore)
	assert.Equal(t, originalScore2, flights[1].RankingScore)

	// Result should have scores calculated
	// Flight 2 will have a non-zero score (it has higher price, longer duration, more stops)
	assert.Greater(t, result[1].RankingScore, float64(0), "Second flight should have non-zero score")
}

func TestCalculateRankingScores_VeryLargeValues(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 1, 1, 0, 8),
		createRankingTestFlight("2", 1e12, 10000, 10, 8),
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 2)

	var score1, score2 float64
	for _, f := range result {
		if f.ID == "1" {
			score1 = f.RankingScore
		} else {
			score2 = f.RankingScore
		}
	}

	// Flight 1 should have score 0 (all minimums)
	assert.Equal(t, float64(0), score1)
	// Flight 2 should have score 1.0 (all maximums: 0.5+0.3+0.2)
	assert.Equal(t, float64(1.0), score2)
}

func TestCalculateRankingScores_ZeroValues(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 0, 0, 0, 8),
		createRankingTestFlight("2", 100, 100, 1, 10),
	}

	result := CalculateRankingScores(flights)

	require.Len(t, result, 2)
	// Should handle zero values without issues
	assert.False(t, math.IsNaN(result[0].RankingScore))
	assert.False(t, math.IsNaN(result[1].RankingScore))
}

// =====================================================
// Normalization Tests
// =====================================================

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"minimum value", 0, 0, 100, 0},
		{"maximum value", 100, 0, 100, 1},
		{"middle value", 50, 0, 100, 0.5},
		{"quarter value", 25, 0, 100, 0.25},
		{"min equals max", 50, 50, 50, 0},
		{"negative range", -50, -100, 0, 0.5},
		{"large values", 5e11, 0, 1e12, 0.5},
		{"small decimal", 0.005, 0, 0.01, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeValue(tt.value, tt.min, tt.max)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestNormalizeValue_AvoidDivisionByZero(t *testing.T) {
	// When min == max, should return 0 instead of NaN/Inf
	result := normalizeValue(100, 100, 100)
	assert.Equal(t, float64(0), result)
	assert.False(t, math.IsNaN(result))
	assert.False(t, math.IsInf(result, 0))
}

// =====================================================
// Range Finding Tests
// =====================================================

func TestFindPriceRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "empty flights",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "single flight",
			flights: []domain.Flight{
				createRankingTestFlight("1", 500000, 120, 0, 8),
			},
			expectedMin: 500000,
			expectedMax: 500000,
		},
		{
			name: "multiple flights",
			flights: []domain.Flight{
				createRankingTestFlight("1", 800000, 120, 0, 8),
				createRankingTestFlight("2", 500000, 120, 0, 8),
				createRankingTestFlight("3", 1200000, 120, 0, 8),
			},
			expectedMin: 500000,
			expectedMax: 1200000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findPriceRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

func TestFindDurationRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin int
		expectedMax int
	}{
		{
			name:        "empty flights",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "single flight",
			flights: []domain.Flight{
				createRankingTestFlight("1", 500000, 120, 0, 8),
			},
			expectedMin: 120,
			expectedMax: 120,
		},
		{
			name: "multiple flights",
			flights: []domain.Flight{
				createRankingTestFlight("1", 500000, 150, 0, 8),
				createRankingTestFlight("2", 500000, 90, 0, 8),
				createRankingTestFlight("3", 500000, 120, 0, 8),
			},
			expectedMin: 90,
			expectedMax: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findDurationRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

func TestFindStopsRange(t *testing.T) {
	tests := []struct {
		name        string
		flights     []domain.Flight
		expectedMin int
		expectedMax int
	}{
		{
			name:        "empty flights",
			flights:     []domain.Flight{},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "all direct",
			flights: []domain.Flight{
				createRankingTestFlight("1", 500000, 120, 0, 8),
				createRankingTestFlight("2", 500000, 120, 0, 10),
			},
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name: "mixed stops",
			flights: []domain.Flight{
				createRankingTestFlight("1", 500000, 120, 2, 8),
				createRankingTestFlight("2", 500000, 120, 0, 8),
				createRankingTestFlight("3", 500000, 120, 1, 8),
			},
			expectedMin: 0,
			expectedMax: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findStopsRange(tt.flights)
			assert.Equal(t, tt.expectedMin, min)
			assert.Equal(t, tt.expectedMax, max)
		})
	}
}

// =====================================================
// SortFlights Tests
// =====================================================

func TestSortFlights_Empty(t *testing.T) {
	result := SortFlights([]domain.Flight{}, domain.SortByPrice)
	assert.Empty(t, result)
}

func TestSortFlights_SingleFlight(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("1", 1000000, 120, 0, 8),
	}

	result := SortFlights(flights, domain.SortByPrice)

	require.Len(t, result, 1)
	assert.Equal(t, "1", result[0].ID)
}

func TestSortFlights_ByPrice(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("expensive", 1500000, 120, 0, 8),
		createRankingTestFlight("cheap", 500000, 120, 0, 8),
		createRankingTestFlight("medium", 1000000, 120, 0, 8),
	}

	result := SortFlights(flights, domain.SortByPrice)

	require.Len(t, result, 3)
	assert.Equal(t, "cheap", result[0].ID)
	assert.Equal(t, "medium", result[1].ID)
	assert.Equal(t, "expensive", result[2].ID)
}

func TestSortFlights_ByDuration(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("long", 1000000, 180, 0, 8),
		createRankingTestFlight("short", 1000000, 60, 0, 8),
		createRankingTestFlight("medium", 1000000, 120, 0, 8),
	}

	result := SortFlights(flights, domain.SortByDuration)

	require.Len(t, result, 3)
	assert.Equal(t, "short", result[0].ID)
	assert.Equal(t, "medium", result[1].ID)
	assert.Equal(t, "long", result[2].ID)
}

func TestSortFlights_ByDeparture(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("evening", 1000000, 120, 0, 18),
		createRankingTestFlight("morning", 1000000, 120, 0, 6),
		createRankingTestFlight("noon", 1000000, 120, 0, 12),
	}

	result := SortFlights(flights, domain.SortByDeparture)

	require.Len(t, result, 3)
	assert.Equal(t, "morning", result[0].ID)
	assert.Equal(t, "noon", result[1].ID)
	assert.Equal(t, "evening", result[2].ID)
}

func TestSortFlights_ByBestValue(t *testing.T) {
	// Pre-calculate ranking scores
	flights := []domain.Flight{
		createRankingTestFlight("worst", 1500000, 180, 2, 8),
		createRankingTestFlight("best", 500000, 60, 0, 8),
		createRankingTestFlight("medium", 1000000, 120, 1, 8),
	}

	// Calculate scores first
	scored := CalculateRankingScores(flights)

	// Sort by best value
	result := SortFlights(scored, domain.SortByBestValue)

	require.Len(t, result, 3)
	// Best (lowest score) should be first
	assert.Equal(t, "best", result[0].ID)
	assert.Equal(t, "medium", result[1].ID)
	assert.Equal(t, "worst", result[2].ID)
}

func TestSortFlights_DefaultSort(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("high-score", 1000000, 120, 0, 8),
		createRankingTestFlight("low-score", 1000000, 120, 0, 10),
	}

	// Set different scores manually
	flights[0].RankingScore = 0.8
	flights[1].RankingScore = 0.2

	// Empty string should default to SortByBestValue
	result := SortFlights(flights, "")

	require.Len(t, result, 2)
	// Lower score should be first
	assert.Equal(t, "low-score", result[0].ID)
	assert.Equal(t, "high-score", result[1].ID)
}

func TestSortFlights_InvalidSortOption(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("high-score", 1000000, 120, 0, 8),
		createRankingTestFlight("low-score", 1000000, 120, 0, 10),
	}

	flights[0].RankingScore = 0.8
	flights[1].RankingScore = 0.2

	// Invalid option should default to SortByBestValue
	result := SortFlights(flights, domain.SortOption("invalid"))

	require.Len(t, result, 2)
	assert.Equal(t, "low-score", result[0].ID)
}

func TestSortFlights_StableSort(t *testing.T) {
	// Create flights with equal sort values
	flights := []domain.Flight{
		createRankingTestFlight("first", 1000000, 120, 0, 8),
		createRankingTestFlight("second", 1000000, 120, 0, 8),
		createRankingTestFlight("third", 1000000, 120, 0, 8),
	}

	// Sort multiple times - order should be preserved for equal values
	result1 := SortFlights(flights, domain.SortByPrice)
	result2 := SortFlights(flights, domain.SortByPrice)

	assert.Equal(t, result1[0].ID, result2[0].ID)
	assert.Equal(t, result1[1].ID, result2[1].ID)
	assert.Equal(t, result1[2].ID, result2[2].ID)

	// Original order should be maintained: first, second, third
	assert.Equal(t, "first", result1[0].ID)
	assert.Equal(t, "second", result1[1].ID)
	assert.Equal(t, "third", result1[2].ID)
}

func TestSortFlights_DoesNotMutateOriginal(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("z", 1500000, 120, 0, 8),
		createRankingTestFlight("a", 500000, 120, 0, 8),
	}

	originalFirstID := flights[0].ID

	result := SortFlights(flights, domain.SortByPrice)

	// Original slice should not be modified
	assert.Equal(t, originalFirstID, flights[0].ID)
	// Result should be sorted differently
	assert.Equal(t, "a", result[0].ID)
}

func TestSortFlights_AllSortOptions(t *testing.T) {
	tests := []struct {
		name          string
		sortBy        domain.SortOption
		expectedFirst string
	}{
		{
			name:          "sort by best value",
			sortBy:        domain.SortByBestValue,
			expectedFirst: "best",
		},
		{
			name:          "sort by price",
			sortBy:        domain.SortByPrice,
			expectedFirst: "cheapest",
		},
		{
			name:          "sort by duration",
			sortBy:        domain.SortByDuration,
			expectedFirst: "shortest",
		},
		{
			name:          "sort by departure",
			sortBy:        domain.SortByDeparture,
			expectedFirst: "earliest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flights := []domain.Flight{
				createRankingTestFlight("other", 1000000, 120, 1, 12),
				createRankingTestFlight("best", 800000, 100, 0, 10),       // Best score
				createRankingTestFlight("cheapest", 500000, 150, 2, 14),   // Lowest price
				createRankingTestFlight("shortest", 1200000, 60, 1, 16),   // Shortest duration
				createRankingTestFlight("earliest", 900000, 110, 0, 6),    // Earliest departure
			}

			// Calculate scores for best value sorting
			scored := CalculateRankingScores(flights)

			result := SortFlights(scored, tt.sortBy)

			require.Len(t, result, 5)
			assert.Equal(t, tt.expectedFirst, result[0].ID)
		})
	}
}

// =====================================================
// Integration Test: Calculate + Sort
// =====================================================

func TestCalculateAndSort_Integration(t *testing.T) {
	flights := []domain.Flight{
		createRankingTestFlight("A", 800000, 105, 0, 8),
		createRankingTestFlight("B", 1200000, 90, 0, 10),
		createRankingTestFlight("C", 650000, 150, 1, 12),
	}

	// Calculate ranking scores
	scored := CalculateRankingScores(flights)

	// Sort by best value
	sorted := SortFlights(scored, domain.SortByBestValue)

	require.Len(t, sorted, 3)

	// Flight A should be best (lowest score ~0.21)
	assert.Equal(t, "A", sorted[0].ID)

	// B and C have same score (~0.50), stable sort maintains original order
	// But we need to verify scores are reasonable
	assert.InDelta(t, 0.21, sorted[0].RankingScore, 0.01)
}

// =====================================================
// Performance Test
// =====================================================

func TestCalculateRankingScores_Performance(t *testing.T) {
	// Create a large list of flights
	n := 10000
	flights := make([]domain.Flight, n)
	for i := 0; i < n; i++ {
		flights[i] = createRankingTestFlight(
			"id-"+intToStr(i),
			float64(500000+i*100),
			60+i%120,
			i%4,
			6+i%12,
		)
	}

	start := time.Now()
	result := CalculateRankingScores(flights)
	elapsed := time.Since(start)

	assert.Len(t, result, n)
	// Should complete in under 100ms for 10k flights
	assert.Less(t, elapsed, 100*time.Millisecond, "Ranking calculation should be O(n)")
}

func TestSortFlights_Performance(t *testing.T) {
	n := 10000
	flights := make([]domain.Flight, n)
	for i := 0; i < n; i++ {
		flights[i] = createRankingTestFlight(
			"id-"+intToStr(i),
			float64(500000+i*100),
			60+i%120,
			i%4,
			6+i%12,
		)
		flights[i].RankingScore = float64(i) / float64(n)
	}

	start := time.Now()
	result := SortFlights(flights, domain.SortByBestValue)
	elapsed := time.Since(start)

	assert.Len(t, result, n)
	// Should complete in under 100ms for 10k flights
	assert.Less(t, elapsed, 100*time.Millisecond, "Sorting should be O(n log n)")
}

// intToStr is a simple helper to convert int to string without fmt.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
