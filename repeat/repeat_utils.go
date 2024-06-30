package repeat

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

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
