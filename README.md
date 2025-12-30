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
- **Data Validation**: Flight data is validated during normalization with graceful error handling

### Data Quality & Validation

The system implements automatic validation of flight data to ensure data integrity:

#### Validation Rules

Each flight is validated after normalization from provider data:

- **Time Consistency**: Arrival time must be chronologically after departure time
- **Required Fields**: All flights must have:
  - Flight number
  - Airline code
  - Departure airport code
  - Arrival airport code

#### Graceful Degradation

Invalid flights are handled gracefully:

- Validation errors are logged with flight details (provider, flight number, error reason)
- Invalid flights are skipped and excluded from search results
- Other valid flights from the same provider are still returned
- Search continues even if some providers return invalid data
- No errors are returned to the user unless all providers fail

#### Duration Mismatch Handling

Provider-reported durations may differ from calculated time differences due to:
- Different calculation methods (elapsed time vs. flight time)
- Timezone considerations
- Data source inconsistencies

**Approach**: Duration mismatches log a warning but don't fail validation, as the calculated time difference is used for ranking and filtering.

### Data Flow

1. **Request** → HTTP handler validates and parses the search request
2. **Scatter** → Use case dispatches concurrent queries to all providers
3. **Gather** → Results are collected with timeout handling
4. **Normalize** → Provider adapters convert responses to domain `Flight` entities
5. **Validate** → Flight data is validated for time consistency and required fields
6. **Filter** → Apply user-specified filters (price, stops, airlines, time range)
7. **Rank** → Calculate ranking scores for best value sorting
8. **Sort** → Apply requested sort order
9. **Transform** → DTO layer converts to API response format
10. **Response** → Return aggregated results with metadata

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
| `arrivalTimeRange` | object | Time range filter with `start` and `end` (HH:MM format) |
| `durationRange` | object | Duration range filter with `minMinutes` and/or `maxMinutes` |

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

```bash
# Short flights with duration range filter
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "durationRange": {
        "minMinutes": 60,
        "maxMinutes": 180
      }
    },
    "sortBy": "duration"
  }'
```

```bash
# Combined filters: budget flights under 3 hours, morning departure
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
      "durationRange": {
        "maxMinutes": 180
      },
      "departureTimeRange": {
        "start": "06:00",
        "end": "12:00"
      }
    },
    "sortBy": "price"
  }'
```

```bash
# Business hours arrival - flights arriving between 8 AM and 5 PM
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "arrivalTimeRange": {
        "start": "08:00",
        "end": "17:00"
      }
    },
    "sortBy": "departure"
  }'
```

```bash
# Morning departure with business hours arrival
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "departureTimeRange": {
        "start": "06:00",
        "end": "12:00"
      },
      "arrivalTimeRange": {
        "start": "08:00",
        "end": "17:00"
      }
    },
    "sortBy": "price"
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

## Filtering

The Flight Search API provides powerful filtering capabilities to help users find flights that match specific criteria. Filters can be combined to create complex queries.

### Available Filters

| Filter | Type | Description | Example |
|--------|------|-------------|---------|
| `maxPrice` | number | Maximum price in IDR | `1000000` |
| `maxStops` | integer | Maximum number of stops (0 = direct) | `0` |
| `airlines` | array | Airline codes to include | `["GA", "JT"]` |
| `departureTimeRange` | object | Departure time window | `{"start": "06:00", "end": "12:00"}` |
| `arrivalTimeRange` | object | Arrival time window | `{"start": "08:00", "end": "17:00"}` |
| `durationRange` | object | Flight duration limits | `{"minMinutes": 60, "maxMinutes": 180}` |

### Filter Behavior

- **All filters are optional** - Omit filters to get unfiltered results
- **Filters are combined with AND logic** - Flights must match all specified filters
- **Empty filter object** - Treated the same as no filters
- **Invalid filter values** - Return 400 Bad Request with validation details

### Duration Range Filter

Filter flights by total flight duration in minutes.

**Fields:**
- `minMinutes` (integer, optional): Minimum duration in minutes
- `maxMinutes` (integer, optional): Maximum duration in minutes

**Validation:**
- Both fields are optional; at least one must be specified
- Values must be positive integers
- `minMinutes` must be ≤ `maxMinutes` if both are specified

**Examples:**

```bash
# Flights under 2 hours (120 minutes)
{
  "filters": {
    "durationRange": {
      "maxMinutes": 120
    }
  }
}
```

```bash
# Flights between 1-3 hours
{
  "filters": {
    "durationRange": {
      "minMinutes": 60,
      "maxMinutes": 180
    }
  }
}
```

```bash
# Only long-haul flights (over 4 hours)
{
  "filters": {
    "durationRange": {
      "minMinutes": 240
    }
  }
}
```

### Arrival Time Range Filter

Filter flights by arrival time of day (time-only, ignores date and timezone).

**Fields:**
- `start` (string, required): Start time in HH:MM format (24-hour)
- `end` (string, required): End time in HH:MM format (24-hour)

**Validation:**
- Both `start` and `end` are required
- Must be in HH:MM format (e.g., "08:00", "17:30")
- Hours: 00-23, Minutes: 00-59
- `start` must be before `end` (no overnight ranges)

**Examples:**

```bash
# Business hours arrival (8 AM - 5 PM)
{
  "filters": {
    "arrivalTimeRange": {
      "start": "08:00",
      "end": "17:00"
    }
  }
}
```

```bash
# Morning arrival only
{
  "filters": {
    "arrivalTimeRange": {
      "start": "06:00",
      "end": "12:00"
    }
  }
}
```

```bash
# Evening arrival for hotel check-in
{
  "filters": {
    "arrivalTimeRange": {
      "start": "18:00",
      "end": "23:00"
    }
  }
}
```

### Departure Time Range Filter

Filter flights by departure time of day (time-only, ignores date and timezone).

**Fields:**
- `start` (string, required): Start time in HH:MM format (24-hour)
- `end` (string, required): End time in HH:MM format (24-hour)

**Validation:**
- Both `start` and `end` are required
- Must be in HH:MM format
- Hours: 00-23, Minutes: 00-59
- `start` must be before `end`

**Example:**

```bash
# Early morning departure (red-eye flights)
{
  "filters": {
    "departureTimeRange": {
      "start": "00:00",
      "end": "06:00"
    }
  }
}
```

### Combined Filter Examples

Filters can be combined to create highly specific queries.

**Example 1: Budget direct flights, morning departure**

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 1,
  "filters": {
    "maxPrice": 800000,
    "maxStops": 0,
    "departureTimeRange": {
      "start": "06:00",
      "end": "12:00"
    }
  },
  "sortBy": "price"
}
```

**Example 2: Quick flights with specific airlines**

```json
{
  "origin": "CGK",
  "destination": "SUB",
  "departureDate": "2025-12-20",
  "passengers": 2,
  "filters": {
    "airlines": ["GA", "ID"],
    "durationRange": {
      "maxMinutes": 90
    }
  },
  "sortBy": "duration"
}
```

**Example 3: Business travel optimization (morning departure, business hours arrival)**

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 1,
  "class": "business",
  "filters": {
    "departureTimeRange": {
      "start": "05:00",
      "end": "09:00"
    },
    "arrivalTimeRange": {
      "start": "08:00",
      "end": "12:00"
    },
    "maxStops": 0
  },
  "sortBy": "best"
}
```

**Example 4: All filters combined**

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 1,
  "filters": {
    "maxPrice": 1200000,
    "maxStops": 1,
    "airlines": ["GA", "JT", "ID"],
    "departureTimeRange": {
      "start": "06:00",
      "end": "18:00"
    },
    "arrivalTimeRange": {
      "start": "08:00",
      "end": "20:00"
    },
    "durationRange": {
      "minMinutes": 60,
      "maxMinutes": 240
    }
  },
  "sortBy": "best"
}
```

### Filter Validation Errors

When filter validation fails, the API returns a 400 Bad Request with detailed error information:

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "Request validation failed",
    "details": {
      "filters.durationRange": "minMinutes must be less than or equal to maxMinutes",
      "filters.arrivalTimeRange.start": "start time must be in HH:MM format"
    }
  }
}
```

**Common validation errors:**

| Error | Cause | Solution |
|-------|-------|----------|
| `minMinutes must be less than or equal to maxMinutes` | Duration range invalid | Ensure min ≤ max |
| `start time must be in HH:MM format` | Invalid time format | Use 24-hour format: "08:00" |
| `start time must be before end time` | Time range invalid | Ensure start < end |
| `invalid time value` | Hour/minute out of range | Hours: 00-23, Minutes: 00-59 |
| `maxPrice must be positive` | Negative price | Use positive numbers only |
| `maxStops must be non-negative` | Negative stops | Use 0 or positive integers |

## Data Validation

The system implements automatic data quality validation to ensure flight data integrity.

### Flight Validation Rules

Every flight is validated after normalization from provider data:

1. **Time Consistency**
   - Arrival time must be chronologically after departure time
   - Violations are logged and the flight is excluded from results
   - Error format: `invalid flight times: arrival time (TIME) must be after departure time (TIME)`

2. **Required Fields**
   - `FlightNumber` - Cannot be empty
   - `Airline.Code` - Cannot be empty (IATA code)
   - `Departure.AirportCode` - Cannot be empty (IATA code)
   - `Arrival.AirportCode` - Cannot be empty (IATA code)
   - Error format: `missing required field: FIELD_NAME`

3. **Duration Mismatch (Warning Only)**
   - Provider-reported duration may differ from calculated time difference
   - Logged as warning but doesn't fail validation
   - System uses calculated duration for ranking and filtering

### Graceful Degradation

The system handles invalid data gracefully:

- **Invalid flights are automatically excluded** from search results
- **Other valid flights** from the same provider are still returned
- **Validation errors are logged** with provider and flight details for debugging
- **Search continues** even if some providers return invalid data
- **No user-facing errors** unless all providers fail completely

### Validation in Action

**Scenario:** Provider returns 10 flights, 2 have invalid times

```
Provider: garuda_indonesia
Total flights from provider: 10
Invalid flights detected: 2
  - GA-123: arrival time (10:00) before departure time (15:00)
  - GA-456: missing required field: Airline.Code
Valid flights returned: 8
```

**Result:** User receives 8 valid flights, invalid flights are logged for monitoring

## Troubleshooting

### Common Issues and Solutions

#### No Results Returned

**Symptom:** Search returns `"total_results": 0` with empty flights array

**Possible Causes:**
1. **Filters too restrictive** - No flights match all criteria
2. **Invalid date** - No flights available for that date in mock data
3. **All providers failed** - Check `metadata.providers_failed` count

**Solutions:**
```bash
# Check metadata for clues
{
  "metadata": {
    "total_results": 0,
    "providers_succeeded": 4,  # If >0, filters may be too strict
    "providers_failed": 0
  }
}

# Try removing filters one by one
# Start with a basic search (no filters)
# Add filters incrementally to identify which filter eliminates all results
```

#### Validation Errors

**Symptom:** 400 Bad Request with validation details

**Common errors:**

1. **"origin is required"**
   ```json
   // Missing required field
   {"destination": "DPS", ...}  // ❌ Missing origin
   
   // Solution: Add all required fields
   {"origin": "CGK", "destination": "DPS", ...}  // ✓
   ```

2. **"invalid IATA airport code format"**
   ```json
   {"origin": "Jakarta", ...}  // ❌ Must be 3-letter code
   {"origin": "cgk", ...}       // ❌ Must be uppercase
   
   {"origin": "CGK", ...}       // ✓ Correct
   ```

3. **"departureDate must be in YYYY-MM-DD format"**
   ```json
   {"departureDate": "12/15/2025"}  // ❌ Wrong format
   {"departureDate": "2025-12-15"}  // ✓ Correct
   ```

4. **"minMinutes must be less than or equal to maxMinutes"**
   ```json
   {
     "durationRange": {
       "minMinutes": 180,
       "maxMinutes": 60  // ❌ Min > Max
     }
   }
   
   // Solution: Fix the range
   {
     "durationRange": {
       "minMinutes": 60,
       "maxMinutes": 180  // ✓ Min ≤ Max
     }
   }
   ```

5. **"start time must be before end time"**
   ```json
   {
     "departureTimeRange": {
       "start": "18:00",
       "end": "06:00"  // ❌ Overnight ranges not supported
     }
   }
   
   // Solution: Use same-day range
   {
     "departureTimeRange": {
       "start": "06:00",
       "end": "18:00"  // ✓
     }
   }
   ```

#### All Providers Failed (503 Error)

**Symptom:** 503 Service Unavailable response

```json
{
  "success": false,
  "error": {
    "code": "service_unavailable",
    "message": "All flight providers are currently unavailable"
  }
}
```

**Possible Causes:**
1. Mock data files not found or corrupted
2. All providers timing out (check `TIMEOUT_PER_PROVIDER` setting)
3. File permission issues

**Solutions:**
1. Verify mock data files exist in expected locations
2. Check server logs for specific provider errors
3. Increase timeout values in `.env` if providers are slow
4. Ensure file paths are correct in provider configuration

#### Slow Response Times

**Symptom:** Search takes >2 seconds to complete

**Diagnostics:**
```json
{
  "metadata": {
    "search_time_ms": 2500,  // > 2000ms is slow
    "providers_succeeded": 3,
    "providers_failed": 1     // Check which provider failed
  }
}
```

**Solutions:**
1. Check `metadata.search_time_ms` to identify if it's a timeout issue
2. Adjust `TIMEOUT_PER_PROVIDER` and `TIMEOUT_GLOBAL_SEARCH` in `.env`
3. One slow/failing provider shouldn't delay others (scatter-gather pattern)
4. Check server logs for timeout details

#### Unexpected Filter Behavior

**Issue:** Filters not working as expected

**Debugging Steps:**

1. **Test filters individually**
   ```bash
   # Test each filter separately to isolate the issue
   # Start with no filters → add one filter → add another
   ```

2. **Check filter data types**
   ```json
   // Wrong: maxPrice as string
   {"maxPrice": "1000000"}  // ❌
   
   // Correct: maxPrice as number
   {"maxPrice": 1000000}    // ✓
   ```

3. **Verify time format**
   ```json
   // Wrong: 12-hour format
   {"start": "6:00 AM"}     // ❌
   
   // Correct: 24-hour HH:MM
   {"start": "06:00"}       // ✓
   ```

4. **Check array format**
   ```json
   // Wrong: airlines as string
   {"airlines": "GA,JT"}    // ❌
   
   // Correct: airlines as array
   {"airlines": ["GA", "JT"]}  // ✓
   ```

### Logging and Debugging

**Enable debug logging:**
```bash
# Set in .env file
LOG_LEVEL=debug
LOG_FORMAT=console  # More readable than JSON for debugging
```

**Check logs for:**
- Provider timeout errors
- Validation failures with flight details
- Filter application details
- Response times per provider

**Example debug log output:**
```
[DEBUG] Applying filters to 45 flights
[WARN] [garuda_indonesia] Flight GA-123 validation failed: arrival time must be after departure time
[INFO] Filters applied: 42 flights remaining
[DEBUG] Ranking 42 flights by best value
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
