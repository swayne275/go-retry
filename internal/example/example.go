package example

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/swayne275/go-retry/backoff"
	"github.com/swayne275/go-retry/retry"
)

func ExampleBackoffFunc() {
	ctx := context.Background()

	// Example backoff middleware that adds the provided duration t to the result.
	withShift := func(t time.Duration, next backoff.Backoff) backoff.BackoffFunc {
		return func() (time.Duration, bool) {
			val, stop := next.Next()
			if stop {
				return 0, true
			}
			return val + t, false
		}
	}

	// Middlewrap wrap another backoff:
	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle the error here, likely from bad input
	}
	b = withShift(5*time.Second, b)

	if err := retry.Do(ctx, b, func(ctx context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithJitter() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle the error here, likely from bad input
	}
	b, err = backoff.WithJitter(1*time.Second, b)
	if err != nil {
		// handle the error here, likely from bad input
	}

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithJitterPercent() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
	b, err = backoff.WithJitterPercent(5, b)
	if err != nil {
		// handle the error here, likely from bad input
	}

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithMaxRetries() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
	b = backoff.WithMaxRetries(3, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithCappedDuration() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
	b = backoff.WithCappedDuration(3*time.Second, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithMaxDuration() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
	b = backoff.WithMaxDuration(5*time.Second, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleNewConstant() {
	b, err := backoff.NewConstant(1 * time.Second)
	if err != nil {
		// handle the error here, likely from bad input
		return
	}

	for i := 0; i < 5; i++ {
		val, _ := b.Next()
		fmt.Printf("%v\n", val)
	}
	// Output:
	// 1s
	// 1s
	// 1s
	// 1s
	// 1s
}

func ExampleNewExponential() {
	b, err := backoff.NewExponential(1 * time.Second)
	if err != nil {
		// handle the error here, likely from bad input
		return
	}

	for i := 0; i < 5; i++ {
		val, _ := b.Next()
		fmt.Printf("%v\n", val)
	}
	// Output:
	// 1s
	// 2s
	// 4s
	// 8s
	// 16s
}

func ExampleNewFibonacci() {
	b, err := backoff.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}

	for i := 0; i < 5; i++ {
		val, _ := b.Next()
		fmt.Printf("%v\n", val)
	}
	// Output:
	// 1s
	// 2s
	// 3s
	// 5s
	// 8s
}

func ExampleDo_simple() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Nanosecond)
	if err != nil {
		// handle error
	}

	i := 0
	if err := retry.Do(ctx, backoff.WithMaxRetries(3, b), func(ctx context.Context) error {
		fmt.Printf("%d\n", i)
		i++
		return retry.RetryableError(fmt.Errorf("oops"))
	}); err != nil {
		// handle error
	}

	// Output:
	// 0
	// 1
	// 2
	// 3
}

func ExampleDo_customRetry() {
	ctx := context.Background()

	b, err := backoff.NewFibonacci(1 * time.Nanosecond)
	if err != nil {
		// handle error
	}

	// This example demonstrates selectively retrying specific errors. Only errors
	// wrapped with RetryableError are eligible to be retried.
	if err := retry.Do(ctx, backoff.WithMaxRetries(3, b), func(ctx context.Context) error {
		resp, err := http.Get("https://google.com/")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		switch resp.StatusCode / 100 {
		case 4:
			return fmt.Errorf("bad response: %v", resp.StatusCode)
		case 5:
			return retry.RetryableError(fmt.Errorf("bad response: %v", resp.StatusCode))
		default:
			return nil
		}
	}); err != nil {
		// handle error
	}
}
