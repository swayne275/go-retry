package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestConstantRetry(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) error {
			return RetryableError(fmt.Errorf("some retryable err"))
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := ConstantRetry(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RetryFunc_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		nonRetryableCnt := 3
		errNonRetryable := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return errNonRetryable
			}

			return RetryableError(fmt.Errorf("some non-retryable error"))
		}

		err := ConstantRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, errFunctionReturnedNonRetryableError) {
			t.Errorf("expected %q to be %q", err, errFunctionReturnedNonRetryableError)
		}
		if !errors.Is(err, errNonRetryable) {
			t.Errorf("expected %q to be %q", err, errNonRetryable)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})
}

func TestExponentialRetry(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) error {
			return RetryableError(fmt.Errorf("some retryable err"))
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := ExponentialRetry(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RetryFunc_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		nonRetryableCnt := 3
		errNonRetryable := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return errNonRetryable
			}

			return RetryableError(fmt.Errorf("some non-retryable error"))
		}

		err := ExponentialRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, errFunctionReturnedNonRetryableError) {
			t.Errorf("expected %q to be %q", err, errFunctionReturnedNonRetryableError)
		}
		if !errors.Is(err, errNonRetryable) {
			t.Errorf("expected %q to be %q", err, errNonRetryable)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})
}

func TestFibonacciRetry(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_context_cancelled", func(t *testing.T) {
		t.Parallel()

		f := func(_ context.Context) error {
			return RetryableError(fmt.Errorf("some retryable err"))
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()

		if err := FibonacciRetry(ctx, 1*time.Nanosecond, f); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RetryFunc_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		nonRetryableCnt := 3
		errNonRetryable := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return errNonRetryable
			}

			return RetryableError(fmt.Errorf("some non-retryable error"))
		}

		err := FibonacciRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, errFunctionReturnedNonRetryableError) {
			t.Errorf("expected %q to be %q", err, errFunctionReturnedNonRetryableError)
		}
		if !errors.Is(err, errNonRetryable) {
			t.Errorf("expected %q to be %q", err, errNonRetryable)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})
}
