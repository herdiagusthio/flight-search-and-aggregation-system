package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
	}

	log := NewWithOutput(cfg, &buf)
	log.Info().Msg("test message")

	// Parse the JSON output
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "info", result["level"])
	assert.Equal(t, "test message", result["message"])
	assert.Equal(t, "test-service", result["service"])
	assert.NotEmpty(t, result["time"])
}

func TestNewLogger_ConsoleFormat(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "console",
		ServiceName: "test-service",
	}

	log := NewWithOutput(cfg, &buf)
	log.Info().Msg("test message")

	// Console format should be human-readable
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "INF")
}

func TestNewLogger_LogLevelFiltering(t *testing.T) {
	tests := []struct {
		name          string
		configLevel   string
		logLevel      string
		shouldLog     bool
	}{
		{"debug logged at debug level", "debug", "debug", true},
		{"info logged at debug level", "debug", "info", true},
		{"debug NOT logged at info level", "info", "debug", false},
		{"info logged at info level", "info", "info", true},
		{"warn logged at info level", "info", "warn", true},
		{"info NOT logged at warn level", "warn", "info", false},
		{"error logged at error level", "error", "error", true},
		{"warn NOT logged at error level", "error", "warn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cfg := Config{
				Level:       tt.configLevel,
				Format:      "json",
				ServiceName: "test",
			}

			log := NewWithOutput(cfg, &buf)

			switch tt.logLevel {
			case "debug":
				log.Debug().Msg("test")
			case "info":
				log.Info().Msg("test")
			case "warn":
				log.Warn().Msg("test")
			case "error":
				log.Error().Msg("test")
			}

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String(), "expected log output")
			} else {
				assert.Empty(t, buf.String(), "expected no log output")
			}
		})
	}
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "invalid",
		Format:      "json",
		ServiceName: "test",
	}

	// Should default to info level without panicking
	log := NewWithOutput(cfg, &buf)
	log.Info().Msg("test")

	assert.NotEmpty(t, buf.String())
}

func TestNewLogger_WithCaller(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:        "info",
		Format:       "json",
		ServiceName:  "test",
		EnableCaller: true,
	}

	log := NewWithOutput(cfg, &buf)
	log.Info().Msg("test")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	// Caller should be present
	assert.Contains(t, result, "caller")
	caller := result["caller"].(string)
	assert.Contains(t, caller, "logger_test.go")
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
	}

	log := NewWithOutput(cfg, &buf)
	logWithContext := log.WithContext("custom_field", "custom_value")
	logWithContext.Info().Msg("test")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "custom_value", result["custom_field"])
}

func TestLogger_WithRequestID(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
	}

	log := NewWithOutput(cfg, &buf)
	logWithReqID := log.WithRequestID("req-123")
	logWithReqID.Info().Msg("test")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "req-123", result["request_id"])
}

func TestLogger_WithProvider(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
	}

	log := NewWithOutput(cfg, &buf)
	logWithProvider := log.WithProvider("garuda")
	logWithProvider.Info().Msg("test")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "garuda", result["provider"])
}

func TestNop(t *testing.T) {
	var buf bytes.Buffer
	log := Nop()

	// Nop logger should not write anything
	log.Info().Msg("this should not appear")

	assert.Empty(t, buf.String())
}

func TestLogger_StructuredFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
	}

	log := NewWithOutput(cfg, &buf)
	log.Info().
		Str("origin", "CGK").
		Str("destination", "DPS").
		Int("passengers", 2).
		Float64("price", 1500000.50).
		Bool("direct", true).
		Msg("Flight search")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "CGK", result["origin"])
	assert.Equal(t, "DPS", result["destination"])
	assert.Equal(t, float64(2), result["passengers"])
	assert.Equal(t, 1500000.50, result["price"])
	assert.Equal(t, true, result["direct"])
	assert.Equal(t, "Flight search", result["message"])
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, false, cfg.EnableCaller)
	assert.Equal(t, "flight-search", cfg.ServiceName)
}

func TestGlobalLogger(t *testing.T) {
	// Reset global logger
	Global = nil

	var buf bytes.Buffer
	cfg := Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "global-test",
	}

	// Initialize with custom logger
	SetGlobal(NewWithOutput(cfg, &buf))

	Info().Msg("global info")

	output := buf.String()
	assert.Contains(t, output, "global info")
	assert.Contains(t, output, "global-test")
}

func TestGlobalLoggerAutoInit(t *testing.T) {
	// Reset global logger
	Global = nil

	// Calling global functions without Init should auto-initialize
	// This should not panic
	Info().Msg("auto-init test")

	// Global should now be set
	assert.NotNil(t, Global)
}
