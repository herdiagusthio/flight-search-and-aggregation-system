# Flight Search Aggregation System

A high-performance flight search aggregation service built with Go, following Clean Architecture principles. This system aggregates flight data from multiple airline providers, applies intelligent ranking, and returns unified search results.

## Features

- 🔍 **Multi-Provider Search** - Search flights across Garuda Indonesia, Lion Air, Batik Air, and AirAsia simultaneously
- ⚡ **Concurrent Queries** - Scatter-gather pattern for parallel provider queries with configurable timeouts
- 🔄 **Graceful Degradation** - Handles partial failures gracefully, returning results from available providers
- 📊 **Intelligent Ranking** - Weighted algorithm considering price, duration, and stops for "best value" sorting
- 🎯 **Flexible Filtering** - Filter by maximum price, stops, airlines, and departure time range
- 📈 **Multiple Sort Options** - Sort by price, duration, departure time, or best value score
- 🛡️ **Production Ready** - Comprehensive error handling, structured logging, and configuration management

## Architecture

The system follows Clean Architecture with clear separation of concerns:

```
┌──────────────────────────────────────────────────────────────┐
│                   HTTP Layer (Echo v4)                       │
│              Handlers, Middleware, Request/Response          │
├──────────────────────────────────────────────────────────────┤
│                    Use Case Layer                            │
│         Flight Search, Filtering, Ranking, Sorting           │
├──────────────────────────────────────────────────────────────┤
│                    Domain Layer                              │
│      Flight, SearchCriteria, FilterOptions, Errors           │
├──────────────────────────────────────────────────────────────┤
│                  Provider Adapters                           │
│      Garuda, Lion Air, Batik Air, AirAsia Normalizers        │
└──────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Request** → HTTP Handler validates and parses the search request
2. **Scatter** → Use case dispatches concurrent queries to all providers
3. **Gather** → Results are collected with timeout handling
4. **Normalize** → Each provider adapter converts response to domain Flight entities
5. **Filter** → Apply user-specified filters (price, stops, airlines)
6. **Rank** → Calculate ranking scores for best value sorting
7. **Sort** → Apply requested sort order
8. **Response** → Return aggregated results with metadata

## Prerequisites

- **Go 1.23+** (tested with Go 1.25.4)
- **Make** (optional, for using Makefile commands)
- **golangci-lint** (optional, for linting)

## Installation

```bash
# Clone the repository
git clone https://github.com/flight-search/flight-search-and-aggregation-system.git
cd flight-search-and-aggregation-system

# Install dependencies
go mod download

# Build the application
make build
# or without Make:
go build -o bin/flight-search cmd/server/main.go
```

## Configuration

Configuration is managed via environment variables. Copy `.env.example` to `.env` for local development:

```bash
cp .env.example .env
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `SERVER_READ_TIMEOUT` | `10s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `TIMEOUT_GLOBAL_SEARCH` | `5s` | Maximum total search duration |
| `TIMEOUT_PER_PROVIDER` | `2s` | Timeout per individual provider |
| `LOG_LEVEL` | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | Log format: `json` (production), `console` (development) |
| `APP_ENV` | `development` | Environment: `development`, `staging`, `production` |

### Timeout Configuration Notes

- `TIMEOUT_PER_PROVIDER` should be less than `TIMEOUT_GLOBAL_SEARCH`
- If a provider exceeds its timeout, results from other providers are still returned
- The global timeout ensures the API always responds within a predictable time

## Running the Application

```bash
# Development mode with live reload
make run

# Or run directly
go run cmd/server/main.go

# With custom configuration
SERVER_PORT=3000 LOG_LEVEL=debug go run cmd/server/main.go

# Production binary
./bin/flight-search
```

The server starts at `http://localhost:8080` by default.

## API Documentation

### Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

The Swagger UI allows you to:
- Explore all available endpoints
- View request/response schemas
- Test API calls directly from the browser

### OpenAPI Specification

The OpenAPI 2.0 specification is available at:
- JSON: `docs/swagger.json`
- YAML: `docs/swagger.yaml`

To regenerate after code changes:
```bash
make swagger
```

### Health Check

```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

### Search Flights

```http
POST /api/v1/flights/search
Content-Type: application/json
```

#### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `origin` | string | Yes | IATA airport code (3 letters, e.g., "CGK") |
| `destination` | string | Yes | IATA airport code (3 letters, e.g., "DPS") |
| `departureDate` | string | Yes | Date in YYYY-MM-DD format |
| `passengers` | integer | Yes | Number of passengers (1-9) |
| `class` | string | No | Travel class: `economy`, `business`, `first` |
| `filters` | object | No | Optional filtering criteria |
| `sortBy` | string | No | Sort option (default: `best`) |

#### Filter Options

| Field | Type | Description |
|-------|------|-------------|
| `maxPrice` | number | Maximum price in IDR |
| `maxStops` | integer | Maximum number of stops (0 = direct only) |
| `airlines` | array | List of airline codes to include (e.g., ["GA", "JT"]) |
| `departureTimeRange` | object | Time range filter with `start` and `end` (HH:MM format) |

#### Sort Options

| Value | Description |
|-------|-------------|
| `best` | Best value score (weighted combination of price, duration, stops) |
| `price` | Lowest price first |
| `duration` | Shortest duration first |
| `departure` | Earliest departure first |

#### Example Request

```bash
# Basic search
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1
  }'
```

```bash
# Search with filters and sorting
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "maxPrice": 1000000,
      "maxStops": 0,
      "airlines": ["GA", "JT"]
    },
    "sortBy": "best"
  }'
```

```bash
# Direct flights only, sorted by price
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 2,
    "class": "economy",
    "filters": {
      "maxStops": 0
    },
    "sortBy": "price"
  }'
```

```bash
# Morning flights with departure time filter
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "departureTimeRange": {
        "start": "04:00",
        "end": "10:00"
      }
    },
    "sortBy": "departure"
  }'
```

#### Success Response (200 OK)

```json
{
  "flights": [
    {
      "id": "garuda-1",
      "flightNumber": "GA 123",
      "airline": {
        "code": "GA",
        "name": "Garuda Indonesia"
      },
      "departure": {
        "airportCode": "CGK",
        "airportName": "Soekarno-Hatta International Airport",
        "terminal": "3",
        "dateTime": "2025-12-15T08:00:00Z"
      },
      "arrival": {
        "airportCode": "DPS",
        "airportName": "Ngurah Rai International Airport",
        "terminal": "D",
        "dateTime": "2025-12-15T10:45:00Z"
      },
      "duration": {
        "totalMinutes": 165,
        "formatted": "2h 45m"
      },
      "price": {
        "amount": 1350000,
        "currency": "IDR"
      },
      "baggage": {
        "cabinKg": 7,
        "checkedKg": 20
      },
      "class": "economy",
      "stops": 0,
      "provider": "garuda_indonesia",
      "rankingScore": 0.85
    }
  ],
  "metadata": {
    "totalResults": 15,
    "searchDurationMs": 1234,
    "providersQueried": ["garuda_indonesia", "lion_air", "batik_air", "airasia"],
    "providersFailed": []
  }
}
```

#### Error Responses

**400 Bad Request** - Validation Error
```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "Request validation failed",
    "details": {
      "origin": "origin is required",
      "departureDate": "departureDate cannot be in the past"
    }
  }
}
```

**503 Service Unavailable** - All Providers Failed
```json
{
  "success": false,
  "error": {
    "code": "service_unavailable",
    "message": "All flight providers are currently unavailable"
  }
}
```

## Testing

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with coverage report
make test-cover

# Run with race detector
go test -race ./...

# Run specific package tests
go test -v ./internal/usecase/...

# Run integration tests only
go test -v ./test/integration/...
```

### Test Coverage

The project maintains high test coverage across all layers:

| Package | Coverage |
|---------|----------|
| `internal/domain` | 82%+ |
| `internal/usecase` | 89%+ |
| `internal/adapter/provider/*` | 93-95% |
| `internal/adapter/http` | 75%+ |
| `internal/config` | 97%+ |

## Project Structure

```
flight-search-and-aggregation-system/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business entities and interfaces
│   │   ├── flight.go            # Flight entity
│   │   ├── search.go            # SearchCriteria
│   │   ├── filter.go            # FilterOptions, SortOptions
│   │   ├── response.go          # SearchResponse, Metadata
│   │   ├── errors.go            # Domain errors
│   │   └── provider.go          # FlightProvider interface
│   ├── usecase/                 # Business logic
│   │   ├── flight_search.go     # Scatter-gather search
│   │   ├── filter.go            # Filtering logic
│   │   └── ranking.go           # Ranking and sorting
│   ├── adapter/
│   │   ├── http/                # HTTP layer
│   │   │   ├── handler.go       # Request handlers
│   │   │   ├── request.go       # Request DTOs and validation
│   │   │   ├── response/        # Response builders
│   │   │   ├── middleware/      # Logging, recovery, request ID
│   │   │   └── routes.go        # Route registration
│   │   └── provider/            # Airline adapters
│   │       ├── garuda/          # Garuda Indonesia
│   │       ├── lionair/         # Lion Air
│   │       ├── batikair/        # Batik Air
│   │       └── airasia/         # AirAsia
│   ├── infrastructure/          # Cross-cutting concerns
│   │   ├── logger/              # Structured logging
│   │   ├── retry/               # Retry utilities
│   │   └── timeutil/            # Time utilities
│   └── config/                  # Configuration management
├── test/
│   ├── integration/             # Integration tests
│   ├── mock/                    # Mock implementations
│   └── testutil/                # Test helpers
├── docs/
│   └── response-mock/           # Provider response mocks
├── development-docs/            # Development documentation
├── .env.example                 # Environment template
├── Makefile                     # Build automation
└── README.md                    # This file
```

## Development

### Code Formatting

```bash
# Format all Go files
make fmt

# Tidy go modules
make tidy
```

### Linting

```bash
# Run golangci-lint
make lint
```

### Building

```bash
# Build binary
make build

# Clean build artifacts
make clean
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Guidelines

- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Use meaningful commit messages
- Update documentation for new features

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Echo](https://echo.labstack.com/) web framework
- Logging powered by [zerolog](https://github.com/rs/zerolog)
- Configuration via [env](https://github.com/caarlos0/env)
