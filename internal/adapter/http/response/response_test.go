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

func TestOK(t *testing.T) {
	_, c, rec := setupEcho()

	data := map[string]string{"message": "success"}
	err := OK(c, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "success", result["message"])
}

func TestCreated(t *testing.T) {
	_, c, rec := setupEcho()

	data := map[string]int{"id": 123}
	err := Created(c, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var result map[string]int
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, 123, result["id"])
}

func TestNoContent(t *testing.T) {
	_, c, rec := setupEcho()

	err := NoContent(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
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

func TestNotFound(t *testing.T) {
	_, c, rec := setupEcho()

	err := NotFound(c, "Resource not found")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "not_found", result.Code)
	assert.Equal(t, "Resource not found", result.Message)
}

func TestUnauthorized(t *testing.T) {
	_, c, rec := setupEcho()

	err := Unauthorized(c, "Invalid token")

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "unauthorized", result.Code)
	assert.Equal(t, "Invalid token", result.Message)
}

func TestForbidden(t *testing.T) {
	_, c, rec := setupEcho()

	err := Forbidden(c, "Access denied")

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var result ErrorDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Equal(t, "forbidden", result.Code)
	assert.Equal(t, "Access denied", result.Message)
}

func TestSuccess(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := Success(data)

	assert.True(t, resp.Success)
	assert.Equal(t, data, resp.Data)
	assert.Nil(t, resp.Error)
}

func TestFailure(t *testing.T) {
	details := map[string]string{"field": "error"}
	resp := Failure("error_code", "Error message", details)

	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "error_code", resp.Error.Code)
	assert.Equal(t, "Error message", resp.Error.Message)
	assert.Equal(t, details, resp.Error.Details)
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

func TestList(t *testing.T) {
	_, c, rec := setupEcho()

	items := []string{"item1", "item2"}
	meta := map[string]int{"page": 1, "total": 10}

	err := List(c, items, meta)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["items"])
	assert.NotNil(t, resp["meta"])
}

func TestList_WithoutMeta(t *testing.T) {
	_, c, rec := setupEcho()

	items := []string{"item1", "item2"}

	err := List(c, items, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["items"])
	assert.Nil(t, resp["meta"])
}
