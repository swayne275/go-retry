package example

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-retry"
	// TODO update this to import my version
)

func ExampleBackoffFunc() {
	ctx := context.Background()

	// Example backoff middleware that adds the provided duration t to the result.
	withShift := func(t time.Duration, next retry.Backoff) retry.BackoffFunc {
		return func() (time.Duration, bool) {
			val, stop := next.Next()
			if stop {
				return 0, true
			}
			return val + t, false
		}
	}

	// Middlewrap wrap another backoff:
	b := retry.NewFibonacci(1 * time.Second)
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

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithJitter(1*time.Second, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithJitterPercent() {
	ctx := context.Background()

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithJitterPercent(5, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithMaxRetries() {
	ctx := context.Background()

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithMaxRetries(3, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithCappedDuration() {
	ctx := context.Background()

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithCappedDuration(3*time.Second, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleWithMaxDuration() {
	ctx := context.Background()

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithMaxDuration(5*time.Second, b)

	if err := retry.Do(ctx, b, func(_ context.Context) error {
		// your retry logic here
		return nil
	}); err != nil {
		// handle the error here
	}
}

func ExampleNewConstant() {
	b, err := retry.NewConstant(1 * time.Second)
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
	b, err := retry.NewExponential(1 * time.Second)
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
