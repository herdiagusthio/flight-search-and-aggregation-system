package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderError(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		underlyingErr  error
		wantContains   []string
		wantUnwrapable bool
		wantRetryable  bool
	}{
		{
			name:           "error message includes provider and underlying error",
			provider:       "garuda",
			underlyingErr:  errors.New("connection failed"),
			wantContains:   []string{"garuda", "connection failed"},
			wantUnwrapable: true,
			wantRetryable:  false, // Default is non-retryable
		},
		{
			name:           "error message with different provider",
			provider:       "lionair",
			underlyingErr:  errors.New("timeout"),
			wantContains:   []string{"lionair", "timeout"},
			wantUnwrapable: true,
			wantRetryable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProviderError(tt.provider, tt.underlyingErr)

			for _, want := range tt.wantContains {
				assert.Contains(t, err.Error(), want)
			}

			if tt.wantUnwrapable {
				assert.True(t, errors.Is(err, tt.underlyingErr))
			}

			assert.Equal(t, tt.wantRetryable, err.Retryable)
		})
	}
}

func TestNewRetryableProviderError(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		underlyingErr error
	}{
		{
			name:          "retryable network error",
			provider:      "garuda",
			underlyingErr: errors.New("temporary network failure"),
		},
		{
			name:          "retryable rate limit error",
			provider:      "lionair",
			underlyingErr: errors.New("rate limit exceeded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRetryableProviderError(tt.provider, tt.underlyingErr)

			assert.Contains(t, err.Error(), tt.provider)
			assert.True(t, errors.Is(err, tt.underlyingErr))
			assert.True(t, err.Retryable)
		})
	}
}

func TestNewProviderTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{name: "lionair provider", provider: "lionair"},
		{name: "garuda provider", provider: "garuda"},
		{name: "airasia provider", provider: "airasia"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProviderTimeoutError(tt.provider)
			assert.Contains(t, err.Error(), tt.provider)
			assert.True(t, errors.Is(err, ErrProviderTimeout))
		})
	}
}

func TestNewProviderUnavailableError(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{name: "airasia provider", provider: "airasia"},
		{name: "batikair provider", provider: "batikair"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewProviderUnavailableError(tt.provider)
			assert.Contains(t, err.Error(), tt.provider)
			assert.True(t, errors.Is(err, ErrProviderUnavailable))
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		message     string
		wantError   string
	}{
		{
			name:      "origin field validation",
			field:     "origin",
			message:   "must be a 3-letter code",
			wantError: "origin: must be a 3-letter code",
		},
		{
			name:      "passengers field validation",
			field:     "passengers",
			message:   "must be at least 1",
			wantError: "passengers: must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message)
			assert.Equal(t, tt.wantError, err.Error())
			assert.Equal(t, tt.field, err.Field)
			assert.Equal(t, tt.message, err.Message)
		})
	}
}

func TestWrapInvalidRequest(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		args         []interface{}
		wantContains string
	}{
		{
			name:         "single argument",
			format:       "field %s is required",
			args:         []interface{}{"origin"},
			wantContains: "field origin is required",
		},
		{
			name:         "multiple arguments",
			format:       "%s must be between %d and %d",
			args:         []interface{}{"passengers", 1, 9},
			wantContains: "passengers must be between 1 and 9",
		},
		{
			name:         "no arguments",
			format:       "invalid request format",
			args:         nil,
			wantContains: "invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapInvalidRequest(tt.format, tt.args...)
			assert.True(t, errors.Is(err, ErrInvalidRequest))
			assert.Contains(t, err.Error(), tt.wantContains)
		})
	}
}

func TestErrorCheckers(t *testing.T) {
	tests := []struct {
		name       string
		checkFunc  func(error) bool
		err        error
		wantResult bool
	}{
		// IsInvalidRequest tests
		{
			name:       "IsInvalidRequest with ErrInvalidRequest",
			checkFunc:  IsInvalidRequest,
			err:        ErrInvalidRequest,
			wantResult: true,
		},
		{
			name:       "IsInvalidRequest with wrapped error",
			checkFunc:  IsInvalidRequest,
			err:        WrapInvalidRequest("test"),
			wantResult: true,
		},
		{
			name:       "IsInvalidRequest with different error",
			checkFunc:  IsInvalidRequest,
			err:        ErrAllProvidersFailed,
			wantResult: false,
		},
		// IsAllProvidersFailed tests
		{
			name:       "IsAllProvidersFailed with ErrAllProvidersFailed",
			checkFunc:  IsAllProvidersFailed,
			err:        ErrAllProvidersFailed,
			wantResult: true,
		},
		{
			name:       "IsAllProvidersFailed with different error",
			checkFunc:  IsAllProvidersFailed,
			err:        ErrInvalidRequest,
			wantResult: false,
		},
		// IsProviderTimeout tests
		{
			name:       "IsProviderTimeout with ErrProviderTimeout",
			checkFunc:  IsProviderTimeout,
			err:        ErrProviderTimeout,
			wantResult: true,
		},
		{
			name:       "IsProviderTimeout with wrapped timeout error",
			checkFunc:  IsProviderTimeout,
			err:        NewProviderTimeoutError("test"),
			wantResult: true,
		},
		{
			name:       "IsProviderTimeout with different error",
			checkFunc:  IsProviderTimeout,
			err:        ErrInvalidRequest,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.checkFunc(tt.err))
		})
	}
}
