package random

import (
	"testing"
)

// TestInt63nValidInput tests that Int63n returns a value within the expected range for valid inputs.
func TestInt63nValidInput(t *testing.T) {
	n := int64(10)
	r := NewLockedRandom(0)

	for i := 0; i < 100; i++ {
		val := r.Int63n(n)
		if val < 0 || val >= n {
			t.Errorf("expected value in range [0, %d), got %d", n, val)
		}
	}
}

// TestInt63nInvalidInput tests that Int63n returns 0 and an error for invalid inputs.
func TestInt63nInvalidInput(t *testing.T) {
	assertPanic := func(t *testing.T, f func(int64) int64, input int64) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()

		_ = f(input)
	}

	invalidInputs := []int64{0, -1, -10}
	r := NewLockedRandom(0)

	for _, n := range invalidInputs {
		assertPanic(t, r.Int63n, n)
	}
}

// TestInt63nConcurrency tests that Int63n behaves correctly when called concurrently.
func TestInt63nConcurrency(t *testing.T) {
	n := int64(10)
	numGoroutines := 100
	results := make(chan int64, numGoroutines)

	r := NewLockedRandom(0)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			val := r.Int63n(n)
			results <- val
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		val := <-results
		if val < 0 || val >= n {
			t.Errorf("expected value in range [0, %d), got %d", n, val)
		}
	}
}
