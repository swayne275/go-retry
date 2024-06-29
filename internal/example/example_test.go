package example

import (
	"context"
	"fmt"
	"time"

	"github.com/swayne275/go-retry"
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
	b, err := retry.NewFibonacci(1 * time.Second)
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

	b, err := retry.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle the error here, likely from bad input
	}
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

	b, err := retry.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
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

	b, err := retry.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
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

	b, err := retry.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
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

	b, err := retry.NewFibonacci(1 * time.Second)
	if err != nil {
		// handle err
	}
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

func ExampleNewFibonacci() {
	b, err := retry.NewFibonacci(1 * time.Second)
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
