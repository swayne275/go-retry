package retry

import (
	"context"
	"fmt"
	"time"
)

// Constant is a wrapper around Retry that uses a constant backoff. It will
// retry the function f until it returns an error, or the context is canceled.
// TODO is this really useful vs an example? would have to extend with a repeat version too.
func Constant(ctx context.Context, t time.Duration, f RetryFunc) error {
	b, err := NewConstant(t)
	if err != nil {
		return fmt.Errorf("failed to create constant backoff: %w", err)
	}

	return Do(ctx, b, f)
}

// NewConstant creates a new constant backoff using the value t. The wait time
// is the provided constant value. It returns an error if t is not greater than 0.
func NewConstant(t time.Duration) (Backoff, error) {
	if t <= 0 {
		return nil, fmt.Errorf("constant backoff must be greater than zero")
	}

	return BackoffFunc(func() (time.Duration, bool) {
		return t, false
	}), nil
}
