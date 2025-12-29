// Package usecase provides the business logic for flight search operations.
package usecase

import (
	"math"
	"sort"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// Ranking algorithm weights.
// These weights determine the importance of each factor in the ranking score.
// The sum of weights equals 1.0 for normalized scoring.
const (
	// weightPrice is the weight for price in ranking calculation.
	// Price has the highest impact on ranking (50%).
	weightPrice = 0.5

	// weightDuration is the weight for flight duration in ranking calculation.
	// Duration has moderate impact on ranking (30%).
	weightDuration = 0.3

	// weightStops is the weight for number of stops in ranking calculation.
	// Stops has the lowest impact on ranking (20%).
	weightStops = 0.2
)

// CalculateRankingScores calculates the ranking score for each flight using a weighted formula.
//
// The ranking algorithm uses normalization to ensure fair comparison across different value ranges:
//
//	Score = (0.5 × NormalizedPrice) + (0.3 × NormalizedDuration) + (0.2 × NormalizedStops)
//
// Where normalized values are in the range [0, 1]:
//   - 0 = best (lowest price, shortest duration, fewest stops)
//   - 1 = worst (highest price, longest duration, most stops)
//
// Lower score = better value flight.
//
// Behavior:
//   - Returns empty slice for empty input
//   - Single flight gets score of 0 (all normalized values are 0 when min == max)
//   - All equal values result in equal scores of 0
//   - Does NOT mutate the original flights slice
//   - Performance is O(n) where n = number of flights
func CalculateRankingScores(flights []domain.Flight) []domain.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Find min/max for normalization
	minPrice, maxPrice := findPriceRange(flights)
	minDuration, maxDuration := findDurationRange(flights)
	minStops, maxStops := findStopsRange(flights)

	// Calculate scores - create a copy to avoid mutating input
	result := make([]domain.Flight, len(flights))
	for i, f := range flights {
		result[i] = f

		normPrice := normalizeValue(f.Price.Amount, minPrice, maxPrice)
		normDuration := normalizeValue(float64(f.Duration.TotalMinutes), float64(minDuration), float64(maxDuration))
		normStops := normalizeValue(float64(f.Stops), float64(minStops), float64(maxStops))

		result[i].RankingScore = (weightPrice * normPrice) +
			(weightDuration * normDuration) +
			(weightStops * normStops)
	}

	return result
}

// normalizeValue normalizes a value to the range [0, 1] based on min and max.
// Returns 0 when min == max (all values equal = all optimal).
// This avoids division by zero and treats uniform values as equally good.
func normalizeValue(value, min, max float64) float64 {
	if max == min {
		return 0 // All values equal = all optimal
	}
	return (value - min) / (max - min)
}

// findPriceRange finds the minimum and maximum price across all flights.
func findPriceRange(flights []domain.Flight) (min, max float64) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxFloat64
	max = 0

	for _, f := range flights {
		if f.Price.Amount < min {
			min = f.Price.Amount
		}
		if f.Price.Amount > max {
			max = f.Price.Amount
		}
	}
	return min, max
}

// findDurationRange finds the minimum and maximum duration (in minutes) across all flights.
func findDurationRange(flights []domain.Flight) (min, max int) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxInt
	max = 0

	for _, f := range flights {
		if f.Duration.TotalMinutes < min {
			min = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > max {
			max = f.Duration.TotalMinutes
		}
	}
	return min, max
}

// findStopsRange finds the minimum and maximum number of stops across all flights.
func findStopsRange(flights []domain.Flight) (min, max int) {
	if len(flights) == 0 {
		return 0, 0
	}

	min = math.MaxInt
	max = 0

	for _, f := range flights {
		if f.Stops < min {
			min = f.Stops
		}
		if f.Stops > max {
			max = f.Stops
		}
	}
	return min, max
}

// SortFlights sorts flights according to the specified sort option.
// Uses stable sorting to maintain consistent order for equal values.
//
// Sort options:
//   - SortByBestValue (default): ascending by RankingScore (lower = better)
//   - SortByPrice: ascending by Price.Amount (cheapest first)
//   - SortByDuration: ascending by Duration.TotalMinutes (shortest first)
//   - SortByDeparture: ascending by Departure.DateTime (earliest first)
//
// Behavior:
//   - Returns empty slice for empty input
//   - Single flight returns as-is
//   - Empty or invalid sortBy defaults to SortByBestValue
//   - Does NOT mutate the original flights slice
func SortFlights(flights []domain.Flight, sortBy domain.SortOption) []domain.Flight {
	if len(flights) == 0 {
		return flights
	}

	// Copy to avoid mutating input
	result := make([]domain.Flight, len(flights))
	copy(result, flights)

	// Single flight doesn't need sorting
	if len(result) == 1 {
		return result
	}

	// Default to best value if sortBy is empty or invalid
	if sortBy == "" || !sortBy.IsValid() {
		sortBy = domain.SortByBestValue
	}

	switch sortBy {
	case domain.SortByBestValue:
		// Lower score = better value
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].RankingScore < result[j].RankingScore
		})
	case domain.SortByPrice:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Price.Amount < result[j].Price.Amount
		})
	case domain.SortByDuration:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Duration.TotalMinutes < result[j].Duration.TotalMinutes
		})
	case domain.SortByDeparture:
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Departure.DateTime.Before(result[j].Departure.DateTime)
		})
	}

	return result
}
