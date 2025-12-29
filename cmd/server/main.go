// Package main is the entry point for the flight search aggregation service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flight-search/flight-search-and-aggregation-system/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	shutdownTimeout = 10 * time.Second
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	// Initialize logger with config
	setupLogger(cfg)

	log.Info().
		Str("env", cfg.App.Env).
		Int("port", cfg.Server.Port).
		Msg("Configuration loaded")

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Configure server timeouts from config
	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout

	// Setup middleware
	setupMiddleware(e)

	// Setup routes
	setupRoutes(e)

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	go func() {
		log.Info().Str("address", addr).Msg("Starting server")
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	gracefulShutdown(e)
}

// setupLogger configures the global zerolog logger based on config.
func setupLogger(cfg *config.Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Use console writer for non-JSON format
	if cfg.Logging.Format != "json" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	// Set log level from config
	switch cfg.Logging.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}


// setupMiddleware configures Echo middleware stack.
func setupMiddleware(e *echo.Echo) {
	// Recovery middleware - recover from panics
	e.Use(middleware.Recover())

	// Request ID middleware
	e.Use(middleware.RequestID())

	// Logger middleware with zerolog integration
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().
				Str("request_id", v.RequestID).
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Msg("HTTP request")
			return nil
		},
	}))
}

// setupRoutes configures the HTTP routes.
func setupRoutes(e *echo.Echo) {
	// Health check endpoint
	e.GET("/health", healthCheckHandler)
}

// healthCheckHandler returns the health status of the service.
func healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// gracefulShutdown handles graceful server shutdown on interrupt signals.
func gracefulShutdown(e *echo.Echo) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
	}

	log.Info().Msg("Server stopped")
}
