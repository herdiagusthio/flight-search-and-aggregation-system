package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/http/response"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
)

// mockUseCase is a mock implementation of FlightSearchUseCase for testing.
type mockUseCase struct {
	searchFunc func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error)
}

func (m *mockUseCase) Search(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, criteria, opts)
	}
	return &domain.SearchResponse{
		SearchCriteria: domain.SearchCriteriaResponse{
			Origin:        criteria.Origin,
			Destination:   criteria.Destination,
			DepartureDate: criteria.DepartureDate,
			Passengers:    criteria.Passengers,
			CabinClass:    criteria.Class,
		},
		Flights: []domain.Flight{},
		Metadata: domain.SearchMetadata{
			TotalResults:       0,
			SearchTimeMs:       100,
			ProvidersQueried:   1,
			ProvidersSucceeded: 1,
			ProvidersFailed:    0,
			CacheHit:           false,
		},
	}, nil
}

// setupTestHandler creates a test Echo instance and FlightHandler.
func setupTestHandler(uc usecase.FlightSearchUseCase) (*echo.Echo, *FlightHandler) {
	e := echo.New()
	h := NewFlightHandler(uc)
	RegisterRoutes(e, h)
	return e, h
}

// makeRequest is a helper to make test requests.
func makeRequest(e *echo.Echo, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// getFutureDate returns a date string for tomorrow.
func getFutureDate() string {
	return time.Now().AddDate(0, 0, 1).Format("2006-01-02")
}

// getPastDate returns a date string for yesterday.
func getPastDate() string {
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
}

// =====================================================
// Handler Tests
// =====================================================

func TestSearchFlights_Success(t *testing.T) {
	mockFlights := []domain.Flight{
		{
			ID:           "1",
			FlightNumber: "GA-123",
			Airline:      domain.AirlineInfo{Code: "GA", Name: "Garuda Indonesia"},
			Price:        domain.PriceInfo{Amount: 1000000, Currency: "IDR"},
			Duration:     domain.DurationInfo{TotalMinutes: 120, Formatted: "2h 0m"},
			Stops:        0,
		},
	}

	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			return &domain.SearchResponse{
				SearchCriteria: domain.SearchCriteriaResponse{
					Origin:        criteria.Origin,
					Destination:   criteria.Destination,
					DepartureDate: criteria.DepartureDate,
					Passengers:    criteria.Passengers,
					CabinClass:    criteria.Class,
				},
				Flights: mockFlights,
				Metadata: domain.SearchMetadata{
					TotalResults:       1,
					SearchTimeMs:       150,
					ProvidersQueried:   1,
					ProvidersSucceeded: 1,
					ProvidersFailed:    0,
					CacheHit:           false,
				},
			}, nil
		},
	}

	e, _ := setupTestHandler(mock)

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response domain.SearchResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, response.Metadata.TotalResults)
	assert.Len(t, response.Flights, 1)
}

func TestSearchFlights_WithFilters(t *testing.T) {
	var capturedOpts usecase.SearchOptions

	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			capturedOpts = opts
			return &domain.SearchResponse{
				SearchCriteria: domain.SearchCriteriaResponse{
					Origin:        criteria.Origin,
					Destination:   criteria.Destination,
					DepartureDate: criteria.DepartureDate,
					Passengers:    criteria.Passengers,
					CabinClass:    criteria.Class,
				},
				Flights:  []domain.Flight{},
				Metadata: domain.SearchMetadata{
					TotalResults:       0,
					ProvidersQueried:   1,
					ProvidersSucceeded: 1,
					ProvidersFailed:    0,
					CacheHit:           false,
				},
			}, nil
		},
	}

	e, _ := setupTestHandler(mock)

	maxPrice := float64(1500000)
	maxStops := 1
	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    2,
		Class:         "business",
		Filters: &FilterDTO{
			MaxPrice: &maxPrice,
			MaxStops: &maxStops,
			Airlines: []string{"GA", "JT"},
		},
		SortBy: "price",
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedOpts.Filters)
	assert.Equal(t, &maxPrice, capturedOpts.Filters.MaxPrice)
	assert.Equal(t, &maxStops, capturedOpts.Filters.MaxStops)
	assert.Equal(t, []string{"GA", "JT"}, capturedOpts.Filters.Airlines)
	assert.Equal(t, domain.SortByPrice, capturedOpts.SortBy)
}

func TestSearchFlights_InvalidJSON(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/flights/search",
		strings.NewReader(`{invalid json`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, response.CodeInvalidRequest, errResp.Code)
}

func TestSearchFlights_MissingRequiredFields(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	tests := []struct {
		name          string
		request       SearchFlightsRequest
		expectedField string
	}{
		{
			name:          "missing origin",
			request:       SearchFlightsRequest{Destination: "DPS", DepartureDate: getFutureDate(), Passengers: 1},
			expectedField: "origin",
		},
		{
			name:          "missing destination",
			request:       SearchFlightsRequest{Origin: "CGK", DepartureDate: getFutureDate(), Passengers: 1},
			expectedField: "destination",
		},
		{
			name:          "missing departureDate",
			request:       SearchFlightsRequest{Origin: "CGK", Destination: "DPS", Passengers: 1},
			expectedField: "departureDate",
		},
		{
			name:          "missing passengers (zero)",
			request:       SearchFlightsRequest{Origin: "CGK", Destination: "DPS", DepartureDate: getFutureDate(), Passengers: 0},
			expectedField: "passengers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", tt.request)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResp response.ErrorDetail
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, response.CodeValidationError, errResp.Code)
			assert.Contains(t, errResp.Details, tt.expectedField)
		})
	}
}

func TestSearchFlights_InvalidAirportCode(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	tests := []struct {
		name          string
		origin        string
		destination   string
		expectedField string
	}{
		{
			name:          "origin too short",
			origin:        "CG",
			destination:   "DPS",
			expectedField: "origin",
		},
		{
			name:          "origin too long",
			origin:        "CGKX",
			destination:   "DPS",
			expectedField: "origin",
		},
		{
			name:          "origin with numbers",
			origin:        "CG1",
			destination:   "DPS",
			expectedField: "origin",
		},
		{
			name:          "destination invalid",
			origin:        "CGK",
			destination:   "12",
			expectedField: "destination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SearchFlightsRequest{
				Origin:        tt.origin,
				Destination:   tt.destination,
				DepartureDate: getFutureDate(),
				Passengers:    1,
			}

			rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResp response.ErrorDetail
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Contains(t, errResp.Details, tt.expectedField)
		})
	}
}

func TestSearchFlights_InvalidDateFormat(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	tests := []struct {
		name string
		date string
	}{
		{"wrong format", "15-12-2025"},
		{"invalid format", "2025/12/15"},
		{"incomplete date", "2025-12"},
		{"text date", "tomorrow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SearchFlightsRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: tt.date,
				Passengers:    1,
			}

			rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResp response.ErrorDetail
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Contains(t, errResp.Details, "departureDate")
		})
	}
}

func TestSearchFlights_SameOriginDestination(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "CGK",
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Details, "destination")
	assert.Contains(t, errResp.Details["destination"], "different")
}

func TestSearchFlights_InvalidPassengers(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	tests := []struct {
		name       string
		passengers int
		expected   string
	}{
		{"zero passengers", 0, "at least 1"},
		{"too many passengers", 10, "exceed 9"},
		{"negative passengers", -1, "at least 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SearchFlightsRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: getFutureDate(),
				Passengers:    tt.passengers,
			}

			rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResp response.ErrorDetail
			err := json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Contains(t, errResp.Details, "passengers")
		})
	}
}

func TestSearchFlights_InvalidClass(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
		Class:         "premium",
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Details, "class")
}

func TestSearchFlights_InvalidSortBy(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
		SortBy:        "invalid_sort",
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Details, "sortBy")
}

func TestSearchFlights_AllProvidersFailed(t *testing.T) {
	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			return nil, domain.ErrAllProvidersFailed
		},
	}

	e, _ := setupTestHandler(mock)

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, response.CodeServiceUnavailable, errResp.Code)
}

func TestSearchFlights_Timeout(t *testing.T) {
	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			return nil, context.DeadlineExceeded
		},
	}

	e, _ := setupTestHandler(mock)

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)

	var errResp response.ErrorDetail
	err := json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, response.CodeTimeout, errResp.Code)
}

func TestSearchFlights_EmptyResults(t *testing.T) {
	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			return &domain.SearchResponse{
				SearchCriteria: domain.SearchCriteriaResponse{
					Origin:        criteria.Origin,
					Destination:   criteria.Destination,
					DepartureDate: criteria.DepartureDate,
					Passengers:    criteria.Passengers,
					CabinClass:    criteria.Class,
				},
				Flights: []domain.Flight{},
				Metadata: domain.SearchMetadata{
					TotalResults:       0,
					SearchTimeMs:       100,
					ProvidersQueried:   1,
					ProvidersSucceeded: 1,
					ProvidersFailed:    0,
					CacheHit:           false,
				},
			}, nil
		},
	}

	e, _ := setupTestHandler(mock)

	req := SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	// Empty results should still return 200
	assert.Equal(t, http.StatusOK, rec.Code)

	var response domain.SearchResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 0, response.Metadata.TotalResults)
	assert.Empty(t, response.Flights)
}

func TestSearchFlights_LowercaseOriginDestination(t *testing.T) {
	var capturedCriteria domain.SearchCriteria

	mock := &mockUseCase{
		searchFunc: func(ctx context.Context, criteria domain.SearchCriteria, opts usecase.SearchOptions) (*domain.SearchResponse, error) {
			capturedCriteria = criteria
			return &domain.SearchResponse{
				SearchCriteria: domain.SearchCriteriaResponse{
					Origin:        criteria.Origin,
					Destination:   criteria.Destination,
					DepartureDate: criteria.DepartureDate,
					Passengers:    criteria.Passengers,
					CabinClass:    criteria.Class,
				},
				Flights:  []domain.Flight{},
				Metadata: domain.SearchMetadata{
					TotalResults:       0,
					ProvidersQueried:   1,
					ProvidersSucceeded: 1,
					ProvidersFailed:    0,
					CacheHit:           false,
				},
			}, nil
		},
	}

	e, _ := setupTestHandler(mock)

	req := SearchFlightsRequest{
		Origin:        "cgk", // lowercase
		Destination:   "dps", // lowercase
		DepartureDate: getFutureDate(),
		Passengers:    1,
	}

	rec := makeRequest(e, http.MethodPost, "/api/v1/flights/search", req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Should be normalized to uppercase
	assert.Equal(t, "CGK", capturedCriteria.Origin)
	assert.Equal(t, "DPS", capturedCriteria.Destination)
}

func TestHealth_Success(t *testing.T) {
	e, _ := setupTestHandler(&mockUseCase{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.HealthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
}

// =====================================================
// Validation Tests
// =====================================================

func TestSearchFlightsRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		request   SearchFlightsRequest
		wantErr   bool
		errFields []string
	}{
		{
			name: "valid request",
			request: SearchFlightsRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: getFutureDate(),
				Passengers:    1,
			},
			wantErr: false,
		},
		{
			name: "valid request with all options",
			request: SearchFlightsRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: getFutureDate(),
				Passengers:    3,
				Class:         "business",
				SortBy:        "price",
				Filters: &FilterDTO{
					MaxPrice: floatPtr(1500000),
					MaxStops: intPtr(1),
					Airlines: []string{"GA"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple errors",
			request: SearchFlightsRequest{
				Origin:        "",
				Destination:   "",
				DepartureDate: "",
				Passengers:    0,
			},
			wantErr:   true,
			errFields: []string{"origin", "destination", "departureDate", "passengers"},
		},
		{
			name: "negative filter values",
			request: SearchFlightsRequest{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: getFutureDate(),
				Passengers:    1,
				Filters: &FilterDTO{
					MaxPrice: floatPtr(-100),
					MaxStops: intPtr(-1),
				},
			},
			wantErr:   true,
			errFields: []string{"filters.maxPrice", "filters.maxStops"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				require.Error(t, err)
				var validationErrs *ValidationErrors
				if errors.As(err, &validationErrs) {
					for _, field := range tt.errFields {
						found := false
						for _, e := range validationErrs.Errors {
							if e.Field == field {
								found = true
								break
							}
						}
						assert.True(t, found, "expected error for field %s", field)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =====================================================
// Converter Tests
// =====================================================

func TestToDomainCriteria(t *testing.T) {
	req := &SearchFlightsRequest{
		Origin:        "cgk", // lowercase
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		Class:         "business",
	}

	criteria := ToDomainCriteria(req)

	assert.Equal(t, "CGK", criteria.Origin)
	assert.Equal(t, "DPS", criteria.Destination)
	assert.Equal(t, "2025-12-15", criteria.DepartureDate)
	assert.Equal(t, 2, criteria.Passengers)
	assert.Equal(t, "business", criteria.Class)
}

func TestToDomainCriteria_Defaults(t *testing.T) {
	req := &SearchFlightsRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    0, // Should default to 1
		Class:         "", // Should default to economy
	}

	criteria := ToDomainCriteria(req)

	assert.Equal(t, 1, criteria.Passengers)
	assert.Equal(t, "economy", criteria.Class)
}

func TestToDomainFilters(t *testing.T) {
	maxPrice := float64(1500000)
	maxStops := 1

	dto := &FilterDTO{
		MaxPrice: &maxPrice,
		MaxStops: &maxStops,
		Airlines: []string{"GA", "JT"},
	}

	filters := ToDomainFilters(dto)

	require.NotNil(t, filters)
	assert.Equal(t, &maxPrice, filters.MaxPrice)
	assert.Equal(t, &maxStops, filters.MaxStops)
	assert.Equal(t, []string{"GA", "JT"}, filters.Airlines)
}

func TestToDomainFilters_Nil(t *testing.T) {
	filters := ToDomainFilters(nil)
	assert.Nil(t, filters)
}

func TestToDomainSortOption(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.SortOption
	}{
		{"best", domain.SortByBestValue},
		{"best_value", domain.SortByBestValue},
		{"price", domain.SortByPrice},
		{"duration", domain.SortByDuration},
		{"departure", domain.SortByDeparture},
		{"", domain.SortByBestValue},
		{"invalid", domain.SortByBestValue},
		{"PRICE", domain.SortByPrice}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToDomainSortOption(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =====================================================
// Route Registration Tests
// =====================================================

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	h := NewFlightHandler(&mockUseCase{})

	RegisterRoutes(e, h)

	// Test that routes are registered
	routes := e.Routes()

	// Check for expected routes
	expectedPaths := map[string]string{
		"/health":               http.MethodGet,
		"/api/v1/flights/search": http.MethodPost,
	}

	for path, method := range expectedPaths {
		found := false
		for _, r := range routes {
			if r.Path == path && r.Method == method {
				found = true
				break
			}
		}
		assert.True(t, found, "expected route %s %s not found", method, path)
	}
}

// Helper functions for creating pointer values
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
