// Package http provides the HTTP handler layer for the flight search API.
package http

import (
	"strings"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
)

// ToDomainCriteria converts a SearchFlightsRequest to domain.SearchCriteria.
func ToDomainCriteria(req *SearchFlightsRequest) domain.SearchCriteria {
	class := strings.ToLower(req.Class)
	if class == "" {
		class = "economy"
	}

	passengers := req.Passengers
	if passengers < 1 {
		passengers = 1
	}

	return domain.SearchCriteria{
		Origin:        strings.ToUpper(req.Origin),
		Destination:   strings.ToUpper(req.Destination),
		DepartureDate: req.DepartureDate,
		Passengers:    passengers,
		Class:         class,
	}
}

// ToDomainFilters converts a FilterDTO to domain.FilterOptions.
func ToDomainFilters(dto *FilterDTO) *domain.FilterOptions {
	if dto == nil {
		return nil
	}

	opts := &domain.FilterOptions{
		MaxPrice: dto.MaxPrice,
		MaxStops: dto.MaxStops,
		Airlines: dto.Airlines,
	}

	// Convert time range if provided
	if dto.DepartureTimeRange != nil {
		opts.DepartureTimeRange = toDomainTimeRange(dto.DepartureTimeRange)
	}

	// Convert arrival time range if provided
	if dto.ArrivalTimeRange != nil {
		opts.ArrivalTimeRange = toDomainTimeRange(dto.ArrivalTimeRange)
	}

	// Convert duration range if provided
	if dto.DurationRange != nil {
		opts.DurationRange = toDomainDurationRange(dto.DurationRange)
	}

	return opts
}

// toDomainTimeRange converts a TimeRangeDTO to domain.TimeRange.
func toDomainTimeRange(dto *TimeRangeDTO) *domain.TimeRange {
	if dto == nil || dto.Start == "" || dto.End == "" {
		return nil
	}

	// Parse start time (HH:MM format)
	startTime, err := time.Parse("15:04", dto.Start)
	if err != nil {
		return nil
	}

	// Parse end time (HH:MM format)
	endTime, err := time.Parse("15:04", dto.End)
	if err != nil {
		return nil
	}

	return &domain.TimeRange{
		Start: startTime,
		End:   endTime,
	}
}

// toDomainDurationRange converts a DurationRangeDTO to domain.DurationRange.
func toDomainDurationRange(dto *DurationRangeDTO) *domain.DurationRange {
	if dto == nil {
		return nil
	}

	// Return nil if both fields are nil (no filter)
	if dto.MinMinutes == nil && dto.MaxMinutes == nil {
		return nil
	}

	return &domain.DurationRange{
		MinMinutes: dto.MinMinutes,
		MaxMinutes: dto.MaxMinutes,
	}
}

// ToDomainSortOption converts a sort string to domain.SortOption.
func ToDomainSortOption(sortBy string) domain.SortOption {
	switch strings.ToLower(sortBy) {
	case "best", "best_value":
		return domain.SortByBestValue
	case "price":
		return domain.SortByPrice
	case "duration":
		return domain.SortByDuration
	case "departure":
		return domain.SortByDeparture
	default:
		return domain.SortByBestValue // Default to best value
	}
}

// ToSearchOptions converts request fields to usecase.SearchOptions.
func ToSearchOptions(req *SearchFlightsRequest) usecase.SearchOptions {
	return usecase.SearchOptions{
		Filters: ToDomainFilters(req.Filters),
		SortBy:  ToDomainSortOption(req.SortBy),
	}
}
