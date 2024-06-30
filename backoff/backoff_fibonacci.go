package backoff

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/swayne275/go-retry/common/backoff"
	"github.com/swayne275/go-retry/retry"
)

type state [2]time.Duration

type fibonacciBackoff struct {
	state unsafe.Pointer
	base  time.Duration
}

// FibonacciRetry is a wrapper around retry that uses a FibonacciRetry backoff. See
// NewFibonacci.
// TODO is this useful or should we move to example?
func FibonacciRetry(ctx context.Context, base time.Duration, f retry.RetryFunc) error {
	b, err := NewFibonacci(base)
	if err != nil {
		return fmt.Errorf("failed to create fibonacci backoff: %w", err)

	}
	return retry.Do(ctx, b, f)
}

// NewFibonacci creates a new Fibonacci backoff that follows the fibonacci sequence
// multipled by base. The wait time is the sum of the previous two wait times on each
// previous attempt base * (1, 2, 3, 5, 8, 13...).
//
// Once it overflows, the function constantly returns the maximum time.Duration
// for a 64-bit integer.
//
// It returns an error if the given base is less than zero.
func NewFibonacci(base time.Duration) (backoff.Backoff, error) {
	if base <= 0 {
		return nil, fmt.Errorf("base must be greater than 0")
	}

	return &fibonacciBackoff{
		base:  base,
		state: unsafe.Pointer(&state{0, base}),
	}, nil
}

// Next implements Backoff. It is safe for concurrent use.
func (b *fibonacciBackoff) Next() (time.Duration, bool) {
	for {
		curr := atomic.LoadPointer(&b.state)
		currState := (*state)(curr)
		next := currState[0] + currState[1]

		if next <= 0 {
			return math.MaxInt64, false
		}

		if atomic.CompareAndSwapPointer(&b.state, curr, unsafe.Pointer(&state{currState[1], next})) {
			return next, false
		}
	}
}

func (b *fibonacciBackoff) Reset() {
	atomic.StorePointer(&b.state, unsafe.Pointer(&state{0, b.base}))
}
