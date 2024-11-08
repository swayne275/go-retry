package repeat

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

var ErrFunctionSignaledToStop = fmt.Errorf("function signaled to stop")
var ErrBackoffSignaledToStop = fmt.Errorf("backoff signaled to stop")

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
			return ErrFunctionSignaledToStop
		}

		next, stop := b.Next()
		if stop {
			return ErrBackoffSignaledToStop
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

// RepeatUntilErrorFunc is a function passed to retry.
// It returns an error if the function should be stopped, nil otherwise.
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
			return fmt.Errorf("%w: %w", ErrFunctionSignaledToStop, err)
		}

		next, stop := b.Next()
		if stop {
			return ErrBackoffSignaledToStop
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

// ConstantRepeat is a wrapper around repeat that uses a constant backoff. It will
// repeat the function f until it returns false, or the context is canceled.
func ConstantRepeat(ctx context.Context, t time.Duration, f RepeatFunc) error {
	b, err := backoff.NewConstant(t)
	if err != nil {
		return fmt.Errorf("failed to create constant backoff: %w", err)
	}

	return Do(ctx, b, f)
}

// ExponentialRetry is a wrapper around repeat that uses an exponential backoff. It will
// repeat the function f until it returns false, or the context is canceled.
func ExponentialRepeat(ctx context.Context, base time.Duration, f RepeatFunc) error {
	b, err := backoff.NewExponential(base)
	if err != nil {
		return fmt.Errorf("failed to create exponential backoff: %w", err)
	}

	return Do(ctx, b, f)
}

// FibonacciRepeat is a wrapper around repeat that uses a FibonacciRetry backoff. It will
// repeat the function f until it returns false, or the context is canceled.
func FibonacciRepeat(ctx context.Context, base time.Duration, f RepeatFunc) error {
	b, err := backoff.NewFibonacci(base)
	if err != nil {
		return fmt.Errorf("failed to create fibonacci backoff: %w", err)

	}
	return Do(ctx, b, f)
}
