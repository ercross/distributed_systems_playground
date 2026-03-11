package retry

import (
	"context"
	"math/rand"
	"time"
)

type BackoffStrategy string

const (
	BackoffNone        BackoffStrategy = "none"
	BackoffFixed       BackoffStrategy = "fixed"
	BackoffExponential BackoffStrategy = "exponential"
)

type Config struct {
	// MaxAttempts is the total number of tries, including the first attempt.
	// Example: MaxAttempts=3 means 1 initial try + 2 retries.
	MaxAttempts int

	// BaseDelay is the starting delay before retries.
	BaseDelay time.Duration

	// MaxDelay caps the computed delay.
	MaxDelay time.Duration

	// Backoff controls how delay grows between attempts.
	Backoff BackoffStrategy

	// Multiplier is used by exponential backoff.
	// Example: delay = BaseDelay * Multiplier^(attempt-1)
	Multiplier float64

	// Jitter enables randomization to avoid synchronized retries.
	Jitter bool

	// JitterFraction is applied to the computed delay.
	// Example: 0.2 means +/-20% randomization.
	JitterFraction float64

	// RetryIf determines whether an error should be retried.
	// If nil, a sensible default is used.
	RetryIf func(error) bool

	// OnRetry is called before sleeping for the next retry.
	OnRetry func(attempt int, err error, nextDelay time.Duration)

	// ClockSleep allows testing with a fake sleeper.
	ClockSleep func(ctx context.Context, d time.Duration) error

	// RandSource allows deterministic testing.
	RandSource *rand.Rand

	Hooks Hooks
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts:    3,
		BaseDelay:      100 * time.Millisecond,
		MaxDelay:       2 * time.Second,
		Backoff:        BackoffExponential,
		Multiplier:     2.0,
		Jitter:         true,
		JitterFraction: 0.2,
		RetryIf:        DefaultRetryIf,
		ClockSleep:     sleepContext,
		RandSource:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func ChaosConfig() Config {
	return Config{
		MaxAttempts:    5,
		BaseDelay:      50 * time.Millisecond,
		MaxDelay:       1500 * time.Millisecond,
		Backoff:        BackoffExponential,
		Multiplier:     2.0,
		Jitter:         true,
		JitterFraction: 0.35,
		RetryIf:        DefaultRetryIf,
		ClockSleep:     sleepContext,
		RandSource:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

type Hooks struct {
	OnAttempt func(attempt int)
	OnSuccess func(attempt int)
	OnFailure func(attempt int, err error)
}
