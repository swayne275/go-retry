package repeat

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/common/backoff"
)

var errFunctionSignaledToStop = fmt.Errorf("function signaled to stop")
var errBackoffSignaledToStop = fmt.Errorf("backoff signaled to stop")

// RepeatFunc is a function passed to retry.
// It returns true if the function should be repeated, false otherwise.
type RepeatFunc func(ctx context.Context) bool

// Do wraps a function with a backoff to repeat as long as f returns true, or until
// the backoff signals to stop.
// The provided context is passed to the RepeatFunc.
func Do(ctx context.Context, b backoff.Backoff, f RepeatFunc) error {
	for {
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !f(ctx) {
			return errFunctionSignaledToStop
		}

		next, stop := b.Next()
		if stop {
			return errBackoffSignaledToStop
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

// RepeatFunc is a function passed to retry.
// It returns true if the function should be repeated, false otherwise.
type RepeatUntilErrorFunc func(ctx context.Context) error

// DoUntilError wraps a function with a backoff to repeat until f returns an error, or
// until the backoff signals to stop.
// The provided context is passed to the RepeatFunc.
func DoUntilError(ctx context.Context, b backoff.Backoff, f RepeatUntilErrorFunc) error {
	for {
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := f(ctx); err != nil {
			return fmt.Errorf("%w: %w", errFunctionSignaledToStop, err)
		}

		next, stop := b.Next()
		if stop {
			return errBackoffSignaledToStop
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
