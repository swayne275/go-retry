package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

// TODO tests should include retryable errors and non retryable errors

// ConstantRetry is a wrapper around retry that uses a constant backoff. It will
// retry the function f until it returns a non-retryable error, or the context is canceled.
func ConstantRetry(ctx context.Context, t time.Duration, f RetryFunc) error {
	b, err := backoff.NewConstant(t)
	if err != nil {
		return fmt.Errorf("failed to create constant backoff: %w", err)
	}

	return Do(ctx, b, f)
}

// ExponentialRetry is a wrapper around retry that uses an exponential backoff. It will
// retry the function f until it returns a non-retryable error, or the context is canceled.
func ExponentialRetry(ctx context.Context, base time.Duration, f RetryFunc) error {
	b, err := backoff.NewExponential(base)
	if err != nil {
		return fmt.Errorf("failed to create exponential backoff: %w", err)
	}

	return Do(ctx, b, f)
}

// FibonacciRetry is a wrapper around retry that uses a FibonacciRetry backoff. It will
// retry the function f until it returns a non-retryable error, or the context is canceled.
func FibonacciRetry(ctx context.Context, base time.Duration, f RetryFunc) error {
	b, err := backoff.NewFibonacci(base)
	if err != nil {
		return fmt.Errorf("failed to create fibonacci backoff: %w", err)

	}
	return Do(ctx, b, f)
}
