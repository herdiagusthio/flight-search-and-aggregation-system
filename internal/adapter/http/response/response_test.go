package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEcho() (*echo.Echo, echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return e, c, rec
}

func TestHealth(t *testing.T) {
	_, c, rec := setupEcho()

	err := Health(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "ok", result.Status)
}

func TestBadRequest(t *testing.T) {
	_, c, rec := setupEcho()

	err := BadRequest(c, "Invalid input")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeInvalidRequest, result.Code)
	assert.Equal(t, "Invalid input", result.Message)
}

func TestInvalidRequestBody(t *testing.T) {
	_, c, rec := setupEcho()

	err := InvalidRequestBody(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeInvalidRequest, result.Code)
	assert.Equal(t, MsgInvalidRequestBody, result.Message)
}

func TestValidationError(t *testing.T) {
	_, c, rec := setupEcho()

	details := map[string]string{
		"email": "must be a valid email",
		"name":  "is required",
	}
	err := ValidationError(c, details)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeValidationError, result.Code)
	assert.Equal(t, MsgValidationFailed, result.Message)
	assert.Equal(t, "must be a valid email", result.Details["email"])
	assert.Equal(t, "is required", result.Details["name"])
}

func TestValidationErrorWithMessage(t *testing.T) {
	_, c, rec := setupEcho()

	err := ValidationErrorWithMessage(c, "Custom validation message")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeValidationError, result.Code)
	assert.Equal(t, "Custom validation message", result.Message)
}

func TestServiceUnavailable(t *testing.T) {
	_, c, rec := setupEcho()

	err := ServiceUnavailable(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeServiceUnavailable, result.Code)
	assert.Equal(t, MsgServiceUnavailable, result.Message)
}

func TestServiceUnavailableWithMessage(t *testing.T) {
	_, c, rec := setupEcho()

	err := ServiceUnavailableWithMessage(c, "Database is down")

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeServiceUnavailable, result.Code)
	assert.Equal(t, "Database is down", result.Message)
}

func TestGatewayTimeout(t *testing.T) {
	_, c, rec := setupEcho()

	err := GatewayTimeout(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeTimeout, result.Code)
	assert.Equal(t, MsgTimeout, result.Message)
}

func TestRequestCancelled(t *testing.T) {
	_, c, rec := setupEcho()

	err := RequestCancelled(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusGatewayTimeout, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeTimeout, result.Code)
	assert.Equal(t, MsgRequestCancelled, result.Message)
}

func TestInternalServerError(t *testing.T) {
	_, c, rec := setupEcho()

	err := InternalServerError(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, CodeInternalError, result.Code)
	assert.Equal(t, MsgInternalError, result.Message)
}

func TestSearchResults(t *testing.T) {
	_, c, rec := setupEcho()

	results := struct {
		Items []string `json:"items"`
		Total int      `json:"total"`
	}{
		Items: []string{"a", "b", "c"},
		Total: 3,
	}

	err := SearchResults(c, results)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Items []string `json:"items"`
		Total int      `json:"total"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 3, resp.Total)
	assert.Len(t, resp.Items, 3)
}
