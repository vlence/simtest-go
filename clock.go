package simtest

import (
	"time"
)

// A Clock reports the current time whenever Now is called.
// An application should use the same clock everywhere to
// tell the time.
type Clock interface {
        // Now returns the current time.
        Now() time.Time

        NewTimer(d time.Duration) Timer

        NewTicker(d time.Duration) Ticker
}

// A Timer represents an action that needs to be completed in the future.
// The interface is deliberately kept similar to that of *time.Timer.
type Timer interface {
        Reset(d time.Duration) bool

        Stop() bool
}

// A Ticker represents an action that needs to be executed at intervals.
// The interface is deliberately kept similar to that of *time.Ticker.
type Ticker interface {
        Reset(d time.Duration)

        Stop()
}
