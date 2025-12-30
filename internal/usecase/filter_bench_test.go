package usecase

import (
	"testing"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
)

// BenchmarkApplyFilters benchmarks the filter application with various filter combinations
func BenchmarkApplyFilters(b *testing.B) {
	// Create test flights
	flights := make([]domain.Flight, 100)
	baseTime := time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)
	
	for i := 0; i < 100; i++ {
		departureTime := baseTime.Add(time.Duration(i*30) * time.Minute)
		arrivalTime := departureTime.Add(2 * time.Hour)
		
		flights[i] = domain.Flight{
			ID:           "test-flight-" + string(rune(i)),
			FlightNumber: "GA-" + string(rune(100+i)),
			Airline: domain.AirlineInfo{
				Code: "GA",
				Name: "Garuda Indonesia",
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				DateTime:    departureTime,
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				DateTime:    arrivalTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: 120,
				Formatted:    "2h",
			},
			Price: domain.PriceInfo{
				Amount:   float64(500000 + i*10000),
				Currency: "IDR",
			},
			Stops:    i % 2, // Alternating direct and 1-stop flights
			Provider: "test",
		}
	}

	b.Run("no_filters", func(b *testing.B) {
		filters := &domain.FilterOptions{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ApplyFilters(flights, filters)
		}
	})

	b.Run("price_filter", func(b *testing.B) {
		maxPrice := 1000000.0
		filters := &domain.FilterOptions{
			MaxPrice: &maxPrice,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ApplyFilters(flights, filters)
		}
	})

	b.Run("duration_filter", func(b *testing.B) {
		minDuration := 60
		maxDuration := 180
		filters := &domain.FilterOptions{
			DurationRange: &domain.DurationRange{
				MinMinutes: &minDuration,
				MaxMinutes: &maxDuration,
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ApplyFilters(flights, filters)
		}
	})

	b.Run("arrival_time_filter", func(b *testing.B) {
		arrivalStart := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
		arrivalEnd := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)
		filters := &domain.FilterOptions{
			ArrivalTimeRange: &domain.TimeRange{
				Start: arrivalStart,
				End:   arrivalEnd,
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ApplyFilters(flights, filters)
		}
	})

	b.Run("all_filters_combined", func(b *testing.B) {
		maxPrice := 1000000.0
		maxStops := 1
		airlines := []string{"GA", "JT"}
		minDuration := 60
		maxDuration := 180
		departureStart := time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)
		departureEnd := time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)
		arrivalStart := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
		arrivalEnd := time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)
		
		filters := &domain.FilterOptions{
			MaxPrice: &maxPrice,
			MaxStops: &maxStops,
			Airlines: airlines,
			DurationRange: &domain.DurationRange{
				MinMinutes: &minDuration,
				MaxMinutes: &maxDuration,
			},
			DepartureTimeRange: &domain.TimeRange{
				Start: departureStart,
				End:   departureEnd,
			},
			ArrivalTimeRange: &domain.TimeRange{
				Start: arrivalStart,
				End:   arrivalEnd,
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ApplyFilters(flights, filters)
		}
	})
}

// BenchmarkRankFlights benchmarks the ranking operation
func BenchmarkRankFlights(b *testing.B) {
	// Create test flights
	flights := make([]domain.Flight, 100)
	baseTime := time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC)
	
	for i := 0; i < 100; i++ {
		departureTime := baseTime.Add(time.Duration(i*30) * time.Minute)
		arrivalTime := departureTime.Add(2 * time.Hour)
		
		flights[i] = domain.Flight{
			ID:           "test-flight-" + string(rune(i)),
			FlightNumber: "GA-" + string(rune(100+i)),
			Airline: domain.AirlineInfo{
				Code: "GA",
				Name: "Garuda Indonesia",
			},
			Departure: domain.FlightPoint{
				AirportCode: "CGK",
				DateTime:    departureTime,
			},
			Arrival: domain.FlightPoint{
				AirportCode: "DPS",
				DateTime:    arrivalTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: 120,
				Formatted:    "2h",
			},
			Price: domain.PriceInfo{
				Amount:   float64(500000 + i*10000),
				Currency: "IDR",
			},
			Stops:    i % 2,
			Provider: "test",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateRankingScores(flights)
	}
}
