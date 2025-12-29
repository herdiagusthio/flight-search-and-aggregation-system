# Domain Layer

This package contains the core business entities and interfaces for the flight search system.

## Contents

- `flight.go` - Flight, AirlineInfo, FlightPoint, DurationInfo, PriceInfo, BaggageInfo entities
- `search.go` - SearchCriteria with validation
- `filter.go` - FilterOptions, SortOption, TimeRange
- `response.go` - SearchResponse, SearchMetadata
- `errors.go` - Domain-specific errors
- `provider.go` - FlightProvider interface

## Principles

- No external dependencies (pure Go)
- Entities are provider-agnostic
- Validation logic lives here
- Interfaces define contracts for adapters
