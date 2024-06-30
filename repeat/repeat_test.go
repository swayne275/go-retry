package repeat

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		b, err := backoff.NewConstant(1 * time.Nanosecond)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		retryFunc := func(_ context.Context) bool { return true }

		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()
		if err = Do(ctx, b, retryFunc); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RepeatFunc_false", func(t *testing.T) {
		t.Parallel()

		b, err := backoff.NewConstant(1 * time.Nanosecond)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		cnt := 0
		maxCnt := 3
		retryFunc := func(_ context.Context) bool {
			cnt++
			return cnt <= maxCnt
		}

		if err = Do(context.Background(), b, retryFunc); err != errFunctionSignaledToStop {
			t.Errorf("expected %q to be %q", err, errFunctionSignaledToStop)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})

	t.Run("exit_on_backoff_stop", func(t *testing.T) {
		t.Parallel()

		b := backoff.WithMaxRetries(3, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		retryFunc := func(_ context.Context) bool { return true }

		if err := Do(context.Background(), b, retryFunc); err != errBackoffSignaledToStop {
			t.Errorf("expected %q to be %q", err, errBackoffSignaledToStop)
		}
	})
}

func TestDoUntilError(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		b, err := backoff.NewConstant(1 * time.Nanosecond)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		retryFunc := func(_ context.Context) error { return nil }

		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()
		if err = DoUntilError(ctx, b, retryFunc); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RepeatFunc_error", func(t *testing.T) {
		t.Parallel()

		b, err := backoff.NewConstant(1 * time.Nanosecond)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		cnt := 0
		maxCnt := 3
		retryFunc := func(_ context.Context) error {
			cnt++
			if cnt > maxCnt {
				return fmt.Errorf("function error")
			}
			return nil
		}

		if err = DoUntilError(context.Background(), b, retryFunc); !errors.Is(err, errFunctionSignaledToStop) {
			t.Errorf("expected %q to contain %q", err, errFunctionSignaledToStop)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})

	t.Run("exit_on_backoff_stop", func(t *testing.T) {
		t.Parallel()

		b := backoff.WithMaxRetries(3, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		retryFunc := func(_ context.Context) error { return nil }

		if err := DoUntilError(context.Background(), b, retryFunc); err != errBackoffSignaledToStop {
			t.Errorf("expected %q to be %q", err, errBackoffSignaledToStop)
		}
	})
}
