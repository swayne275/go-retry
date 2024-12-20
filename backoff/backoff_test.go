package backoff

import (
	"context"
	"testing"
	"time"
)

func TestWithJitter_BadValues(t *testing.T) {
	t.Parallel()

	t.Run("jitter is negative", func(t *testing.T) {
		backoffJitter := -1 * time.Millisecond
		_, err := WithJitter(backoffJitter, BackoffFunc(func() (time.Duration, bool) {
			return 0, false
		}))
		if err == nil {
			t.Error("expected error, got none")
		}

		if err != ErrInvalidJitter {
			t.Errorf("expected %v, got %v", ErrInvalidJitter, err)
		}
	})

	t.Run("jitter is zero", func(t *testing.T) {
		backoffJitter := 0 * time.Millisecond
		_, err := WithJitter(backoffJitter, BackoffFunc(func() (time.Duration, bool) {
			return 0, false
		}))
		if err == nil {
			t.Error("expected error, got none")
		}

		if err != ErrInvalidJitter {
			t.Errorf("expected %v, got %v", ErrInvalidJitter, err)
		}
	})
}

func TestWithJitter(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	backoffJitter := 250 * time.Millisecond

	sawJitter := false
	for i := 0; i < 100_000; i++ {
		b, err := WithJitter(backoffJitter, BackoffFunc(func() (time.Duration, bool) {
			return baseDuration, false
		}))
		if err != nil {
			t.Fatalf("failed to create backoff with jitter: %v", err)
		}

		val, stop := b.Next()
		if stop {
			t.Errorf("should not stop")
		}

		if val != baseDuration {
			sawJitter = true
		}

		if min, max := baseDuration-backoffJitter, baseDuration+backoffJitter; val < min || val > max {
			t.Errorf("expected %v to be between %v and %v", val, min, max)
		}
	}

	if !sawJitter {
		t.Fatal("expected to see jitter, all values were the same")
	}
}

func TestWithJitterPercent_BadValues(t *testing.T) {
	t.Parallel()

	t.Run("jitter is 0", func(t *testing.T) {
		backoffJitterPercent := uint64(0)
		_, err := WithJitterPercent(backoffJitterPercent, BackoffFunc(func() (time.Duration, bool) {
			return 0, false
		}))
		if err == nil {
			t.Error("expected error, got none")
		}

		if err != ErrInvalidJitterPercent {
			t.Errorf("expected %v, got %v", ErrInvalidJitterPercent, err)
		}
	})

	t.Run("jitter is >100", func(t *testing.T) {
		backoffJitterPercent := uint64(101)
		_, err := WithJitterPercent(backoffJitterPercent, BackoffFunc(func() (time.Duration, bool) {
			return 0, false
		}))
		if err == nil {
			t.Error("expected error, got none")
		}

		if err != ErrInvalidJitterPercent {
			t.Errorf("expected %v, got %v", ErrInvalidJitterPercent, err)
		}
	})
}

func TestWithJitterPercent(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	jitterPercent := uint64(5)
	minBackoff := time.Duration(100-jitterPercent) * baseDuration / 100
	maxBackoff := time.Duration(100+jitterPercent) * baseDuration / 100

	sawJitter := false
	for i := 0; i < 100_000; i++ {
		b, err := WithJitterPercent(jitterPercent, BackoffFunc(func() (time.Duration, bool) {
			return baseDuration, false
		}))
		if err != nil {
			t.Fatalf("failed to create backoff with jitter percent: %v", err)
		}

		val, stop := b.Next()
		if stop {
			t.Errorf("should not stop")
		}

		if val != baseDuration {
			sawJitter = true
		}

		if val < minBackoff || val > maxBackoff {
			t.Errorf("expected %v to be between %v and %v", val, minBackoff, maxBackoff)
		}
	}

	if !sawJitter {
		t.Fatal("expected to see jitter, all values were the same")
	}
}

func TestWithMaxRetries(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	maxRetries := uint64(3)
	backoff := WithMaxRetries(maxRetries, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	// First 3 attempts succeed
	for i := uint64(0); i < maxRetries; i++ {
		val, stop := backoff.Next()
		if stop {
			t.Errorf("should not stop")
		}
		if val != baseDuration {
			t.Errorf("expected %v to be %v", val, baseDuration)
		}
	}

	// Now we stop
	val, stop := backoff.Next()
	if !stop {
		t.Errorf("should stop")
	}
	if val != 0 {
		t.Errorf("expected %v to be %v", val, 0)
	}
}

func TestWithCappedDuration(t *testing.T) {
	t.Parallel()

	baseDuration := 5 * time.Second
	cappedDuration := 3 * time.Second
	backoff := WithCappedDuration(cappedDuration, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	val, stop := backoff.Next()
	if stop {
		t.Errorf("should not stop")
	}
	if val != cappedDuration {
		t.Errorf("expected %v to be %v", val, cappedDuration)
	}
}

func TestWithMaxDuration(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	maxDuration := 250 * time.Millisecond
	backoff := WithMaxDuration(maxDuration, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	validateMaxDuration(t, backoff, maxDuration)
}

func TestWithContext(t *testing.T) {
	t.Parallel()

	baseDuration := 2 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	backoff := WithContext(ctx, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	val, stop := backoff.Next()
	if stop {
		t.Errorf("should not stop")
	}
	if val != baseDuration {
		t.Errorf("expected %v to be %v", val, baseDuration)
	}

	cancel()
	_, stop = backoff.Next()
	if !stop {
		t.Errorf("should stop after context cancel")
	}
}

func TestResettableBackoff(t *testing.T) {
	var attempt uint64
	backoff := WithReset(func() Backoff {
		attempt = 0

		return BackoffFunc(func() (time.Duration, bool) {
			attempt++
			return time.Duration(attempt) * time.Second, false
		})
	}, BackoffFunc(func() (time.Duration, bool) {
		attempt++
		return time.Duration(attempt) * time.Second, false
	}))

	// Call Next a few times
	for i := 0; i < 3; i++ {
		val, stop := backoff.Next()
		if stop {
			t.Fatal("should not stop")
		}
		if val != time.Duration(i+1)*time.Second {
			t.Errorf("expected %v to be %v", val, time.Duration(i+1)*time.Second)
		}
	}

	// Reset the backoff
	backoff.Reset()

	// Call Next again and verify that the state has been reset
	val, stop := backoff.Next()
	if stop {
		t.Fatal("should not stop after reset")
	}
	if val != time.Second {
		t.Errorf("expected %v to be %v", val, time.Second)
	}
}

func TestResettableBackoff_WithJitter(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	jitterDuration := 1 * time.Second
	b, err := WithJitter(jitterDuration, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))
	if err != nil {
		t.Fatalf("failed to create backoff with jitter: %v", err)
	}

	// reset it and verify that we are still within the jitter range
	b.reset()

	sawJitter := false
	for i := 0; i < 100_000; i++ {
		val, stop := b.Next()
		if stop {
			t.Errorf("should not stop")
		}

		if val != 1*time.Second {
			sawJitter = true
		}
		if min, max := baseDuration-jitterDuration, baseDuration+jitterDuration; val < min || val > max {
			t.Errorf("expected %v to be between %v and %v", val, min, max)
		}
	}

	if !sawJitter {
		t.Fatal("expected to see jitter, all values were the same")
	}
}

func TestResettableBackoff_WithJitterPercent(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	jitterPercent := uint64(5)
	minBackoff := time.Duration(100-jitterPercent) * baseDuration / 100
	maxBackoff := time.Duration(100+jitterPercent) * baseDuration / 100
	b, err := WithJitterPercent(jitterPercent, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))
	if err != nil {
		t.Fatalf("failed to create backoff with jitter percent: %v", err)
	}

	// reset it and verify that we are still within the jitter range
	b.reset()

	sawJitter := false
	for i := 0; i < 100_000; i++ {
		val, stop := b.Next()
		if stop {
			t.Errorf("should not stop")
		}

		if val != 1*time.Second {
			sawJitter = true
		}
		if val < minBackoff || val > maxBackoff {
			t.Errorf("expected %v to be between %v and %v", val, minBackoff, maxBackoff)
		}
	}

	if !sawJitter {
		t.Fatal("expected to see jitter, all values were the same")
	}
}

func TestResettableBackoff_WithMaxRetries(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	maxRetries := uint64(3)
	backoff := WithMaxRetries(maxRetries, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	// First 3 attempts succeed
	for i := uint64(0); i < maxRetries; i++ {
		val, stop := backoff.Next()
		if stop {
			t.Errorf("should not stop")
		}
		if val != baseDuration {
			t.Errorf("expected %v to be %v", val, baseDuration)
		}
	}

	backoff.reset()

	// reset - should get 3 more succeessful attempts
	for i := uint64(0); i < maxRetries; i++ {
		val, stop := backoff.Next()
		if stop {
			t.Errorf("should not stop after reset")
		}
		if val != baseDuration {
			t.Errorf("expected %v to be %v", val, baseDuration)
		}
	}

	// Now we stop
	val, stop := backoff.Next()
	if !stop {
		t.Errorf("should stop")
	}
	if val != 0 {
		t.Errorf("expected %v to be %v", val, 0)
	}
}

func TestResettableBackoff_WithCappedDuration(t *testing.T) {
	t.Parallel()

	baseDuration := 5 * time.Second
	cappedDuration := 3 * time.Second
	backoff := WithCappedDuration(cappedDuration, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	val, stop := backoff.Next()
	if stop {
		t.Errorf("should not stop")
	}
	if val != cappedDuration {
		t.Errorf("expected %v to be %v", val, cappedDuration)
	}

	// verify that we still have cappedDuration after a reset
	backoff.reset()

	val, stop = backoff.Next()
	if stop {
		t.Errorf("should not stop")
	}
	if val != cappedDuration {
		t.Errorf("expected %v to be %v", val, cappedDuration)
	}
}

func TestResettableBackoff_WithMaxDuration(t *testing.T) {
	t.Parallel()

	baseDuration := 1 * time.Second
	maxDuration := 250 * time.Millisecond
	backoff := WithMaxDuration(maxDuration, BackoffFunc(func() (time.Duration, bool) {
		return baseDuration, false
	}))

	validateMaxDuration(t, backoff, maxDuration)

	// a reset should clear it, and we do the process again
	backoff.reset()

	validateMaxDuration(t, backoff, maxDuration)
}

// TestResettableBackoff_MultipleDecorators ensures that multiple decorators can be applied to a ResettableBackoff
// and that the decorators are still observed after a reset.
func TestResettableBackoff_MultipleDecorators(t *testing.T) {
	base := 1 * time.Second
	cappedDuration := 5 * time.Second
	maxRetries := uint64(7)
	expected := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		5 * time.Second,
		5 * time.Second,
		5 * time.Second,
	}

	b, err := NewFibonacci(base)
	if err != nil {
		t.Fatalf("failed to create fibonacci backoff: %v", err)
	}

	cappedBackoff := WithCappedDuration(cappedDuration, b)
	maxRetriesBackoff := WithMaxRetries(maxRetries, cappedBackoff)

	for _, tc := range expected {
		val, stop := maxRetriesBackoff.Next()
		if stop {
			t.Errorf("pre reset should not stop")
		}
		if val != tc {
			t.Errorf("pre reset expected %v to be %v", val, tc)
		}
	}

	// we expect it to stop after the max number of retries
	_, stop := maxRetriesBackoff.Next()
	if !stop {
		t.Errorf("pre reset should stop")
	}

	// reset it and verify that we repeat the above
	maxRetriesBackoff.Reset()

	for _, tc := range expected {
		val, stop := maxRetriesBackoff.Next()
		if stop {
			t.Errorf("post reset should not stop")
		}
		if val != tc {
			t.Errorf("post reset expected %v to be %v", val, tc)
		}
	}

	// we again expect it to stop after the max number of retries
	_, stop = maxRetriesBackoff.Next()
	if !stop {
		t.Errorf("post reset should stop")
	}
}

func validateMaxDuration(t *testing.T, b *ResettableBackoff, maxDuration time.Duration) {
	t.Helper()

	// Take once, within timeout.
	val, stop := b.Next()
	if stop {
		t.Error("should not stop")
	}

	if val > maxDuration {
		t.Errorf("expected %v to be less than %v", val, maxDuration)
	}

	// sleep for 80% of max duration
	longSleep80 := time.Duration(8) * maxDuration / 10
	time.Sleep(longSleep80)

	// Take again, remainder contines
	val, stop = b.Next()
	if stop {
		t.Error("should not stop")
	}

	// val should be <= 20% of max duration since we slept for 80% of it above
	shortSleep20 := time.Duration(2) * maxDuration / 10
	if val > shortSleep20 {
		t.Errorf("expected %v to be less than %v", val, shortSleep20)
	}

	time.Sleep(shortSleep20)

	// Now we stop
	val, stop = b.Next()
	if !stop {
		t.Errorf("should stop")
	}
	if val != 0 {
		t.Errorf("expected %v to be %v", val, 0)
	}
}
