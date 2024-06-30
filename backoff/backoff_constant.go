package backoff

import (
	"fmt"
	"time"

	"github.com/swayne275/go-retry/common/backoff"
)

// NewConstant creates a new constant backoff using the value t. The wait time
// is the provided constant value. It returns an error if t is not greater than 0.
func NewConstant(t time.Duration) (backoff.Backoff, error) {
	if t <= 0 {
		return nil, fmt.Errorf("constant backoff must be greater than zero")
	}

	return BackoffFunc(func() (time.Duration, bool) {
		return t, false
	}), nil
}
