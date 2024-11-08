// Package retry provides helpers for retrying.
//
// This package defines flexible interfaces for retrying Go functions that may
// be flakey or eventually consistent. It abstracts the "backoff" (how long to
// wait between tries) and "retry" (execute the function again) mechanisms for
// maximum flexibility. Furthermore, everything is an interface, so you can
// define your own implementations.
//
// The package is modeled after Go's built-in HTTP package, making it easy to
// customize the built-in backoff with your own custom logic. Additionally,
// callers specify which errors are retryable by wrapping them. This is helpful
// with complex operations where only certain results should retry.
package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

var ErrNonRetryable = fmt.Errorf("function returned non retryable error")
var errBackoffSignaledToStop = fmt.Errorf("backoff signaled to stop")

// RetryFunc is a function passed to retry.
type RetryFunc func(ctx context.Context) error

type retryableError struct {
	err error
}

// RetryableError marks an error as retryable.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err}
}

// Unwrap implements error wrapping.
func (e *retryableError) Unwrap() error {
	return e.err
}

// Error returns the error string.
func (e *retryableError) Error() string {
	if e.err == nil {
		return "retryable: <nil>"
	}
	return "retryable: " + e.err.Error()
}

// Do wraps a function with a backoff to retry. It will retry until f returns either
// nil or a non-retryable error.
// The provided context is the same context passed to the RetryFunc.
func Do(ctx context.Context, b backoff.Backoff, f RetryFunc) error {
	for {
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := f(ctx)
		if err == nil {
			return nil
		}

		// Not retryable
		var rerr *retryableError
		if !errors.As(err, &rerr) {
			return fmt.Errorf("%w: %w", ErrNonRetryable, err)
		}

		next, stop := b.Next()
		if stop {
			return fmt.Errorf("%w: %w", errBackoffSignaledToStop, rerr.Unwrap())
		}

		// ctx.Done() has priority, so we test it alone first
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		t := time.NewTimer(next)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
			continue
		}
	}
}

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
