package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	var attempts int32

	err := Do(context.Background(), func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}, DefaultConfig)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), attempts)
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("temporary error")

	err := Do(context.Background(), func() error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return expectedErr
		}
		return nil
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(3), attempts)
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("persistent error")

	err := Do(context.Background(), func() error {
		atomic.AddInt32(&attempts, 1)
		return expectedErr
	}, Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, int32(3), attempts)
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var attempts int32

	// Cancel context after first attempt
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("temporary error")
	}, Config{
		MaxAttempts:  10,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	// Should have made at least 1 attempt
	assert.GreaterOrEqual(t, attempts, int32(1))
}

func TestDo_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var attempts int32

	err := Do(ctx, func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("temporary error")
	}, Config{
		MaxAttempts:  10,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestDo_RetryIfPredicate(t *testing.T) {
	var attempts int32
	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	// Should stop on non-retryable error
	err := Do(context.Background(), func() error {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			return retryableErr
		}
		return nonRetryableErr
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
		RetryIf: func(err error) bool {
			return err == retryableErr
		},
	})

	assert.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)
	assert.Equal(t, int32(2), attempts) // First retry + stop on non-retryable
}

func TestDo_ExponentialBackoff(t *testing.T) {
	var delays []time.Duration
	var attempts int32

	start := time.Now()
	err := Do(context.Background(), func() error {
		delays = append(delays, time.Since(start))
		count := atomic.AddInt32(&attempts, 1)
		if count < 4 {
			return errors.New("temporary")
		}
		return nil
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0, // No jitter for predictable timing
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(4), attempts)

	// Verify exponential backoff pattern
	// delays should be approximately: 10ms, 20ms, 40ms
	if len(delays) >= 4 {
		// First attempt is immediate
		assert.Less(t, delays[0], time.Duration(5*time.Millisecond))

		// Subsequent attempts should show increasing delays
		// (with some tolerance for execution time)
		assert.Greater(t, delays[1], time.Duration(8*time.Millisecond))
		assert.Greater(t, delays[2], time.Duration(25*time.Millisecond))
		assert.Greater(t, delays[3], time.Duration(55*time.Millisecond))
	}
}

func TestDo_MaxDelayRespected(t *testing.T) {
	var attempts int32
	start := time.Now()

	err := Do(context.Background(), func() error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("error")
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     60 * time.Millisecond, // Low max delay
		Multiplier:   10.0,                  // High multiplier
		JitterFactor: 0,
	})

	elapsed := time.Since(start)
	assert.Error(t, err)

	// With 4 delays capped at 60ms each, total should be around 240ms
	// (much less than if uncapped: 50 + 500 + 5000 + 50000 ms)
	assert.Less(t, elapsed, 400*time.Millisecond)
}

func TestDo_ZeroMaxAttempts(t *testing.T) {
	var attempts int32

	err := Do(context.Background(), func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}, Config{
		MaxAttempts: 0, // Should default to 1
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(1), attempts)
}

func TestDoWithResult_Success(t *testing.T) {
	var attempts int32

	result, err := DoWithResult(context.Background(), func() (string, error) {
		atomic.AddInt32(&attempts, 1)
		return "success", nil
	}, DefaultConfig)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, int32(1), attempts)
}

func TestDoWithResult_SuccessAfterRetries(t *testing.T) {
	var attempts int32

	result, err := DoWithResult(context.Background(), func() (int, error) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return 0, errors.New("temporary")
		}
		return 42, nil
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, int32(3), attempts)
}

func TestDoWithResult_MaxAttemptsExceeded(t *testing.T) {
	var attempts int32
	expectedErr := errors.New("persistent error")

	result, err := DoWithResult(context.Background(), func() (string, error) {
		atomic.AddInt32(&attempts, 1)
		return "partial", expectedErr
	}, Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "partial", result) // Last result is returned
	assert.Equal(t, int32(3), attempts)
}

func TestDoWithResult_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var attempts int32

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	result, err := DoWithResult(ctx, func() (string, error) {
		atomic.AddInt32(&attempts, 1)
		return "", errors.New("temporary")
	}, Config{
		MaxAttempts:  10,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, result)
}

func TestDoWithResult_WithStruct(t *testing.T) {
	type Flight struct {
		ID    string
		Price float64
	}

	var attempts int32

	result, err := DoWithResult(context.Background(), func() (Flight, error) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			return Flight{}, errors.New("temporary")
		}
		return Flight{ID: "GA-123", Price: 1500000}, nil
	}, Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	})

	assert.NoError(t, err)
	assert.Equal(t, "GA-123", result.ID)
	assert.Equal(t, float64(1500000), result.Price)
}

func TestPermanentError(t *testing.T) {
	originalErr := errors.New("validation failed")
	permanent := NewPermanent(originalErr)

	assert.True(t, IsPermanent(permanent))
	assert.Equal(t, "validation failed", permanent.Error())

	// Unwrap should return original error
	var pErr *Permanent
	assert.True(t, errors.As(permanent, &pErr))
	assert.Equal(t, originalErr, pErr.Unwrap())
}

func TestPermanentError_Nil(t *testing.T) {
	permanent := NewPermanent(nil)
	assert.Nil(t, permanent)
}

func TestIsPermanent(t *testing.T) {
	assert.True(t, IsPermanent(NewPermanent(errors.New("test"))))
	assert.False(t, IsPermanent(errors.New("regular error")))
	assert.False(t, IsPermanent(nil))
}

func TestSkipPermanent(t *testing.T) {
	regular := errors.New("regular")
	permanent := NewPermanent(errors.New("permanent"))

	assert.True(t, SkipPermanent(regular))
	assert.False(t, SkipPermanent(permanent))
}

func TestDo_WithSkipPermanent(t *testing.T) {
	var attempts int32

	err := Do(context.Background(), func() error {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			return errors.New("retryable")
		}
		return NewPermanent(errors.New("permanent"))
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
		RetryIf:      SkipPermanent,
	})

	assert.Error(t, err)
	assert.True(t, IsPermanent(err))
	assert.Equal(t, int32(2), attempts) // 1 retryable + 1 permanent
}

func TestConfig_Builders(t *testing.T) {
	cfg := DefaultConfig.
		WithMaxAttempts(5).
		WithInitialDelay(200 * time.Millisecond).
		WithMaxDelay(5 * time.Second).
		WithRetryIf(SkipPermanent)

	assert.Equal(t, 5, cfg.MaxAttempts)
	assert.Equal(t, 200*time.Millisecond, cfg.InitialDelay)
	assert.Equal(t, 5*time.Second, cfg.MaxDelay)
	assert.NotNil(t, cfg.RetryIf)
}

func TestDefaultConfig(t *testing.T) {
	assert.Equal(t, 3, DefaultConfig.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, DefaultConfig.InitialDelay)
	assert.Equal(t, 2*time.Second, DefaultConfig.MaxDelay)
	assert.Equal(t, 2.0, DefaultConfig.Multiplier)
	assert.Equal(t, 0.1, DefaultConfig.JitterFactor)
}

func TestProviderConfig(t *testing.T) {
	assert.Equal(t, 3, ProviderConfig.MaxAttempts)
	assert.Equal(t, 200*time.Millisecond, ProviderConfig.InitialDelay)
	assert.Equal(t, 5*time.Second, ProviderConfig.MaxDelay)
	assert.Equal(t, 0.2, ProviderConfig.JitterFactor)
}

func TestAggressiveConfig(t *testing.T) {
	assert.Equal(t, 5, AggressiveConfig.MaxAttempts)
	assert.Equal(t, 50*time.Millisecond, AggressiveConfig.InitialDelay)
	assert.Equal(t, 1*time.Second, AggressiveConfig.MaxDelay)
}

func TestPermanent_ErrorWithNil(t *testing.T) {
	permanent := &Permanent{Err: nil}
	assert.Equal(t, "permanent error", permanent.Error())
}

func TestDoWithResult_RetryIfPredicate(t *testing.T) {
	var attempts int32
	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	result, err := DoWithResult(context.Background(), func() (int, error) {
		count := atomic.AddInt32(&attempts, 1)
		if count == 1 {
			return 0, retryableErr
		}
		return 99, nonRetryableErr
	}, Config{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
		RetryIf: func(err error) bool {
			return err == retryableErr
		},
	})

	assert.Error(t, err)
	assert.Equal(t, nonRetryableErr, err)
	assert.Equal(t, 99, result)
	assert.Equal(t, int32(2), attempts)
}

func TestDo_ContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var attempts int32

	err := Do(ctx, func() error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}, DefaultConfig)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Equal(t, int32(0), attempts) // No attempts made
}
