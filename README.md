# Retry

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/mod/github.com/swayne275/go-retry)

Builds off of the wonderful work of https://github.com/sethvargo/go-retry but adds additional functionality:


## Synopsis

Retry is a Go library for facilitating retry logic and backoff strategies. It builds off the work of [go-retry](https://github.com/sethvargo/go-retry) and adds additional functionality, such as infinite retries until an error occurs and the ability to reset backoff durations. The library is highly extensible, allowing you to implement custom backoff strategies and integrate them seamlessly into your applications.

## Features

- **Infinite Retries Until Error** - Repeat will continually do whatever is in the RepeatFunc until it returns an error. It does not observe RetryableError, as there is no need for this functionality.

- **Backoff Reset** - You might want to reset a backoff of non-constant duration (e.g., if something happens that says you should poll faster, like a worker having a job available).

- **Extensible** - Inspired by Go's built-in HTTP package, this Go backoff and retry library is extensible via middleware. You can write custom backoff functions or use a provided filter.

- **Independent** - No external dependencies besides the Go standard library, meaning it won't bloat your project.

- **Concurrent** - Unless otherwise specified, everything is safe for concurrent use.

### Backoff Strategies

In addition to your own custom algorithms, there are built-in algorithms for backoff in the library.

#### Constant Backoff
Retries at a constant interval.

Example:

```text
1s -> 1s -> 1s -> 1s -> 1s -> 1s -> 1s
```

Usage:

```golang
NewConstant(1 * time.Second)
```

#### Exponential Backoff
Retries with exponentially increasing intervals.

Arguably the most common backoff, the next value is double the previous value.

Example:

```text
1s -> 2s -> 4s -> 8s -> 16s -> 32s -> 64s
```

Usage:

```golang
NewExponential(1 * time.Second)
```

#### Fibonacci Backoff
Retries with intervals following the Fibonacci sequence.

The next value is the sum of the current value and the previous value. This means retires happen quickly at first, but then gradually take slower, ideal for network-type issues.

Example:

```text
1s -> 2s -> 3s -> 5s -> 8s -> 13s
```

Usage:

```golang
NewFibonacci(1 * time.Second)
```

### Modifiers (Middleware)

The built-in backoff algorithms never terminate and have no caps or limits - you control their behavior with middleware. There's built-in middleware, but you can also write custom middleware.

#### Jitter
Adds randomness to the backoff intervals to prevent thundering herd problems.

To reduce the chances of a thundering herd, add random jitter to the returned
value.

```golang
backoff, err := NewFibonacci(1 * time.Second)

// Return the next value, +/- 500ms
backoffWithJitter, err := WithJitter(500*time.Millisecond, backoff)

// Return the next value, +/- 5% of the result
backoffWithJitterPercent, err := WithJitterPercent(5, backoff)
```

#### Capped Duration
Limits the maximum duration between retries.

To ensure an individual calculated duration never exceeds a value, use a cap:

```golang
backoff, err := NewFibonacci(1 * time.Second)

// Ensure the maximum value is 2s. In this example, the sleep values would be
// 1s, 1s, 2s, 2s, 2s, 2s...
backoffWithCap := WithCappedDuration(2 * time.Second, backoff)
```

#### Max Duration
Limits the maximum total time a backoff should execute.

For a best-effort limit on the total execution time, specify a max duration:

```golang
backoff, err := NewFibonacci(1 * time.Second)

// Ensure the maximum total retry time is 5s.
backoffWithMaxDuration = WithMaxDuration(5 * time.Second, backoff)
```

#### Max Retries
Limits the number of retry attempts.

To terminate a retry, specify the maximum number of _retries_. Note this is _retries_, not _attempts_. Attempts is retries + 1.

```golang
backoff, err := NewFibonacci(1 * time.Second)

// Stop after 4 retries, when the 5th attempt has failed. In this example, the worst case elapsed
// time would be 1s + 1s + 2s + 3s = 7s.
backoffWithMaxRetries = WithMaxRetries(4, backoff)
```

#### Context-Aware Backoff
Stops the backoff if the provided context is Done.

```golang
backoff, err := NewFibonacci(1 * time.Second)
ctx, cancel := context.WithTimeout(context.Background, 1 * time.Millisecond)

// backoff will return stop == true when context is cancelled
backoffWithContext := WithContext(ctx, backoff)
```

## Installation

To install the library, use the following command:

```sh
go get github.com/swayne275/go-retry
```

## Usage

### Basic Retry

This will retry the provided function until it either succeeds or returns a non-retryable error.

```golang
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/swayne275/go-retry"
    "github.com/swayne275/go-retry/backoff"
)

func main() {
    ctx := context.Background()
    backoff := backoff.NewConstant(1 * time.Second)

    err := retry.Do(ctx, backoff, func(ctx context.Context) error {
        // Your retryable function logic here
        return nil
    })

    if err != nil {
        fmt.Printf("Operation failed: %v\n", err)
    }
}
```

### Infinite Repeat Until Non Retryable Error

This will repeat the function until it returns a non-retryable error.

```golang
package main

import (
    "context"
    "fmt"

    "github.com/swayne275/go-retry/repeat"
)

func main() {
    ctx := context.Background()
    backoff := backoff.NewExponential(1 * time.Second)

    err := repeat.Do(ctx, backoff, func(ctx context.Context) bool {
        // Your function logic here - return false to stop repeating
        return true
    })

    if err != nil {
        // you can check why the repeat stopped with errors.Is() and the defined
        // types in the repeat package
        fmt.Printf("Operation failed: %v\n", err)
    }
}
```

### Backoff Reset

```golang
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/swayne275/go-retry"
    "github.com/swayne275/go-retry/backoff"
)

func main() {
    ctx := context.Background()
    backoff := backoff.NewExponential(1 * time.Second)
    resetFunc := func() Backoff {
      // define how the backoff shoudl be reset, likely something like:
      newB, err := NewExponential(base)
		  if err != nil {
			  t.Fatalf("failed to reset exponential backoff: %v", err)
		  }

		  return newB
    }
    backoffWithReset = backoff.WithReset(resetFunc, backoff)

    err := retry.Do(ctx, backoffWithReset, func(ctx context.Context) error {
        // Your retryable function logic here

        // something happens that makes you want to reset back to a shorter backoff
        b.Reset()

        return nil
    })

    if err != nil {
        fmt.Printf("Operation failed: %v\n", err)
    }
}
```

### Real World Example: Connecting to a SQL Database

```golang
package main

import (
  "context"
  "database/sql"
  "log"
  "time"

  "github.com/swayne275/go-retry"
)

func main() {
  db, err := sql.Open("mysql", "...")
  if err != nil {
    log.Fatal(err)
  }

  ctx := context.Background()
  if err := retry.FibonacciRetry(ctx, 1*time.Second, func(ctx context.Context) error {
    if err := db.PingContext(ctx); err != nil {
      // This marks the error as retryable
      return retry.RetryableError(err)
    }
    return nil
  }); err != nil {
    log.Fatal(err)
  }
}
```

## Benchmarks

Here are benchmarks against some other popular Go backoff and retry libraries. You can run these benchmarks yourself via the `benchmark/` folder. Commas and spacing fixed for clarity.

swayne275 is roughly in line with sethvargo, which is what it's based off of.

```text
cpu: Apple M3 Max
Benchmark/cenkalti-16         	 36881959	        32.40 ns/op	       0 B/op	       0 allocs/op
Benchmark/lestrrat-16         	  1990288	        612.5 ns/op	     128 B/op	       2 allocs/op
Benchmark/sethvargo-16        	345455916	        3.538 ns/op	       0 B/op	       0 allocs/op
Benchmark/swayne275-16        	339995433	        3.510 ns/op	       0 B/op	       0 allocs/op
```

## Notes and Caveats

- Randomization uses `math/rand` seeded with the Unix timestamp instead of `crypto/rand`.
- Ordering of addition of multiple modifiers will make a difference. For example; ensure you add `CappedDuration` before `WithMaxDuration`, otherwise it may bail out too early. Another example is you could add `Jitter` before or after capping depending on your desired outcome.
