package backoff

import (
	"sync"
	"time"

	"github.com/swayne275/go-retry/common/backoff"
)

// TODO clean up interface, struct, etc

var _ backoff.Backoff = (BackoffFunc)(nil)

// BackoffFunc is a backoff expressed as a function.
type BackoffFunc func() (time.Duration, bool)

// Next implements Backoff.
func (b BackoffFunc) Next() (time.Duration, bool) {
	return b()
}

func (b BackoffFunc) Reset() {}

type ResettableBackoff struct {
	backoff.Backoff
	// reset returns the backoff to its initial state.
	reset func()
}

func (b *ResettableBackoff) Next() (time.Duration, bool) {
	return b.Backoff.Next()
}

func (b *ResettableBackoff) Reset() {
	b.reset()
}

func WithReset(reset func() backoff.Backoff, next backoff.Backoff) *ResettableBackoff {
	rb := &ResettableBackoff{
		Backoff: next,
	}
	rb.reset = func() {
		rb.Backoff = reset()
	}

	return rb
}

// WithJitter wraps a backoff function and adds the specified jitter. j can be
// interpreted as "+/- j". For example, if j were 5 seconds and the backoff
// returned 20s, the value could be between 15 and 25 seconds. The value can
// never be less than 0.
func WithJitter(j time.Duration, next backoff.Backoff) *ResettableBackoff {
	r := newLockedRandom(time.Now().UnixNano())

	nextWithJitter := BackoffFunc(func() (time.Duration, bool) {
		val, stop := next.Next()
		if stop {
			return 0, true
		}

		diff := time.Duration(r.Int63n(int64(j)*2) - int64(j))
		val = val + diff
		if val < 0 {
			val = 0
		}
		return val, false
	})

	reset := func() backoff.Backoff {
		next.Reset()
		return nextWithJitter
	}

	return WithReset(reset, nextWithJitter)
}

// WithJitterPercent wraps a backoff function and adds the specified jitter
// percentage. j can be interpreted as "+/- j%". For example, if j were 5 and
// the backoff returned 20s, the value could be between 19 and 21 seconds. The
// value can never be less than 0 or greater than 100.
func WithJitterPercent(j uint64, next backoff.Backoff) *ResettableBackoff {
	r := newLockedRandom(time.Now().UnixNano())

	nextWithJitterPercent := BackoffFunc(func() (time.Duration, bool) {
		val, stop := next.Next()
		if stop {
			return 0, true
		}

		// Get a value between -j and j, the convert to a percentage
		top := r.Int63n(int64(j)*2) - int64(j)
		pct := 1 - float64(top)/100.0

		val = time.Duration(float64(val) * pct)
		if val < 0 {
			val = 0
		}
		return val, false
	})

	reset := func() backoff.Backoff {
		next.Reset()
		return nextWithJitterPercent
	}

	return WithReset(reset, nextWithJitterPercent)
}

// WithMaxRetries executes the backoff function up until the maximum attempts.
func WithMaxRetries(max uint64, next backoff.Backoff) *ResettableBackoff {
	var l sync.Mutex
	var attempt uint64

	nextWithMaxRetries := BackoffFunc(func() (time.Duration, bool) {
		l.Lock()
		defer l.Unlock()

		if attempt >= max {
			return 0, true
		}
		attempt++

		val, stop := next.Next()
		if stop {
			return 0, true
		}

		return val, false
	})

	reset := func() backoff.Backoff {
		l.Lock()
		defer l.Unlock()
		attempt = 0

		next.Reset()
		return nextWithMaxRetries
	}

	return WithReset(reset, nextWithMaxRetries)
}

// WithCappedDuration sets a maximum on the duration returned from the next
// backoff. This is NOT a total backoff time, but rather a cap on the maximum
// value a backoff can return. Without another middleware, the backoff will
// continue infinitely.
func WithCappedDuration(cap time.Duration, next backoff.Backoff) *ResettableBackoff {
	nextWithCappedDuration := BackoffFunc(func() (time.Duration, bool) {
		val, stop := next.Next()
		if stop {
			return 0, true
		}

		if val <= 0 || val > cap {
			val = cap
		}
		return val, false
	})

	reset := func() backoff.Backoff {
		next.Reset()
		return nextWithCappedDuration
	}

	return WithReset(reset, nextWithCappedDuration)
}

// WithMaxDuration sets a maximum on the total amount of time a backoff should
// execute. It's best-effort, and should not be used to guarantee an exact
// amount of time.
func WithMaxDuration(timeout time.Duration, next backoff.Backoff) *ResettableBackoff {
	var l sync.RWMutex
	start := time.Now()

	nextWithMaxDuration := BackoffFunc(func() (time.Duration, bool) {
		l.RLock()
		defer l.RUnlock()

		diff := timeout - time.Since(start)
		if diff <= 0 {
			return 0, true
		}

		val, stop := next.Next()
		if stop {
			return 0, true
		}

		if val <= 0 || val > diff {
			val = diff
		}
		return val, false
	})

	reset := func() backoff.Backoff {
		l.Lock()
		defer l.Unlock()
		start = time.Now()

		next.Reset()
		return nextWithMaxDuration
	}

	return WithReset(reset, nextWithMaxDuration)
}