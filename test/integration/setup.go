// Package integration provides helpers and integration tests for the flight search system.
// Integration tests verify that components work together correctly, including
// HTTP handlers, use cases, and mock providers.
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo/v4"

	httpAdapter "github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/http"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
)

// TestServer wraps an Echo instance and provides helper methods for integration testing.
type TestServer struct {
	Echo    *echo.Echo
	Handler *httpAdapter.FlightHandler
}

// NewTestServer creates a new test server with the given use case.
func NewTestServer(uc usecase.FlightSearchUseCase) *TestServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	handler := httpAdapter.NewFlightHandler(uc)
	httpAdapter.RegisterRoutes(e, handler)

	return &TestServer{
		Echo:    e,
		Handler: handler,
	}
}

// Request represents a test HTTP request configuration.
type Request struct {
	Method      string
	Path        string
	Body        interface{}
	ContentType string
}

// Response represents a test HTTP response.
type Response struct {
	Code    int
	Body    []byte
	Headers http.Header
}

// Do executes a test request and returns the response.
func (ts *TestServer) Do(req Request) Response {
	var bodyReader *bytes.Reader
	if req.Body != nil {
		bodyBytes, _ := json.Marshal(req.Body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, bodyReader)

	if req.ContentType != "" {
		httpReq.Header.Set(echo.HeaderContentType, req.ContentType)
	} else if req.Body != nil {
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	rec := httptest.NewRecorder()
	ts.Echo.ServeHTTP(rec, httpReq)

	return Response{
		Code:    rec.Code,
		Body:    rec.Body.Bytes(),
		Headers: rec.Header(),
	}
}

// SearchRequest creates a search request with the given parameters.
func (ts *TestServer) SearchRequest(body interface{}) Response {
	return ts.Do(Request{
		Method: http.MethodPost,
		Path:   "/api/v1/flights/search",
		Body:   body,
	})
}

// HealthRequest makes a health check request.
func (ts *TestServer) HealthRequest() Response {
	return ts.Do(Request{
		Method: http.MethodGet,
		Path:   "/health",
	})
}

// ParseSearchResponse parses the response body as a SearchResponse.
func (r *Response) ParseSearchResponse() (*domain.SearchResponse, error) {
	var resp domain.SearchResponse
	if err := json.Unmarshal(r.Body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ParseError parses the response body to extract error information.
func (r *Response) ParseError() (map[string]interface{}, error) {
	var errResp map[string]interface{}
	if err := json.Unmarshal(r.Body, &errResp); err != nil {
		return nil, err
	}
	return errResp, nil
}

// SearchRequestBody is a helper struct for building search request bodies.
type SearchRequestBody struct {
	Origin        string                 `json:"origin"`
	Destination   string                 `json:"destination"`
	DepartureDate string                 `json:"departureDate"`
	Passengers    int                    `json:"passengers"`
	Class         string                 `json:"class,omitempty"`
	Filters       map[string]interface{} `json:"filters,omitempty"`
	SortBy        string                 `json:"sortBy,omitempty"`
}

// DefaultSearchRequest returns a valid search request body for testing.
// Uses a date 30 days in the future to avoid past date validation errors.
func DefaultSearchRequest() SearchRequestBody {
	futureDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	return SearchRequestBody{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: futureDate,
		Passengers:    1,
	}
}

// CreateUseCase creates a use case with the given providers and default configuration.
func CreateUseCase(providers []domain.FlightProvider) usecase.FlightSearchUseCase {
	return usecase.NewFlightSearchUseCase(providers, nil)
}

// CreateUseCaseWithConfig creates a use case with custom configuration.
func CreateUseCaseWithConfig(providers []domain.FlightProvider, config *usecase.Config) usecase.FlightSearchUseCase {
	return usecase.NewFlightSearchUseCase(providers, config)
}

// FutureDate returns a date string 30 days in the future in YYYY-MM-DD format.
func FutureDate() string {
	return time.Now().AddDate(0, 0, 30).Format("2006-01-02")
}

// DefaultSearchCriteria returns a valid search criteria for testing use case directly.
func DefaultSearchCriteria() domain.SearchCriteria {
	return domain.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: FutureDate(),
		Passengers:    1,
	}
}
