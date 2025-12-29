// Package retry provides a generic retry mechanism with exponential backoff.
package retry

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// Config holds the retry configuration options.
type Config struct {
	// MaxAttempts is the maximum number of retry attempts (including the initial attempt).
	MaxAttempts int

	// InitialDelay is the delay before the first retry.
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64

	// JitterFactor is the factor for random jitter (0.0 to 1.0).
	// A value of 0.1 means up to 10% jitter will be added.
	JitterFactor float64

	// RetryIf is an optional predicate to determine if an error is retryable.
	// If nil, all errors are considered retryable.
	RetryIf func(error) bool
}

// DefaultConfig provides sensible defaults for retry behavior.
var DefaultConfig = Config{
	MaxAttempts:  3,
	InitialDelay: 100 * time.Millisecond,
	MaxDelay:     2 * time.Second,
	Multiplier:   2.0,
	JitterFactor: 0.1,
	RetryIf:      nil, // Retry all errors
}

// ProviderConfig is optimized for external API calls.
var ProviderConfig = Config{
	MaxAttempts:  3,
	InitialDelay: 200 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
	JitterFactor: 0.2,
	RetryIf:      nil,
}

// AggressiveConfig retries more times with shorter delays.
var AggressiveConfig = Config{
	MaxAttempts:  5,
	InitialDelay: 50 * time.Millisecond,
	MaxDelay:     1 * time.Second,
	Multiplier:   1.5,
	JitterFactor: 0.1,
	RetryIf:      nil,
}

// Do executes the function with retry logic.
// It returns nil if the function succeeds, or the last error if all attempts fail.
func Do(ctx context.Context, fn func() error, cfg Config) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return lastErr
		}

		// Don't sleep after last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate sleep time with jitter
		sleepTime := calculateSleepTime(delay, cfg.MaxDelay, cfg.JitterFactor)

		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepTime):
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * cfg.Multiplier)
	}

	return lastErr
}

// DoWithResult executes a function that returns a value with retry logic.
// This is the generic version for functions that return a result.
func DoWithResult[T any](ctx context.Context, fn func() (T, error), cfg Config) (T, error) {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	var result T
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return result, err
		}

		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Check if error is retryable
		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return result, lastErr
		}

		// Don't sleep after last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate sleep time with jitter
		sleepTime := calculateSleepTime(delay, cfg.MaxDelay, cfg.JitterFactor)

		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(sleepTime):
		}

		// Exponential backoff
		delay = time.Duration(float64(delay) * cfg.Multiplier)
	}

	return result, lastErr
}

// calculateSleepTime computes the sleep duration with jitter and max cap.
func calculateSleepTime(delay, maxDelay time.Duration, jitterFactor float64) time.Duration {
	// Add jitter
	jitter := time.Duration(rand.Float64() * float64(delay) * jitterFactor)
	sleepTime := delay + jitter

	// Cap at max delay
	if sleepTime > maxDelay {
		sleepTime = maxDelay
	}

	return sleepTime
}

// Permanent wraps an error to indicate it should not be retried.
type Permanent struct {
	Err error
}

func (p *Permanent) Error() string {
	if p.Err == nil {
		return "permanent error"
	}
	return p.Err.Error()
}

func (p *Permanent) Unwrap() error {
	return p.Err
}

// NewPermanent creates a permanent (non-retryable) error.
func NewPermanent(err error) error {
	if err == nil {
		return nil
	}
	return &Permanent{Err: err}
}

// IsPermanent checks if an error is permanent (non-retryable).
func IsPermanent(err error) bool {
	var permanent *Permanent
	return errors.As(err, &permanent)
}

// SkipPermanent is a RetryIf predicate that skips permanent errors.
func SkipPermanent(err error) bool {
	return !IsPermanent(err)
}

// WithRetryIf returns a new config with the given RetryIf predicate.
func (c Config) WithRetryIf(fn func(error) bool) Config {
	c.RetryIf = fn
	return c
}

// WithMaxAttempts returns a new config with the given max attempts.
func (c Config) WithMaxAttempts(n int) Config {
	c.MaxAttempts = n
	return c
}

// WithInitialDelay returns a new config with the given initial delay.
func (c Config) WithInitialDelay(d time.Duration) Config {
	c.InitialDelay = d
	return c
}

// WithMaxDelay returns a new config with the given max delay.
func (c Config) WithMaxDelay(d time.Duration) Config {
	c.MaxDelay = d
	return c
}
