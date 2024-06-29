package retry_test

import (
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/swayne275/go-retry"
)

func TestFibonacciBackoff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		base      time.Duration
		tries     int
		exp       []time.Duration
		expectErr bool
	}{
		{
			name:  "single",
			base:  1 * time.Nanosecond,
			tries: 1,
			exp: []time.Duration{
				1 * time.Nanosecond,
			},
		},
		{
			name:  "max",
			base:  10 * time.Millisecond,
			tries: 5,
			exp: []time.Duration{
				10 * time.Millisecond,
				20 * time.Millisecond,
				30 * time.Millisecond,
				50 * time.Millisecond,
				80 * time.Millisecond,
			},
		},
		{
			name:  "many",
			base:  1 * time.Nanosecond,
			tries: 14,
			exp: []time.Duration{
				1 * time.Nanosecond,
				2 * time.Nanosecond,
				3 * time.Nanosecond,
				5 * time.Nanosecond,
				8 * time.Nanosecond,
				13 * time.Nanosecond,
				21 * time.Nanosecond,
				34 * time.Nanosecond,
				55 * time.Nanosecond,
				89 * time.Nanosecond,
				144 * time.Nanosecond,
				233 * time.Nanosecond,
				377 * time.Nanosecond,
				610 * time.Nanosecond,
			},
		},
		{
			name:  "overflow",
			base:  100_000 * time.Hour,
			tries: 10,
			exp: []time.Duration{
				100_000 * time.Hour,
				200_000 * time.Hour,
				300_000 * time.Hour,
				500_000 * time.Hour,
				800_000 * time.Hour,
				1_300_000 * time.Hour,
				2_100_000 * time.Hour,
				math.MaxInt64,
				math.MaxInt64,
				math.MaxInt64,
			},
		},
		{
			name:      "bad input duration",
			base:      0 * time.Nanosecond,
			tries:     0,
			exp:       []time.Duration{},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := retry.NewFibonacci(tc.base)
			if tc.expectErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultsCh := make(chan time.Duration, tc.tries)
			for i := 0; i < tc.tries; i++ {
				go func() {
					r, _ := b.Next()
					resultsCh <- r
				}()
			}

			results := make([]time.Duration, tc.tries)
			for i := 0; i < tc.tries; i++ {
				select {
				case val := <-resultsCh:
					results[i] = val
				case <-time.After(5 * time.Second):
					t.Fatal("timeout")
				}
			}
			sort.Slice(results, func(i, j int) bool {
				return results[i] < results[j]
			})

			if !reflect.DeepEqual(results, tc.exp) {
				t.Errorf("expected \n\n%v\n\n to be \n\n%v\n\n", results, tc.exp)
			}
		})
	}
}

func TestFibonacciBackoff_WithReset(t *testing.T) {
	base := 1 * time.Second
	numRounds := 5
	expected := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		8 * time.Second,
	}

	b, err := retry.NewFibonacci(base)
	if err != nil {
		t.Fatalf("failed to create fibonacci backoff: %v", err)
	}

	resettableB := retry.WithReset(func() retry.Backoff {
		newB, err := retry.NewFibonacci(base)
		if err != nil {
			t.Fatalf("failed to reset fibonacci backoff: %v", err)
		}

		return newB
	}, b)

	// test pre reset
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expected[i] {
			t.Errorf("pre reset: expected %v to be %v", val, expected[i])
		}
	}

	// test post reset. since we reset we expect the same sequence of values as before
	resettableB.Reset()
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expected[i] {
			t.Errorf("post reset: expected %v to be %v", val, expected[i])
		}
	}
}

func TestFibonacciBackoff_WithReset_ChangeBase(t *testing.T) {
	base := 1 * time.Second
	numRounds := 5
	expected := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		8 * time.Second,
	}

	b, err := retry.NewFibonacci(base)
	if err != nil {
		t.Fatalf("failed to create fibonacci backoff: %v", err)
	}

	newBase := 2 * time.Second
	newExpected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		6 * time.Second,
		10 * time.Second,
		16 * time.Second,
	}
	resettableB := retry.WithReset(func() retry.Backoff {
		newB, err := retry.NewFibonacci(newBase)
		if err != nil {
			t.Fatalf("failed to reset fibonacci backoff: %v", err)
		}

		return newB
	}, b)

	// test pre reset
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expected[i] {
			t.Errorf("pre reset: expected %v to be %v", val, expected[i])
		}
	}

	// test post reset. since we reset with a new base we expect new values
	resettableB.Reset()
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != newExpected[i] {
			t.Errorf("post reset: expected %v to be %v", val, expected[i])
		}
	}
}

func TestFibonacciBackoff_WithCappedDuration_WithReset(t *testing.T) {
	base := 1 * time.Second
	cappedDuration := 5 * time.Second
	numRounds := 5
	expectedCapped := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		5 * time.Second,
	}

	b, err := retry.NewFibonacci(base)
	if err != nil {
		t.Fatalf("failed to create fibonacci backoff: %v", err)
	}

	cappedB := retry.WithCappedDuration(cappedDuration, b)

	// test pre reset
	for i := 0; i < numRounds; i++ {
		val, _ := cappedB.Next()
		if val != expectedCapped[i] {
			t.Errorf("pre reset: expected %v to be %v", val, expectedCapped[i])
		}
	}

	// test post reset. since we reset we expect the same sequence of values as before
	// and the cap should still be applied.
	cappedB.Reset()
	for i := 0; i < numRounds; i++ {
		val, _ := cappedB.Next()
		if val != expectedCapped[i] {
			t.Errorf("post reset: expected %v to be %v", val, expectedCapped[i])
		}
	}

	// test post user-defined reset.
	// since we define the reset function without decorators, the decorators should not be observed.
	expectedAfterExplicitReset := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		3 * time.Second,
		5 * time.Second,
		8 * time.Second,
	}

	resettableB := retry.WithReset(func() retry.Backoff {
		// don't set a cap on the explicit reset
		newB, err := retry.NewFibonacci(base)
		if err != nil {
			t.Fatalf("failed to reset fibonacci backoff: %v", err)
		}

		return newB
	}, cappedB)

	resettableB.Reset()
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expectedAfterExplicitReset[i] {
			t.Errorf("post reset: expected %v to be %v", val, expectedAfterExplicitReset[i])
		}
	}
}
