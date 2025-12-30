# Flight Search API Documentation

This document provides detailed API documentation for the Flight Search Aggregation System.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not require authentication. Future versions may implement API key or OAuth2 authentication.

---

## Endpoints

### Health Check

Check if the service is running and healthy.

```http
GET /health
```

#### Response

**200 OK**
```json
{
  "status": "ok"
}
```

---

### Search Flights

Search for available flights across all configured airline providers.

```http
POST /api/v1/flights/search
Content-Type: application/json
```

#### Request Body

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "passengers": 1,
  "class": "economy",
  "filters": {
    "maxPrice": 2000000,
    "maxStops": 1,
    "airlines": ["GA", "JT"],
    "departureTimeRange": {
      "start": "06:00",
      "end": "12:00"
    },
    "arrivalTimeRange": {
      "start": "08:00",
      "end": "17:00"
    },
    "durationRange": {
      "minMinutes": 60,
      "maxMinutes": 240
    }
  },
  "sortBy": "best"
}
```

#### Request Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `origin` | string | ✅ Yes | IATA airport code (3 uppercase letters) | `"CGK"` |
| `destination` | string | ✅ Yes | IATA airport code (3 uppercase letters) | `"DPS"` |
| `departureDate` | string | ✅ Yes | Date in YYYY-MM-DD format (must be today or future) | `"2025-12-15"` |
| `passengers` | integer | ✅ Yes | Number of passengers (1-9) | `1` |
| `class` | string | No | Travel class | `"economy"`, `"business"`, `"first"` |
| `filters` | object | No | Optional filtering criteria | See below |
| `sortBy` | string | No | Sort order (default: `"best"`) | `"best"`, `"price"`, `"duration"`, `"departure"` |

#### Filter Object

All filter fields are optional. Filters are combined with AND logic.

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `maxPrice` | number | Maximum price in IDR | `2000000` |
| `maxStops` | integer | Maximum stops (0 = direct flights only) | `1` |
| `airlines` | array | Airline codes to include (case-sensitive) | `["GA", "JT", "ID"]` |
| `departureTimeRange` | object | Departure time window (time-of-day only) | `{"start": "06:00", "end": "12:00"}` |
| `arrivalTimeRange` | object | Arrival time window (time-of-day only) | `{"start": "08:00", "end": "17:00"}` |
| `durationRange` | object | Flight duration range in minutes | `{"minMinutes": 60, "maxMinutes": 240}` |

#### Time Range Object

Used for `departureTimeRange` and `arrivalTimeRange` filters. Compares time-of-day only (ignores date and timezone).

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `start` | string | ✅ Yes | Start time in HH:MM format (24-hour) | `"06:00"` |
| `end` | string | ✅ Yes | End time in HH:MM format (24-hour) | `"12:00"` |

**Validation:**
- Both `start` and `end` are required
- Must be in HH:MM format (hours: 00-23, minutes: 00-59)
- `start` must be before `end` (no overnight ranges)

#### Duration Range Object

Used for `durationRange` filter to limit flight duration.

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `minMinutes` | integer | No | Minimum duration in minutes | `60` (1 hour) |
| `maxMinutes` | integer | No | Maximum duration in minutes | `240` (4 hours) |

**Validation:**
- At least one of `minMinutes` or `maxMinutes` must be specified
- Values must be positive integers
- If both specified, `minMinutes` must be ≤ `maxMinutes`

#### Sort Options

| Value | Description |
|-------|-------------|
| `best` | Best value score - weighted combination of price, duration, and stops (default) |
| `price` | Lowest price first |
| `duration` | Shortest flight duration first |
| `departure` | Earliest departure time first |

---

#### Success Response

**200 OK**

```json
{
  "flights": [
    {
      "id": "garuda-1",
      "flightNumber": "GA 123",
      "airline": {
        "code": "GA",
        "name": "Garuda Indonesia",
        "logo": ""
      },
      "departure": {
        "airportCode": "CGK",
        "airportName": "Soekarno-Hatta International Airport",
        "terminal": "3",
        "dateTime": "2025-12-15T08:00:00Z",
        "timezone": "Asia/Jakarta"
      },
      "arrival": {
        "airportCode": "DPS",
        "airportName": "Ngurah Rai International Airport",
        "terminal": "D",
        "dateTime": "2025-12-15T10:45:00Z",
        "timezone": "Asia/Makassar"
      },
      "duration": {
        "totalMinutes": 165,
        "formatted": "2h 45m"
      },
      "price": {
        "amount": 1350000,
        "currency": "IDR",
        "formatted": "IDR 1,350,000"
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

#### Response Fields

##### Flight Object

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique flight identifier |
| `flightNumber` | string | Airline flight number |
| `airline` | object | Airline information |
| `departure` | object | Departure details |
| `arrival` | object | Arrival details |
| `duration` | object | Flight duration |
| `price` | object | Pricing information |
| `baggage` | object | Baggage allowance |
| `class` | string | Travel class |
| `stops` | integer | Number of stops |
| `provider` | string | Source provider identifier |
| `rankingScore` | number | Calculated ranking score (0-1, higher is better) |

##### Metadata Object

| Field | Type | Description |
|-------|------|-------------|
| `totalResults` | integer | Total flights returned |
| `searchDurationMs` | integer | Search execution time in milliseconds |
| `providersQueried` | array | List of providers that were queried |
| `providersFailed` | array | List of providers that failed or timed out |

---

#### Error Responses

##### 400 Bad Request - Validation Error

Returned when request validation fails.

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "Request validation failed",
    "details": {
      "origin": "origin is required",
      "departureDate": "departureDate cannot be in the past",
      "passengers": "passengers must be at least 1"
    }
  }
}
```

**Common Validation Errors:**

| Field | Error | Cause |
|-------|-------|-------|
| `origin` | "origin is required" | Missing origin field |
| `origin` | "origin must be a valid 3-letter IATA airport code" | Invalid format (e.g., lowercase, numbers, wrong length) |
| `destination` | "origin and destination must be different" | Same airport for origin and destination |
| `departureDate` | "departureDate is required" | Missing date |
| `departureDate` | "departureDate must be in YYYY-MM-DD format" | Invalid format |
| `departureDate` | "departureDate cannot be in the past" | Past date (disabled in mock mode) |
| `passengers` | "passengers must be at least 1" | Zero or negative |
| `passengers` | "passengers cannot exceed 9" | Too many passengers |
| `filters.departureTimeRange.start` | "start time must be in HH:MM format" | Invalid time format |
| `filters.departureTimeRange` | "start time must be before end time" | Invalid range (start ≥ end) |
| `filters.arrivalTimeRange.start` | "start time must be in HH:MM format" | Invalid time format |
| `filters.arrivalTimeRange` | "start time must be before end time" | Invalid range (start ≥ end) |
| `filters.durationRange` | "minMinutes must be less than or equal to maxMinutes" | Invalid range (min > max) |
| `filters.durationRange` | "minMinutes must be positive" | Negative or zero value |
| `filters.maxPrice` | "maxPrice must be positive" | Negative or zero value |
| `filters.maxStops` | "maxStops must be non-negative" | Negative value |

##### 503 Service Unavailable

Returned when all airline providers fail to respond.

```json
{
  "success": false,
  "error": {
    "code": "service_unavailable",
    "message": "All flight providers are currently unavailable"
  }
}
```

##### 504 Gateway Timeout

Returned when the request exceeds the global timeout.

```json
{
  "success": false,
  "error": {
    "code": "timeout",
    "message": "Request timed out"
  }
}
```

##### 500 Internal Server Error

Returned for unexpected server errors.

```json
{
  "success": false,
  "error": {
    "code": "internal_error",
    "message": "An unexpected error occurred"
  }
}
```

---

## Examples

### Basic Search

```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1
  }'
```

### Search with Filters

```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 2,
    "class": "economy",
    "filters": {
      "maxPrice": 1500000,
      "maxStops": 0,
      "airlines": ["GA"]
    },
    "sortBy": "price"
  }'
```

### Morning Flights Only

```bash
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
      }
    },
    "sortBy": "departure"
  }'
```

### Business Hours Arrival

```bash
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
    "sortBy": "best"
  }'
```

### Short Flights Only (Under 3 Hours)

```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "filters": {
      "durationRange": {
        "maxMinutes": 180
      }
    },
    "sortBy": "duration"
  }'
```

### Combined Filters: Budget Morning Flights with Business Hours Arrival

```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "class": "economy",
    "filters": {
      "maxPrice": 1200000,
      "maxStops": 0,
      "departureTimeRange": {
        "start": "05:00",
        "end": "09:00"
      },
      "arrivalTimeRange": {
        "start": "08:00",
        "end": "12:00"
      },
      "durationRange": {
        "minMinutes": 60,
        "maxMinutes": 180
      }
    },
    "sortBy": "price"
  }'
```

---

## Airline Providers

The system aggregates flights from the following providers:

| Provider | Code | Name |
|----------|------|------|
| `garuda_indonesia` | GA | Garuda Indonesia |
| `lion_air` | JT | Lion Air |
| `batik_air` | ID | Batik Air |
| `airasia` | QZ | AirAsia |

---

## Rate Limiting

Currently, no rate limiting is implemented. For production deployments, consider adding rate limiting middleware.

---

## Changelog

### v1.0.0

- Initial API release
- Multi-provider flight search
- Filtering and sorting capabilities
- Partial failure handling
