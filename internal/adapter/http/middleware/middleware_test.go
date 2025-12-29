package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// Request ID Middleware Tests
// =====================================================

func TestRequestID_GeneratesNewID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)

	// Check response header contains request ID
	reqID := rec.Header().Get(RequestIDHeader)
	assert.NotEmpty(t, reqID, "should generate request ID")
	assert.Len(t, reqID, 36, "should be UUID format (36 chars)")

	// Check context has the same request ID
	ctxID := GetRequestID(c)
	assert.Equal(t, reqID, ctxID, "context ID should match header ID")
}

func TestRequestID_PropagatesExistingID(t *testing.T) {
	e := echo.New()
	existingID := "existing-request-id-12345"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)

	// Check response header contains the original ID
	respID := rec.Header().Get(RequestIDHeader)
	assert.Equal(t, existingID, respID, "should propagate existing request ID")

	// Check context has the same ID
	ctxID := GetRequestID(c)
	assert.Equal(t, existingID, ctxID)
}

func TestGetRequestID_ReturnsEmptyWhenNotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Don't run middleware, just check GetRequestID
	reqID := GetRequestID(c)
	assert.Empty(t, reqID, "should return empty string when not set")
}

// =====================================================
// Request Logging Middleware Tests
// =====================================================

func TestRequestLogger_LogsRequestDetails(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf).With().Timestamp().Logger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test?foo=bar", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set request ID first (simulating middleware chain)
	c.Set("request_id", "test-req-id-123")

	handler := RequestLogger(logger)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)

	// Parse log output
	logOutput := logBuf.String()
	assert.NotEmpty(t, logOutput)

	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(logOutput), &logEntry)
	require.NoError(t, err, "log output should be valid JSON")

	// Verify logged fields
	assert.Equal(t, "test-req-id-123", logEntry["request_id"])
	assert.Equal(t, "POST", logEntry["method"])
	assert.Equal(t, "/api/v1/test", logEntry["path"])
	assert.Equal(t, "foo=bar", logEntry["query"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.Contains(t, logEntry, "duration_ms")
	assert.Equal(t, "TestAgent/1.0", logEntry["user_agent"])
	assert.Equal(t, "HTTP request", logEntry["message"])
}

func TestRequestLogger_LogsClientIP(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestLogger(logger)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "192.168.1.100", logEntry["client_ip"])
}

func TestRequestLogger_LogsErrorStatus(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestLogger(logger)(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "not found")
	})

	err := handler(c)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, float64(404), logEntry["status"])
	assert.Equal(t, "warn", logEntry["level"], "4xx should log at warn level")
}

func TestRequestLogger_LogsServerErrorStatus(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestLogger(logger)(func(c echo.Context) error {
		return c.String(http.StatusInternalServerError, "error")
	})

	err := handler(c)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, float64(500), logEntry["status"])
	assert.Equal(t, "error", logEntry["level"], "5xx should log at error level")
}

func TestRequestLogger_MeasuresDuration(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestLogger(logger)(func(c echo.Context) error {
		// Small delay to ensure duration > 0
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	duration, ok := logEntry["duration_ms"].(float64)
	assert.True(t, ok, "duration_ms should be a number")
	assert.GreaterOrEqual(t, duration, float64(0), "duration should be >= 0")
}

// =====================================================
// Recovery Middleware Tests
// =====================================================

func TestRecover_CatchesPanic(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set request ID for correlation
	c.Set("request_id", "panic-test-id")

	handler := Recover(logger)(func(c echo.Context) error {
		panic("test panic message")
	})

	// Should not panic - middleware catches it
	assert.NotPanics(t, func() {
		_ = handler(c)
	})
}

func TestRecover_Returns500OnPanic(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Recover(logger)(func(c echo.Context) error {
		panic("test panic")
	})

	_ = handler(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, false, response["success"])
	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "internal_error", errorObj["code"])
	assert.Equal(t, "An unexpected error occurred", errorObj["message"])
}

func TestRecover_LogsPanicWithStackTrace(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("request_id", "stack-test-id")

	handler := Recover(logger)(func(c echo.Context) error {
		panic("stack trace test panic")
	})

	_ = handler(c)

	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "error", logEntry["level"])
	assert.Equal(t, "stack-test-id", logEntry["request_id"])
	assert.Equal(t, "stack trace test panic", logEntry["panic"])
	assert.Contains(t, logEntry, "stack")
	stack, ok := logEntry["stack"].(string)
	assert.True(t, ok)
	assert.True(t, strings.Contains(stack, "goroutine"), "stack should contain goroutine info")
	assert.Equal(t, "Panic recovered", logEntry["message"])
}

func TestRecover_HandlesErrorPanic(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Recover(logger)(func(c echo.Context) error {
		var slice []int
		_ = slice[10] // Causes panic: index out of range
		return nil
	})

	assert.NotPanics(t, func() {
		_ = handler(c)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRecover_PassesThroughNormalRequests(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/normal", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Recover(logger)(func(c echo.Context) error {
		return c.String(http.StatusOK, "normal response")
	})

	err := handler(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "normal response", rec.Body.String())
	assert.Empty(t, logBuf.String(), "should not log anything for normal requests")
}

func TestRecoverWithConfig_DisableStackPrint(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	config := RecoveryConfig{
		DisablePrintStack: true,
	}

	handler := RecoverWithConfig(logger, config)(func(c echo.Context) error {
		panic("no stack test")
	})

	_ = handler(c)

	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.NotContains(t, logEntry, "stack", "stack should not be logged when disabled")
}

// =====================================================
// Integration Tests - Middleware Chain
// =====================================================

func TestMiddlewareChain_IntegrationOrder(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()

	// Apply middleware in correct order
	e.Use(RequestID())
	e.Use(RequestLogger(logger))
	e.Use(Recover(logger))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader))

	// Verify log contains request ID
	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuf.Bytes(), &logEntry)
	require.NoError(t, err)
	assert.NotEmpty(t, logEntry["request_id"])
}

func TestMiddlewareChain_PanicRecoveryWithLogging(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()

	e.Use(RequestID())
	e.Use(RequestLogger(logger))
	e.Use(Recover(logger))

	e.GET("/panic", func(c echo.Context) error {
		panic("integration test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	assert.NotPanics(t, func() {
		e.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader))
}

// =====================================================
// Setup Helper Tests
// =====================================================

func TestSetup_AppliesAllMiddleware(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	Setup(e, logger)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "setup test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader), "RequestID middleware should set header")
	assert.NotEmpty(t, logBuf.String(), "RequestLogger middleware should log")
}

func TestSetup_RecoversPanic(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	Setup(e, logger)

	e.GET("/panic", func(c echo.Context) error {
		panic("setup panic test")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		e.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestSetupWithConfig_AppliesCustomConfig(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	e := echo.New()
	config := RecoveryConfig{DisablePrintStack: true}
	SetupWithConfig(e, logger, config)

	e.GET("/panic", func(c echo.Context) error {
		panic("config panic test")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse log output - there may be multiple JSON entries, find the panic one
	logLines := strings.Split(strings.TrimSpace(logBuf.String()), "\n")
	var panicLogEntry map[string]interface{}
	for _, line := range logLines {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			if msg, ok := entry["message"].(string); ok && msg == "Panic recovered" {
				panicLogEntry = entry
				break
			}
		}
	}
	require.NotNil(t, panicLogEntry, "should have panic log entry")
	assert.NotContains(t, panicLogEntry, "stack", "stack should be disabled via config")
}

func TestChain_ReturnsMiddlewareSlice(t *testing.T) {
	var logBuf bytes.Buffer
	logger := zerolog.New(&logBuf)

	chain := Chain(logger)

	assert.Len(t, chain, 3, "Chain should return 3 middleware functions")

	// Verify they work by applying to echo
	e := echo.New()
	for _, mw := range chain {
		e.Use(mw)
	}

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "chain test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(RequestIDHeader))
}
