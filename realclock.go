package simtest

import "time"

// realClock represents the real world clock. Now will always return
// the current real world time according to the underlying system.
type realClock struct{}

// Now returns the result of calling time.Now.
func (clock *realClock) Now() time.Time {
        return time.Now()
}

// NewTimer returns the result of calling time.NewTimer.
func (*realClock) NewTimer(d time.Duration) Timer {
        return time.NewTimer(d)
}

// NewTicker returns the result of calling time.NewTicker.
func (*realClock) NewTicker(d time.Duration) Ticker {
        return time.NewTicker(d)
}

// RealClock represents real world time. Calling Now on RealClock
// will always return your system's current time. Use this to tell
// the time when running your app in production.
var RealClock Clock = &realClock{}
