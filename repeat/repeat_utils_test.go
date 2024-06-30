package repeat

import (
	"context"
	"testing"
	"time"
)

func TestConstantRepeat(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) bool { return true }
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := ConstantRepeat(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RepeatFunc_false", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxCnt := 3
		f := func(_ context.Context) bool {
			cnt++

			return cnt <= maxCnt
		}

		if err := ConstantRepeat(context.Background(), 1*time.Nanosecond, f); err != errFunctionSignaledToStop {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})
}

func TestExponentialRepeat(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) bool { return true }
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := ExponentialRepeat(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RepeatFunc_false", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxCnt := 3
		f := func(_ context.Context) bool {
			cnt++

			return cnt <= maxCnt
		}

		if err := ExponentialRepeat(context.Background(), 1*time.Nanosecond, f); err != errFunctionSignaledToStop {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})
}

func TestFibonacciRepeat(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) bool { return true }
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := FibonacciRepeat(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RepeatFunc_false", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxCnt := 3
		f := func(_ context.Context) bool {
			cnt++

			return cnt <= maxCnt
		}

		if err := FibonacciRepeat(context.Background(), 1*time.Nanosecond, f); err != errFunctionSignaledToStop {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})
}
