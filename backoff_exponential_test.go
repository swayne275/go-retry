package retry_test

import (
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/sethvargo/go-retry"
)

func TestExponentialBackoff(t *testing.T) {
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
			name:  "many",
			base:  1 * time.Nanosecond,
			tries: 14,
			exp: []time.Duration{
				1 * time.Nanosecond,
				2 * time.Nanosecond,
				4 * time.Nanosecond,
				8 * time.Nanosecond,
				16 * time.Nanosecond,
				32 * time.Nanosecond,
				64 * time.Nanosecond,
				128 * time.Nanosecond,
				256 * time.Nanosecond,
				512 * time.Nanosecond,
				1024 * time.Nanosecond,
				2048 * time.Nanosecond,
				4096 * time.Nanosecond,
				8192 * time.Nanosecond,
			},
		},
		{
			name:  "overflow",
			base:  100_000 * time.Hour,
			tries: 10,
			exp: []time.Duration{
				100_000 * time.Hour,
				200_000 * time.Hour,
				400_000 * time.Hour,
				800_000 * time.Hour,
				1_600_000 * time.Hour,
				math.MaxInt64,
				math.MaxInt64,
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

			b, err := retry.NewExponential(tc.base)
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

func TestExponentialBackoff_WithReset(t *testing.T) {
	base := 2 * time.Second
	numRounds := 3
	expected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}

	b, err := retry.NewExponential(base)
	if err != nil {
		t.Fatalf("failed to create exponential backoff: %v", err)
	}

	// TODO should calling code even provide a reset func???
	resettableB := retry.WithReset(func() {}, b)

	// test pre reset
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expected[i] {
			t.Errorf("pre reset: expected %v to be %v", val, expected[i])
		}

	}

	resettableB.Reset()

	// test post reset. since we reset we expect the same sequence of values as before
	// TODO this doesn't work
	for i := 0; i < numRounds; i++ {
		val, _ := resettableB.Next()
		if val != expected[i] {
			t.Errorf("post reset: expected %v to be %v", val, expected[i])
		}

	}
}
