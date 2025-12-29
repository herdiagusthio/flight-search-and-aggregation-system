// Package logger provides structured logging using zerolog.
// It supports JSON and console output formats with configurable log levels.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Config holds the logger configuration options.
type Config struct {
	// Level is the minimum log level (debug, info, warn, error, fatal, panic)
	Level string `env:"LOG_LEVEL" envDefault:"info"`

	// Format is the output format (json, console)
	Format string `env:"LOG_FORMAT" envDefault:"json"`

	// EnableCaller adds caller information to log entries
	EnableCaller bool `env:"LOG_CALLER" envDefault:"false"`

	// ServiceName is the name of the service for log context
	ServiceName string `env:"SERVICE_NAME" envDefault:"flight-search"`
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Level:        "info",
		Format:       "json",
		EnableCaller: false,
		ServiceName:  "flight-search",
	}
}

// Logger wraps zerolog.Logger with additional context.
type Logger struct {
	zerolog.Logger
}

// New creates a new Logger with the given configuration.
func New(cfg Config) *Logger {
	return NewWithOutput(cfg, os.Stdout)
}

// NewWithOutput creates a new Logger with custom output writer.
// This is useful for testing.
func NewWithOutput(cfg Config, output io.Writer) *Logger {
	// Parse and set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Configure output format
	var writer io.Writer = output
	if cfg.Format == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	// Build logger context
	ctx := zerolog.New(writer).
		Level(level).
		With().
		Timestamp().
		Str("service", cfg.ServiceName)

	// Add caller if enabled
	if cfg.EnableCaller {
		ctx = ctx.Caller()
	}

	return &Logger{
		Logger: ctx.Logger(),
	}
}

// WithContext returns a new logger with additional context fields.
func (l *Logger) WithContext(key, value string) *Logger {
	return &Logger{
		Logger: l.With().Str(key, value).Logger(),
	}
}

// WithRequestID returns a logger with request ID context.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithContext("request_id", requestID)
}

// WithProvider returns a logger with provider context.
func (l *Logger) WithProvider(provider string) *Logger {
	return l.WithContext("provider", provider)
}

// Nop returns a disabled logger that produces no output.
// Useful for testing when logs are not needed.
func Nop() *Logger {
	return &Logger{
		Logger: zerolog.Nop(),
	}
}

// Global is the global logger instance.
// It should be initialized at application startup.
var Global *Logger

// Init initializes the global logger with the given configuration.
func Init(cfg Config) {
	Global = New(cfg)
}

// SetGlobal sets a custom logger as the global logger.
func SetGlobal(l *Logger) {
	Global = l
}

// Info returns an info level event from the global logger.
func Info() *zerolog.Event {
	if Global == nil {
		Init(DefaultConfig())
	}
	return Global.Info()
}

// Error returns an error level event from the global logger.
func Error() *zerolog.Event {
	if Global == nil {
		Init(DefaultConfig())
	}
	return Global.Error()
}

// Debug returns a debug level event from the global logger.
func Debug() *zerolog.Event {
	if Global == nil {
		Init(DefaultConfig())
	}
	return Global.Debug()
}

// Warn returns a warn level event from the global logger.
func Warn() *zerolog.Event {
	if Global == nil {
		Init(DefaultConfig())
	}
	return Global.Warn()
}

// Fatal returns a fatal level event from the global logger.
func Fatal() *zerolog.Event {
	if Global == nil {
		Init(DefaultConfig())
	}
	return Global.Fatal()
}
