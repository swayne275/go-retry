package backoff

import "time"

// Backoff is an interface that backs off.
type Backoff interface {
	// Next returns the time duration to wait and whether to stop.
	Next() (next time.Duration, stop bool)
	// Reset sets the undecorated backoff back to its initial parameters
	Reset()
}
