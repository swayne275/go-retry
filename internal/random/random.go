package random

import (
	"fmt"
	"math/rand"
	"sync"
)

type LockedSource struct {
	src *rand.Rand
	mu  sync.Mutex
}

var _ rand.Source64 = (*LockedSource)(nil)

func NewLockedRandom(seed int64) *LockedSource {
	return &LockedSource{src: rand.New(rand.NewSource(seed))}
}

// Int63 mimics math/rand.(*Rand).Int63 with mutex locked.
func (r *LockedSource) Int63() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.src.Int63()
}

// Seed mimics math/rand.(*Rand).Seed with mutex locked.
func (r *LockedSource) Seed(seed int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.src.Seed(seed)
}

// Uint64 mimics math/rand.(*Rand).Uint64 with mutex locked.
func (r *LockedSource) Uint64() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.src.Uint64()
}

// Int63n mimics math/rand.(*Rand).Int63n with mutex locked.
// It will panic if n is not a positive, non-zero integer
func (r *LockedSource) Int63n(n int64) int64 {
	if n <= 0 {
		panic(fmt.Sprintf("invalid argument. n must be positive and non-zero, got %d", n))
	}

	if n&(n-1) == 0 { // n is power of two, can mask
		return r.Int63() & (n - 1)
	}

	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := r.Int63()
	for v > max {
		v = r.Int63()
	}
	return v % n
}
