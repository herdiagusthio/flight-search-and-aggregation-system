# HTTP Adapter

This package contains HTTP handlers and middleware for the API.

## Contents

- `handler.go` - FlightHandler with search endpoint
- `request.go` - Request DTOs and validation
- `response.go` - Response DTOs and error responses
- `routes.go` - Route registration
- `middleware/` - Custom middleware (logging, recovery, request ID)

## Endpoints

- `POST /api/v1/flights/search` - Search for flights
- `GET /health` - Health check
