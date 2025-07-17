package simtest

import (
	"sync"
	"time"
)

var simClock *SimClock

// A Clock reports the current time whenever Now is called.
// An application should use the same clock everywhere to
// tell the time.
type Clock interface {
        // Now returns the current time.
        Now() time.Time

        // NewTimer returns a timer that fires after d time has passed.
        NewTimer(d time.Duration) Timer

        // NewTicker returns a ticker that fires every d intervals.
        NewTicker(d time.Duration) Ticker
}

// SimClock represents a simulated clock. The clock moves forward
// in time only when Tick is called.
type SimClock struct{
        mu *sync.RWMutex
        now time.Time
}

// NewSimClock returns a SimClock whose current time is now.
func NewSimClock(now time.Time) *SimClock {
        if simClock != nil {
                return simClock
        }

        simClock = new(SimClock)
        simClock.mu = new(sync.RWMutex)
        simClock.now = now

        go listenForTimerEvents()

        return simClock
}

// Now returns the current time. To move
// the time forward call Tick. Repeated calls to Now will return
// the same time until Tick is called.
func (clock *SimClock) Now() time.Time {
        clock.mu.RLock()
        defer clock.mu.RUnlock()

        return clock.now
}

// NewTimer returns a *SimTimer.
func (clock *SimClock) NewTimer(d time.Duration) Timer {
        return newSimTimer(clock.Now().Add(d))
}

// NewTicker returns a *SimTicker.
func (clock *SimClock) NewTicker(d time.Duration) Ticker {
        ch := make(chan time.Time)
        ticker := newSimTicker(d, (<-chan time.Time)(ch), clock)

        return ticker
}

// Tick moves the clock's time forward by tickSize and returns
// the current time.
func (clock *SimClock) Tick(tickSize time.Duration) time.Time {
        clock.mu.Lock()
        defer clock.mu.Unlock()

        now := clock.now.Add(tickSize)
        clock.now = now

        timerTickChan <- now

        return now
}

// Stop stops this clock and its associated timers and tickers.
// Any pending timers and tickers are never fired after Stop is
// called.
func (clock *SimClock) Stop() {
        stopListeningForTimerEvents()
        simClock = nil
}
