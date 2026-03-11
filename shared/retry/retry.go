package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"
)

func DefaultRetryIf(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Retry common temporary network failures.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Fallback: retry a few common transport-ish conditions by message.
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "connection refused"),
		strings.Contains(msg, "broken pipe"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "temporarily unavailable"),
		strings.Contains(msg, "no such host"),
		strings.Contains(msg, "server misbehaving"),
		strings.Contains(msg, "tls handshake timeout"),
		strings.Contains(msg, "unexpected eof"):
		return true
	default:
		return false
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeConfig(cfg Config) Config {
	def := DefaultConfig()

	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = def.MaxAttempts
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = def.BaseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = def.MaxDelay
	}
	if cfg.Backoff == "" {
		cfg.Backoff = def.Backoff
	}
	if cfg.Multiplier <= 1 {
		cfg.Multiplier = def.Multiplier
	}
	if cfg.JitterFraction < 0 {
		cfg.JitterFraction = def.JitterFraction
	}
	if cfg.RetryIf == nil {
		cfg.RetryIf = def.RetryIf
	}
	if cfg.ClockSleep == nil {
		cfg.ClockSleep = def.ClockSleep
	}
	if cfg.RandSource == nil {
		cfg.RandSource = def.RandSource
	}

	return cfg
}

func computeDelay(cfg Config, attempt int) time.Duration {
	// attempt is 1-based retry number, not overall try number.
	// attempt=1 means first retry after the initial failure.
	var delay time.Duration

	switch cfg.Backoff {
	case BackoffNone:
		delay = 0
	case BackoffFixed:
		delay = cfg.BaseDelay
	case BackoffExponential:
		f := float64(cfg.BaseDelay) * math.Pow(cfg.Multiplier, float64(attempt-1))
		delay = time.Duration(f)
	default:
		delay = cfg.BaseDelay
	}

	if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	if cfg.Jitter && delay > 0 && cfg.JitterFraction > 0 {
		// Randomize in range [delay*(1-j), delay*(1+j)]
		j := cfg.JitterFraction
		minF := 1 - j
		maxF := 1 + j
		factor := minF + cfg.RandSource.Float64()*(maxF-minF)
		delay = time.Duration(float64(delay) * factor)
	}

	if delay < 0 {
		return 0
	}
	return delay
}

type AttemptError struct {
	Attempt int
	Err     error
}

func (e AttemptError) Error() string {
	return fmt.Sprintf("attempt %d failed: %v", e.Attempt, e.Err)
}

func (e AttemptError) Unwrap() error {
	return e.Err
}

func Do(ctx context.Context, cfg Config, fn func(context.Context) error) error {
	cfg = normalizeConfig(cfg)

	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if cfg.Hooks.OnAttempt != nil {
			cfg.Hooks.OnAttempt(attempt)
		}

		err := fn(ctx)
		if err == nil {
			if cfg.Hooks.OnSuccess != nil {
				cfg.Hooks.OnSuccess(attempt)
			}
			return nil
		}

		lastErr = AttemptError{Attempt: attempt, Err: err}

		if cfg.Hooks.OnFailure != nil {
			cfg.Hooks.OnFailure(attempt, lastErr)
		}

		// No more attempts left.
		if attempt == cfg.MaxAttempts {
			break
		}

		// Failure is not retryable.
		if !cfg.RetryIf(err) {
			return lastErr
		}

		delay := computeDelay(cfg, attempt)
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt, err, delay)
		}

		if err := cfg.ClockSleep(ctx, delay); err != nil {
			return err
		}
	}

	return lastErr
}
