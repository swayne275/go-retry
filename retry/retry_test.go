package retry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/swayne275/go-retry/backoff"
)

func TestRetryableError(t *testing.T) {
	t.Parallel()

	err := RetryableError(fmt.Errorf("oops"))
	if got, want := err.Error(), "retryable: "; !strings.Contains(got, want) {
		t.Errorf("expected %v to contain %v", got, want)
	}
}

func TestDo(t *testing.T) {
	t.Parallel()

	t.Run("exit_on_non_retryable", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		maxRetryBackoff := backoff.WithMaxRetries(3, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		var i int
		if err := Do(ctx, maxRetryBackoff, func(_ context.Context) error {
			i++
			return fmt.Errorf("oops") // not retryable
		}); err == nil {
			t.Fatal("expected err")
		}

		if got, want := i, 1; got != want {
			t.Errorf("expected %v to be %v", got, want)
		}
	})

	t.Run("unwraps", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		maxRetryBackoff := backoff.WithMaxRetries(1, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		err := Do(ctx, maxRetryBackoff, func(_ context.Context) error {
			return RetryableError(io.EOF)
		})
		if err == nil {
			t.Fatal("expected err")
		}

		if !errors.Is(err, io.EOF) {
			t.Errorf("expected %q to be %q", err, io.EOF)
		}
	})

	t.Run("exit_no_error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		maxRetryBackoff := backoff.WithMaxRetries(3, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		var i int
		if err := Do(ctx, maxRetryBackoff, func(_ context.Context) error {
			i++
			return nil // no error
		}); err != nil {
			t.Fatal("expected no err")
		}

		if got, want := i, 1; got != want {
			t.Errorf("expected %v to be %v", got, want)
		}
	})

	t.Run("exit_on_context_canceled", func(t *testing.T) {
		t.Parallel()

		b, err := backoff.NewConstant(1 * time.Nanosecond)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		retryFunc := func(_ context.Context) error {
			return RetryableError(fmt.Errorf("some retryable error"))
		}

		go func() {
			time.Sleep(10 * time.Nanosecond)
			cancel()
		}()
		if err = Do(ctx, b, retryFunc); err != context.Canceled {
			t.Errorf("expected %q to be %q", err, context.Canceled)
		}
	})

	t.Run("exit_on_RetryFunc_nonretryable_error", func(t *testing.T) {
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

			return RetryableError(fmt.Errorf("some retryable error"))
		}

		if err = Do(context.Background(), b, retryFunc); !errors.Is(err, ErrNonRetryable) {
			t.Errorf("expected %q to contain %q", err, ErrNonRetryable)
		}
		if cnt != maxCnt+1 {
			t.Errorf("expected %d to be %d", cnt, maxCnt+1)
		}
	})

	t.Run("exit_on_backoff_stop", func(t *testing.T) {
		t.Parallel()

		maxRetryBackoff := backoff.WithMaxRetries(3, backoff.BackoffFunc(func() (time.Duration, bool) {
			return 1 * time.Nanosecond, false
		}))

		errUnderlyingRetryable := RetryableError(fmt.Errorf("some retryable error"))
		err := Do(context.Background(), maxRetryBackoff, func(_ context.Context) error {
			return RetryableError(errUnderlyingRetryable)
		})
		if !errors.Is(err, errBackoffSignaledToStop) {
			t.Errorf("expected %q to be %q", err, errBackoffSignaledToStop)
		}
		if !errors.Is(err, errUnderlyingRetryable) {
			t.Errorf("expected %q to be %q", err, errUnderlyingRetryable)
		}
	})
}

func TestCancel(t *testing.T) {
	for i := 0; i < 100000; i++ {
		ctx, cancel := context.WithCancel(context.Background())

		calls := 0
		rf := func(ctx context.Context) error {
			calls++
			// Never succeed.
			// Always return a RetryableError
			return RetryableError(errors.New("nope"))
		}

		const delay time.Duration = time.Millisecond
		b, err := backoff.NewConstant(delay)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		const maxRetries = 5
		b = backoff.WithMaxRetries(maxRetries, b)

		const jitter time.Duration = 5 * time.Millisecond
		b, err = backoff.WithJitter(jitter, b)
		if err != nil {
			t.Fatalf("failed to add jitter: %v", err)
		}

		// Here we cancel the Context *before* the call to Do
		cancel()
		Do(ctx, b, rf)

		if calls > 1 {
			t.Errorf("rf was called %d times instead of 0 or 1", calls)
		}
	}
}

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
		nonRetryableError := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return nonRetryableError
			}

			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := ConstantRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, ErrNonRetryable) {
			t.Errorf("expected %q to be %q", err, ErrNonRetryable)
		}
		if !errors.Is(err, nonRetryableError) {
			t.Errorf("expected %q to be %q", err, nonRetryableError)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})

	t.Run("retry_until_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxRetries := 5
		nonRetryableErr := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++
			if cnt >= maxRetries {
				return nonRetryableErr
			}
			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := ConstantRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, nonRetryableErr) {
			t.Errorf("expected %q to be %q", err, nonRetryableErr)
		}
		if cnt != maxRetries {
			t.Errorf("expected %d to be %d", cnt, maxRetries)
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
		nonRetriableErr := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return nonRetriableErr
			}

			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := ExponentialRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, ErrNonRetryable) {
			t.Errorf("expected %q to be %q", err, ErrNonRetryable)
		}
		if !errors.Is(err, nonRetriableErr) {
			t.Errorf("expected %q to be %q", err, nonRetriableErr)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})

	t.Run("retry_until_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxRetries := 5
		nonRetriableErr := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++
			if cnt == maxRetries {
				return nonRetriableErr
			}
			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := ExponentialRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, nonRetriableErr) {
			t.Errorf("expected %q to be %q", err, nonRetriableErr)
		}
		if cnt != maxRetries {
			t.Errorf("expected %d to be %d", cnt, maxRetries)
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
		nonRetriableErr := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++

			if cnt > nonRetryableCnt {
				return nonRetriableErr
			}

			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := FibonacciRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, ErrNonRetryable) {
			t.Errorf("expected %q to be %q", err, ErrNonRetryable)
		}
		if !errors.Is(err, nonRetriableErr) {
			t.Errorf("expected %q to be %q", err, nonRetriableErr)
		}
		if cnt != nonRetryableCnt+1 {
			t.Errorf("expected %d to be %d", cnt, nonRetryableCnt+1)
		}
	})

	t.Run("retry_until_nonretryable_error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		maxRetries := 5
		nonRetryableErr := fmt.Errorf("some non-retryable error")
		f := func(_ context.Context) error {
			cnt++
			if cnt == maxRetries {
				return nonRetryableErr
			}
			return RetryableError(fmt.Errorf("some retryable error"))
		}

		err := FibonacciRetry(context.Background(), 1*time.Nanosecond, f)
		if !errors.Is(err, nonRetryableErr) {
			t.Errorf("expected %q to be %q", err, nonRetryableErr)
		}
		if cnt != maxRetries {
			t.Errorf("expected %d to be %d", cnt, maxRetries)
		}
	})
}
