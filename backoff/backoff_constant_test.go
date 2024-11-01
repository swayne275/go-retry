package backoff

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestConstantBackoff(t *testing.T) {
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
				10 * time.Millisecond,
				10 * time.Millisecond,
				10 * time.Millisecond,
				10 * time.Millisecond,
			},
		},
		{
			name:  "many",
			base:  1 * time.Nanosecond,
			tries: 14,
			exp: []time.Duration{
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
				1 * time.Nanosecond,
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

			b, err := NewConstant(tc.base)
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

func TestConstantBackoff_WithReset(t *testing.T) {
	expectedDuration := 3 * time.Second
	b, err := NewConstant(expectedDuration)
	if err != nil {
		t.Fatalf("failed to create constant backoff: %v", err)
	}

	resettableB := WithReset(func() Backoff {
		return b
	}, b)
	resettableB.Reset()

	val, _ := resettableB.Next()
	if val != expectedDuration {
		t.Errorf("expected %v to be %v", val, expectedDuration)
	}
}

func TestConstantBackoff_WithCappedDuration_WithReset(t *testing.T) {
	expectedDuration := 3 * time.Second
	cappedDuration := 2 * time.Second

	b, err := NewConstant(expectedDuration)
	if err != nil {
		t.Fatalf("failed to create constant backoff: %v", err)
	}

	resettableB := WithCappedDuration(cappedDuration, b)

	val, _ := resettableB.Next()
	if val != cappedDuration {
		t.Fatalf("expected %v to be %v due to capped duration before reset", val, cappedDuration)
	}

	resettableB.Reset()
	val, _ = resettableB.Next()
	if val != cappedDuration {
		t.Fatalf("expected %v to be %v due to capped duration after reset", val, cappedDuration)
	}
}

func TestConstantBackoff_ExplicitReset(t *testing.T) {
	expectedDuration := 3 * time.Second
	cappedDuration := 2 * time.Second

	b, err := NewConstant(expectedDuration)
	if err != nil {
		t.Fatalf("failed to create constant backoff: %v", err)
	}

	resettableB := WithCappedDuration(cappedDuration, b)

	val, _ := resettableB.Next()
	if val != cappedDuration {
		t.Fatalf("expected %v to be %v due to capped duration before reset", val, cappedDuration)
	}

	// now we're going to explicitly pass in a reset function that DOES NOT observe the cap,
	// and we expect the reset to no longer have the cap

	explicitylyResettableB := WithReset(func() Backoff {
		b, err := NewConstant(expectedDuration)
		if err != nil {
			t.Fatalf("failed to create constant backoff: %v", err)
		}

		return b
	}, resettableB)

	// before reset it should observe the cap
	val, _ = explicitylyResettableB.Next()
	if val != cappedDuration {
		t.Fatalf("expected %v to be %v due to capped duration after reset", val, cappedDuration)
	}

	// after reset the cap should go away
	explicitylyResettableB.Reset()
	val, _ = explicitylyResettableB.Next()
	if val != expectedDuration {
		t.Fatalf("expected %v to be %v due to reset without adding back capped duration", val, expectedDuration)
	}
}
