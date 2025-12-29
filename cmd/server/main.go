// Package main is the entry point for the flight search aggregation service.
//
//	@title						Flight Search Aggregation API
//	@version					1.0.0
//	@description				A high-performance flight search aggregation service that queries multiple airline providers and returns unified results.
//
//	@contact.name				API Support
//	@contact.url				https://github.com/flight-search/flight-search-and-aggregation-system/issues
//
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@schemes					http https
//
//	@externalDocs.description	Technical Documentation
//	@externalDocs.url			https://github.com/flight-search/flight-search-and-aggregation-system/blob/main/docs/architecture.md
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
	echoSwagger "github.com/swaggo/echo-swagger"

	// Import generated docs for swagger
	_ "github.com/flight-search/flight-search-and-aggregation-system/docs"

	// Application layers
	flighthttp "github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/http"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/provider/airasia"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/provider/batikair"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/provider/garuda"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/adapter/provider/lionair"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/domain"
	"github.com/flight-search/flight-search-and-aggregation-system/internal/usecase"
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
	setupRoutes(e, cfg)

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
func setupRoutes(e *echo.Echo, cfg *config.Config) {
	// Health check endpoint (root level for load balancers)
	e.GET("/health", healthCheckHandler)

	// Initialize providers with mock data paths
	mockBasePath := "docs/response-mock"
	providers := []domain.FlightProvider{
		garuda.NewAdapter(mockBasePath + "/garuda_indonesia_search_response.json"),
		lionair.NewAdapter(mockBasePath + "/lion_air_search_response.json"),
		batikair.NewAdapter(mockBasePath + "/batik_air_search_response.json"),
		airasia.NewAdapter(mockBasePath + "/airasia_search_response.json"),
	}

	// Initialize use case with config
	ucConfig := &usecase.Config{
		GlobalTimeout:   cfg.Timeouts.GlobalSearch,
		ProviderTimeout: cfg.Timeouts.PerProvider,
	}
	flightUseCase := usecase.NewFlightSearchUseCase(providers, ucConfig)

	// Initialize handler
	flightHandler := flighthttp.NewFlightHandler(flightUseCase)

	// API v1 routes
	api := e.Group("/api/v1")
	api.POST("/flights/search", flightHandler.SearchFlights)

	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}

// healthCheckHandler returns the health status of the service.
// Note: This endpoint is at the root level (/health), not under /api/v1
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
