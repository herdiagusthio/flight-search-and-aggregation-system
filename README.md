# Flight Search Aggregation System

A high-performance flight search aggregation service built with Go, following Clean Architecture principles. This system aggregates flight data from multiple airline providers, applies intelligent ranking, and returns unified search results.

## Features

- 🔍 **Multi-Provider Search** - Aggregates flights from Garuda Indonesia, Lion Air, Batik Air, and AirAsia
- ⚡ **Concurrent Queries** - Scatter-gather pattern with parallel provider queries and configurable timeouts
- 🔄 **Graceful Degradation** - Returns partial results when providers fail or timeout
- 📊 **Intelligent Ranking** - Weighted scoring algorithm combining price, duration, and stops
- 🎯 **Flexible Filtering** - Filter by price, stops, airlines, and departure time range
- 📈 **Multiple Sort Options** - Sort by best value, price, duration, or departure time
- 🔧 **Swagger/OpenAPI** - Interactive API documentation and testing interface
- 🛡️ **Production Ready** - Comprehensive error handling, structured logging, and environment-based configuration

## Architecture

The system follows Clean Architecture with clear separation of concerns:

```
┌──────────────────────────────────────────────────────────────┐
│                   HTTP Layer (Echo v4)                       │
│        Handlers, DTOs, Middleware, Request/Response          │
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

### Key Design Decisions

- **DTO Layer**: Separates domain models from API responses with snake_case JSON formatting
- **Provider Abstraction**: Each airline implements the `FlightProvider` interface for consistent integration
- **Timeout Management**: Two-tier timeout system (per-provider and global search limits)
- **Error Isolation**: Provider failures don't cascade; successful responses are still returned

### Data Flow

1. **Request** → HTTP handler validates and parses the search request
2. **Scatter** → Use case dispatches concurrent queries to all providers
3. **Gather** → Results are collected with timeout handling
4. **Normalize** → Provider adapters convert responses to domain `Flight` entities
5. **Filter** → Apply user-specified filters (price, stops, airlines, time range)
6. **Rank** → Calculate ranking scores for best value sorting
7. **Sort** → Apply requested sort order
8. **Transform** → DTO layer converts to API response format
9. **Response** → Return aggregated results with metadata

## Prerequisites

- **Go 1.23+** (module uses Go 1.25.4)
- **Make** (optional, for using Makefile commands)
- **swag** (optional, for regenerating Swagger docs): `go install github.com/swaggo/swag/cmd/swag@latest`
- **golangci-lint** (optional, for linting): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

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

# Generate Swagger documentation (optional)
make swagger
# or:
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
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
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 15,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 285,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "QZ7250_AirAsia",
      "provider": "AirAsia",
      "airline": {
        "name": "AirAsia",
        "code": "QZ"
      },
      "flight_number": "QZ7250",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T15:15:00+07:00",
        "timestamp": 1734246900
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T20:35:00+08:00",
        "timestamp": 1734267300
      },
      "duration": {
        "total_minutes": 260,
        "formatted": "4h 20m"
      },
      "stops": 1,
      "price": {
        "amount": 485000,
        "currency": "IDR"
      },
      "available_seats": 88,
      "cabin_class": "economy",
      "aircraft": null,
      "amenities": [],
      "baggage": {
        "carry_on": "Cabin baggage only",
        "checked": "Additional fee"
      }
    }
  ]
}
```

**Response Field Descriptions:**

- `search_criteria`: Original search parameters
- `metadata.total_results`: Number of flights returned after filtering
- `metadata.providers_queried`: Total number of providers contacted
- `metadata.providers_succeeded`: Providers that returned results successfully
- `metadata.providers_failed`: Providers that failed or timed out
- `metadata.search_time_ms`: Total search execution time in milliseconds
- `metadata.cache_hit`: Whether results came from cache (currently always `false`)
- `flights[].timestamp`: Unix timestamp (seconds since epoch)
- `flights[].baggage`: Formatted baggage information (e.g., "7 kg" → "Cabin baggage only")

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

The project maintains comprehensive test coverage across all layers.

```bash
# Run all tests
make test
# or:
go test ./...

# Run tests in short mode (skip integration tests)
make test-short
# or:
go test -short ./...

# Run with verbose output
go test -v ./...

# Run with coverage report
make test-cover
# or:
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html

# Run with race detector
make test-race
# or:
go test -race ./...

# Run specific package tests
go test -v ./internal/usecase/...

# Run integration tests only
make test-integration
# or:
go test -v ./test/integration/...

# Run benchmarks
make bench
# or:
go test -bench=. -benchmem ./...
```

### Test Coverage

The project maintains high test coverage:

| Package | Coverage |
|---------|----------|
| `internal/domain` | 82%+ |
| `internal/usecase` | 89%+ |
| `internal/adapter/provider/*` | 93-95% |
| `internal/adapter/http` | 75%+ |
| `internal/config` | 97%+ |

### Test Organization

- **Unit Tests**: Co-located with source files (`*_test.go`)
- **Integration Tests**: Located in `test/integration/`
- **Mock Implementations**: Located in `test/mock/` and generated via `//go:generate mockgen`
- **Test Utilities**: Shared helpers in `test/testutil/`

## Project Structure

```
flight-search-and-aggregation-system/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point and Swagger annotations
├── internal/
│   ├── domain/                  # Business entities and interfaces
│   │   ├── flight.go            # Flight entity
│   │   ├── search.go            # SearchCriteria and validation
│   │   ├── filter.go            # FilterOptions, SortOptions, TimeRange
│   │   ├── response.go          # SearchResponse, SearchMetadata
│   │   ├── errors.go            # Domain-specific errors
│   │   └── provider.go          # FlightProvider interface
│   ├── usecase/                 # Business logic
│   │   ├── flight_search.go     # Scatter-gather search orchestration
│   │   ├── filter.go            # Flight filtering logic
│   │   ├── ranking.go           # Ranking and sorting algorithms
│   │   └── options.go           # Use case configuration options
│   ├── adapter/
│   │   ├── http/                # HTTP layer
│   │   │   ├── handler.go       # Request handlers
│   │   │   ├── request.go       # Request validation
│   │   │   ├── response.go      # Response builders
│   │   │   ├── dto.go           # DTO transformation layer
│   │   │   ├── converter.go     # Domain to DTO converters
│   │   │   ├── routes.go        # Route registration
│   │   │   ├── swagger_types.go # Swagger documentation types
│   │   │   ├── middleware/      # Request logging, recovery, etc.
│   │   │   └── response/        # Response formatting utilities
│   │   └── provider/            # Airline provider adapters
│   │       ├── garuda/          # Garuda Indonesia adapter
│   │       ├── lionair/         # Lion Air adapter
│   │       ├── batikair/        # Batik Air adapter
│   │       └── airasia/         # AirAsia adapter
│   ├── infrastructure/          # Cross-cutting concerns
│   │   ├── logger/              # Structured logging (zerolog)
│   │   ├── retry/               # Retry utilities
│   │   └── timeutil/            # Time utilities and timezone handling
│   └── config/                  # Configuration management
│       └── config.go            # Environment variable loading
├── test/
│   ├── integration/             # Integration tests
│   │   ├── handler_test.go      # HTTP handler integration tests
│   │   ├── usecase_test.go      # Use case integration tests
│   │   ├── concurrent_test.go   # Concurrency tests
│   │   └── setup.go             # Test setup and utilities
│   ├── mock/                    # Mock implementations
│   │   └── provider.go          # Mock flight provider
│   └── testutil/                # Test helper functions
│       └── helpers.go           # Shared test utilities
├── docs/
│   ├── swagger.json             # Generated OpenAPI 2.0 spec (JSON)
│   ├── swagger.yaml             # Generated OpenAPI 2.0 spec (YAML)
│   ├── docs.go                  # Generated Swagger documentation
│   └── response-mock/           # Sample provider responses
│       ├── garuda_indonesia_search_response.json
│       ├── lion_air_search_response.json
│       ├── batik_air_search_response.json
│       └── airasia_search_response.json
├── development-docs/            # Development documentation
│   └── requirements/            # Sample request/response files
│       └── expected_result.json # Expected API output format
├── bin/                         # Compiled binaries (gitignored)
├── .env.example                 # Environment variable template
├── .env                         # Local environment (gitignored)
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # This file
```

### Key Directories

- **`cmd/`**: Application entry points
- **`internal/domain/`**: Core business logic and entities (framework-agnostic)
- **`internal/usecase/`**: Application-specific business rules
- **`internal/adapter/`**: External integrations (HTTP, providers)
- **`internal/infrastructure/`**: Technical capabilities (logging, retry, time utilities)
- **`test/`**: All test code (integration tests, mocks, utilities)

## Development

### Available Make Commands

```bash
make help              # Show all available commands

# Build
make build             # Build production binary
make build-debug       # Build with debug symbols
make clean             # Remove build artifacts

# Run
make run               # Run the application
make run-dev           # Run with development settings (debug logs)

# Test
make test              # Run all tests
make test-short        # Run tests in short mode
make test-cover        # Run tests with coverage report
make test-race         # Run tests with race detector
make test-integration  # Run integration tests only
make bench             # Run benchmarks

# Code Quality
make fmt               # Format code
make vet               # Run go vet
make lint              # Run golangci-lint
make check             # Run fmt, vet, and lint

# Dependencies
make deps              # Download dependencies
make tidy              # Tidy go modules
make verify            # Verify dependencies

# Documentation
make swagger           # Generate Swagger documentation

# Code Generation
make generate          # Run go generate
make mocks             # Generate mocks using mockgen
```

### Code Formatting and Linting

```bash
# Format all Go files
make fmt

# Run linter (requires golangci-lint)
make lint

# Run all checks
make check
```

### Regenerating Swagger Documentation

After modifying API handlers or Swagger annotations:

```bash
make swagger
```

This regenerates `docs/swagger.json`, `docs/swagger.yaml`, and `docs/docs.go`.

### Generating Mocks

Mock interfaces are generated using `mockgen`:

```bash
# Generate all mocks
make mocks

# Or use go generate
go generate ./...
```

### Development Guidelines

- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Use meaningful commit messages (conventional commits format preferred)
- Update documentation for new features
- Ensure code is properly formatted (`make fmt`)
- Fix all linter warnings (`make lint`)
- Add Swagger annotations for new API endpoints

## Known Limitations

- **No Caching**: Search results are not cached (metadata always shows `cache_hit: false`)
- **Mock Data**: Provider adapters currently use static mock JSON responses
- **Date Validation**: Past dates are accepted (validation removed to support testing with historical mock data)
- **In-Memory Only**: No persistent storage or database integration
- **Single Region**: Mock data uses Indonesian airports and airlines only

## Assumptions

- All providers return data in a known JSON format
- Flight times include timezone information (`+07:00`, `+08:00`)
- Prices are in Indonesian Rupiah (IDR)
- Airport codes follow IATA 3-letter format
- Maximum 9 passengers per search
- Departure time range filter compares time-of-day only (ignores date and timezone)

## Tech Stack

### Core Dependencies

- **[Echo v4](https://echo.labstack.com/)** - High-performance HTTP web framework
- **[zerolog](https://github.com/rs/zerolog)** - Structured, leveled logging
- **[env](https://github.com/caarlos0/env)** - Environment variable parsing
- **[godotenv](https://github.com/joho/godotenv)** - `.env` file loading
- **[uuid](https://github.com/google/uuid)** - UUID generation

### API Documentation

- **[swag](https://github.com/swaggo/swag)** - Swagger/OpenAPI code generator
- **[echo-swagger](https://github.com/swaggo/echo-swagger)** - Swagger UI middleware for Echo

### Testing

- **[testify](https://github.com/stretchr/testify)** - Testing toolkit with assertions
- **[gomock](https://github.com/uber-go/mock)** - Mock generation framework

## Acknowledgments

- Built with [Echo](https://echo.labstack.com/) web framework
- Logging powered by [zerolog](https://github.com/rs/zerolog)
- Configuration management via [env](https://github.com/caarlos0/env)
- API documentation with [swaggo](https://github.com/swaggo/swag)
