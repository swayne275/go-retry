package backoff

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry/common/backoff"
	"github.com/swayne275/go-retry/retry"
)

// ConstantRetry is a wrapper around retry that uses a constant backoff. It will
// retry the function f until it returns an error, or the context is canceled.
func ConstantRetry(ctx context.Context, t time.Duration, f retry.RetryFunc) error {
	b, err := NewConstant(t)
	if err != nil {
		return fmt.Errorf("failed to create constant backoff: %w", err)
	}

	return retry.Do(ctx, b, f)
}

// NewConstant creates a new constant backoff using the value t. The wait time
// is the provided constant value. It returns an error if t is not greater than 0.
func NewConstant(t time.Duration) (backoff.Backoff, error) {
	if t <= 0 {
		return nil, fmt.Errorf("constant backoff must be greater than zero")
	}

	return BackoffFunc(func() (time.Duration, bool) {
		return t, false
	}), nil
}
