// Package config provides application configuration management.
// It loads configuration from environment variables with support for .env files.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Timeouts TimeoutConfig
	Logging  LoggingConfig
	App      AppConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s"`
}

// TimeoutConfig holds timeout settings for flight search operations.
type TimeoutConfig struct {
	GlobalSearch time.Duration `env:"TIMEOUT_GLOBAL_SEARCH" envDefault:"5s"`
	PerProvider  time.Duration `env:"TIMEOUT_PER_PROVIDER" envDefault:"2s"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env string `env:"APP_ENV" envDefault:"development"`
}

// Load reads configuration from environment variables.
// It attempts to load a .env file first (optional - won't fail if missing).
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found, using environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// MustLoad loads configuration or panics on error.
// Use this in main() where configuration is required to start.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// validate checks configuration values for correctness.
func validate(cfg *Config) error {
	// Validate server port
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	// Validate timeouts are positive
	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("SERVER_READ_TIMEOUT must be positive")
	}
	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("SERVER_WRITE_TIMEOUT must be positive")
	}
	if cfg.Timeouts.GlobalSearch <= 0 {
		return fmt.Errorf("TIMEOUT_GLOBAL_SEARCH must be positive")
	}
	if cfg.Timeouts.PerProvider <= 0 {
		return fmt.Errorf("TIMEOUT_PER_PROVIDER must be positive")
	}

	// Validate per-provider timeout is less than global timeout
	if cfg.Timeouts.PerProvider >= cfg.Timeouts.GlobalSearch {
		return fmt.Errorf("TIMEOUT_PER_PROVIDER (%s) should be less than TIMEOUT_GLOBAL_SEARCH (%s)",
			cfg.Timeouts.PerProvider, cfg.Timeouts.GlobalSearch)
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error; got %q", cfg.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("LOG_FORMAT must be one of: json, console; got %q", cfg.Logging.Format)
	}

	// Validate app environment
	validEnvs := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvs[cfg.App.Env] {
		return fmt.Errorf("APP_ENV must be one of: development, staging, production; got %q", cfg.App.Env)
	}

	return nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
