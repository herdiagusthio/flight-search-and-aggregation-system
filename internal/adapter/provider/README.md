# Provider Adapters

This package contains adapters for each airline provider.

## Providers

- `garuda/` - Garuda Indonesia adapter
- `lionair/` - Lion Air adapter
- `batikair/` - Batik Air adapter
- `airasia/` - AirAsia adapter

## Responsibilities

- Implement `FlightProvider` interface
- Parse provider-specific JSON formats
- Normalize data to domain `Flight` entities
- Handle provider-specific edge cases
