package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad_Defaults tests that all default values load correctly without any env vars.
func TestLoad_Defaults(t *testing.T) {
	// Clear all config-related env vars
	clearEnvVars(t)

	cfg, err := Load()
	require.NoError(t, err)

	// Server defaults
	assert.Equal(t, 8080, cfg.Server.Port, "default server port")
	assert.Equal(t, "10s", cfg.Server.ReadTimeout.String(), "default read timeout")
	assert.Equal(t, "10s", cfg.Server.WriteTimeout.String(), "default write timeout")

	// Timeout defaults
	assert.Equal(t, "5s", cfg.Timeouts.GlobalSearch.String(), "default global search timeout")
	assert.Equal(t, "2s", cfg.Timeouts.PerProvider.String(), "default per-provider timeout")

	// Logging defaults
	assert.Equal(t, "info", cfg.Logging.Level, "default log level")
	assert.Equal(t, "json", cfg.Logging.Format, "default log format")

	// App defaults
	assert.Equal(t, "development", cfg.App.Env, "default app environment")
}

// TestLoad_EnvironmentOverrides tests that environment variables override defaults.
func TestLoad_EnvironmentOverrides(t *testing.T) {
	clearEnvVars(t)

	// Set custom values
	setEnvVars(t, map[string]string{
		"SERVER_PORT":           "3000",
		"SERVER_READ_TIMEOUT":   "30s",
		"SERVER_WRITE_TIMEOUT":  "30s",
		"TIMEOUT_GLOBAL_SEARCH": "10s",
		"TIMEOUT_PER_PROVIDER":  "3s",
		"LOG_LEVEL":             "debug",
		"LOG_FORMAT":            "console",
		"APP_ENV":               "production",
	})

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 3000, cfg.Server.Port)
	assert.Equal(t, "30s", cfg.Server.ReadTimeout.String())
	assert.Equal(t, "30s", cfg.Server.WriteTimeout.String())
	assert.Equal(t, "10s", cfg.Timeouts.GlobalSearch.String())
	assert.Equal(t, "3s", cfg.Timeouts.PerProvider.String())
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "console", cfg.Logging.Format)
	assert.Equal(t, "production", cfg.App.Env)
}

// TestLoad_PartialOverrides tests that only overridden values change.
func TestLoad_PartialOverrides(t *testing.T) {
	clearEnvVars(t)

	// Only override port
	setEnvVars(t, map[string]string{
		"SERVER_PORT": "9000",
	})

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 9000, cfg.Server.Port, "overridden port")
	assert.Equal(t, "10s", cfg.Server.ReadTimeout.String(), "default read timeout")
	assert.Equal(t, "info", cfg.Logging.Level, "default log level")
}

// TestLoad_Validation_PortRange tests port validation boundaries.
func TestLoad_Validation_PortRange(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
		errMsg  string
	}{
		{"valid port 1", "1", false, ""},
		{"valid port 80", "80", false, ""},
		{"valid port 8080", "8080", false, ""},
		{"valid port 65535", "65535", false, ""},
		{"invalid port 0", "0", true, "SERVER_PORT must be between 1 and 65535"},
		{"invalid port negative", "-1", true, "SERVER_PORT must be between 1 and 65535"},
		{"invalid port too high", "65536", true, "SERVER_PORT must be between 1 and 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"SERVER_PORT": tt.port})

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoad_Validation_PositiveTimeouts tests that timeouts must be positive.
func TestLoad_Validation_PositiveTimeouts(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		value  string
		errMsg string
	}{
		{"zero read timeout", "SERVER_READ_TIMEOUT", "0s", "SERVER_READ_TIMEOUT must be positive"},
		{"negative read timeout", "SERVER_READ_TIMEOUT", "-1s", "SERVER_READ_TIMEOUT must be positive"},
		{"zero write timeout", "SERVER_WRITE_TIMEOUT", "0s", "SERVER_WRITE_TIMEOUT must be positive"},
		{"negative write timeout", "SERVER_WRITE_TIMEOUT", "-1s", "SERVER_WRITE_TIMEOUT must be positive"},
		{"zero global search timeout", "TIMEOUT_GLOBAL_SEARCH", "0s", "TIMEOUT_GLOBAL_SEARCH must be positive"},
		{"negative global search timeout", "TIMEOUT_GLOBAL_SEARCH", "-1s", "TIMEOUT_GLOBAL_SEARCH must be positive"},
		{"zero per-provider timeout", "TIMEOUT_PER_PROVIDER", "0s", "TIMEOUT_PER_PROVIDER must be positive"},
		{"negative per-provider timeout", "TIMEOUT_PER_PROVIDER", "-1s", "TIMEOUT_PER_PROVIDER must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{tt.envVar: tt.value})

			cfg, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Nil(t, cfg)
		})
	}
}

// TestLoad_Validation_PerProviderLessThanGlobal tests that per-provider timeout must be less than global.
func TestLoad_Validation_PerProviderLessThanGlobal(t *testing.T) {
	clearEnvVars(t)

	// Set per-provider equal to global
	setEnvVars(t, map[string]string{
		"TIMEOUT_GLOBAL_SEARCH": "5s",
		"TIMEOUT_PER_PROVIDER":  "5s",
	})

	cfg, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TIMEOUT_PER_PROVIDER")
	assert.Contains(t, err.Error(), "should be less than")
	assert.Nil(t, cfg)

	// Set per-provider greater than global
	setEnvVars(t, map[string]string{
		"TIMEOUT_GLOBAL_SEARCH": "5s",
		"TIMEOUT_PER_PROVIDER":  "10s",
	})

	cfg, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TIMEOUT_PER_PROVIDER")
	assert.Contains(t, err.Error(), "should be less than")
	assert.Nil(t, cfg)
}

// TestLoad_Validation_LogLevel tests log level validation.
func TestLoad_Validation_LogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{"valid debug", "debug", false},
		{"valid info", "info", false},
		{"valid warn", "warn", false},
		{"valid error", "error", false},
		{"invalid trace", "trace", true},
		{"invalid fatal", "fatal", true},
		// Note: empty string uses default value "info" due to envDefault tag
		{"invalid random", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"LOG_LEVEL": tt.level})

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "LOG_LEVEL must be one of")
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoad_Validation_LogFormat tests log format validation.
func TestLoad_Validation_LogFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"valid json", "json", false},
		{"valid console", "console", false},
		{"invalid text", "text", true},
		// Note: empty string uses default value "json" due to envDefault tag
		{"invalid random", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"LOG_FORMAT": tt.format})

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "LOG_FORMAT must be one of")
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoad_Validation_AppEnv tests app environment validation.
func TestLoad_Validation_AppEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{"valid development", "development", false},
		{"valid staging", "staging", false},
		{"valid production", "production", false},
		{"invalid local", "local", true},
		// Note: empty string uses default value "development" due to envDefault tag
		{"invalid random", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"APP_ENV": tt.env})

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "APP_ENV must be one of")
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

// TestLoad_DurationParsing tests that duration strings are parsed correctly.
func TestLoad_DurationParsing(t *testing.T) {
	clearEnvVars(t)

	setEnvVars(t, map[string]string{
		"SERVER_READ_TIMEOUT":   "1m30s",
		"SERVER_WRITE_TIMEOUT":  "2m",
		"TIMEOUT_GLOBAL_SEARCH": "500ms",
		"TIMEOUT_PER_PROVIDER":  "100ms",
	})

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "1m30s", cfg.Server.ReadTimeout.String())
	assert.Equal(t, "2m0s", cfg.Server.WriteTimeout.String())
	assert.Equal(t, "500ms", cfg.Timeouts.GlobalSearch.String())
	assert.Equal(t, "100ms", cfg.Timeouts.PerProvider.String())
}

// TestMustLoad_Success tests MustLoad with valid config.
func TestMustLoad_Success(t *testing.T) {
	clearEnvVars(t)

	assert.NotPanics(t, func() {
		cfg := MustLoad()
		assert.NotNil(t, cfg)
	})
}

// TestMustLoad_Panic tests MustLoad panics on invalid config.
func TestMustLoad_Panic(t *testing.T) {
	clearEnvVars(t)
	setEnvVars(t, map[string]string{"SERVER_PORT": "0"})

	assert.Panics(t, func() {
		MustLoad()
	})
}

// TestConfig_IsDevelopment tests the IsDevelopment helper method.
func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"development", true},
		{"staging", false},
		{"production", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"APP_ENV": tt.env})

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.IsDevelopment())
		})
	}
}

// TestConfig_IsProduction tests the IsProduction helper method.
func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"development", false},
		{"staging", false},
		{"production", true},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			clearEnvVars(t)
			setEnvVars(t, map[string]string{"APP_ENV": tt.env})

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.IsProduction())
		})
	}
}

// Helper functions

// clearEnvVars clears all config-related environment variables.
func clearEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"SERVER_PORT",
		"SERVER_READ_TIMEOUT",
		"SERVER_WRITE_TIMEOUT",
		"TIMEOUT_GLOBAL_SEARCH",
		"TIMEOUT_PER_PROVIDER",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"APP_ENV",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

// setEnvVars sets multiple environment variables.
func setEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		os.Setenv(k, v)
	}
}
