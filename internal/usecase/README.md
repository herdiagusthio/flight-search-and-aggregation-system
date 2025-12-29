# Use Case Layer

This package contains the business logic for the flight search system.

## Contents

- `flight_search.go` - Flight search use case with scatter-gather pattern
- `filter.go` - Filtering logic implementation
- `ranking.go` - Ranking algorithm and sorting

## Responsibilities

- Orchestrate provider calls concurrently
- Aggregate results from multiple providers
- Apply filtering based on user criteria
- Calculate ranking scores
- Sort results based on user preference
- Handle timeouts and partial failures
