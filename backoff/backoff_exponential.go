package backoff

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

type exponentialBackoff struct {
	base    time.Duration
	attempt uint64
}

// NewExponential creates a new exponential backoff using the starting value of
// base and doubling on each failure (1, 2, 4, 8, 16, 32, 64...), up to max.
//
// Once it overflows, the function constantly returns the maximum time.Duration
// for a 64-bit integer.
//
// It returns an error if the given base is less than zero.
func NewExponential(base time.Duration) (Backoff, error) {
	if base <= 0 {
		return nil, fmt.Errorf("base must be greater than 0")
	}

	return &exponentialBackoff{
		base: base,
	}, nil
}

// Next implements Backoff. It is safe for concurrent use.
func (b *exponentialBackoff) Next() (time.Duration, bool) {
	next := b.base << (atomic.AddUint64(&b.attempt, 1) - 1)
	if next <= 0 {
		atomic.AddUint64(&b.attempt, ^uint64(0))
		next = math.MaxInt64
	}

	return next, false
}

func (b *exponentialBackoff) Reset() {
	atomic.StoreUint64(&b.attempt, 0)
}
