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
}

// realClock represents the real world clock. Now will always return
// the current real world time according to the underlying system.
type realClock struct{}

// Now returns the current time.
func (clock *realClock) Now() time.Time {
        return time.Now()
}

// RealClock represents real world time. Calling Now on RealClock
// will always return your system's current time. Use this to tell
// the time when running your app in production.
var RealClock Clock = &realClock{}

// SimClock represents a simulated clock. The clock moves forward
// in time only when Tick is called.
type SimClock struct{
        now time.Time
}

// Now returns the current time. To move
// the time forward call Tick. Repeated calls to Now will return
// the same time until Tick is called.
func (clock *SimClock) Now() time.Time {
        return clock.now
}

// Tick moves the clock's time forward by tickSize.
func (clock *SimClock) Tick(tickSize time.Duration) {
        clock.now = clock.now.Add(tickSize)
}
