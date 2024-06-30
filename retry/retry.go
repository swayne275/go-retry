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

	"github.com/swayne275/go-retry/common/backoff"
)

var errFunctionReturnedNonRetryableError = fmt.Errorf("function returned non retryable error")
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
			return fmt.Errorf("%w: %w", errFunctionReturnedNonRetryableError, err)
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
